package backend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

const (
	HIGHLIGHT_REQUEST_BATCH_MAX = 2000
)

type Response struct {
	Highlights []Highlight `json:"highlights"`
}

type Highlight struct {
	Content string   `json:"content"`
	URL     string   `json:"url"`
	Title   string   `json:"title"`
	Created string   `json:"created,omitempty"`
	Tags    []string `json:"tags,omitempty"`
	Author  string   `json:"author,omitempty"`
}

type Notado struct {
	logger    *slog.Logger
	UserAgent string
}

func (n *Notado) SendBookmarks(payloads []Response, token string) (int, error) {
	// Flatten all highlights from all payloads into a single slice
	var allHighlights []Highlight
	for _, payload := range payloads {
		allHighlights = append(allHighlights, payload.Highlights...)
	}

	if len(allHighlights) == 0 {
		n.logger.Info("No highlights to send to Notado")
		return 0, nil
	}

	// Create GraphQL mutation payload
	graphqlPayload := struct {
		Query     string      `json:"query"`
		Variables interface{} `json:"variables"`
	}{
		Query: `mutation ImportNotes($notes: [NewImportNote!]!) {
			importNotes(notes: $notes)
		}`,
		Variables: struct {
			Notes []Highlight `json:"notes"`
		}{
			Notes: allHighlights,
		},
	}

	data, err := json.Marshal(graphqlPayload)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal GraphQL payload: %+v", err)
	}

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, "https://notado.app/graphql", bytes.NewBuffer(data))
	if err != nil {
		return 0, fmt.Errorf("failed to construct request: %+v", err)
	}

	req.Header.Add("X-API-TOKEN", token)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", n.UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to send request to Notado: %+v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err == nil {
			n.logger.Error("Received a non-200 response from Notado",
				slog.Int("status_code", resp.StatusCode),
				slog.String("response", string(body)),
			)
		}
		return 0, fmt.Errorf("received a non-200 status code from Notado: code %d", resp.StatusCode)
	}

	// Optionally, you could parse the GraphQL response here to check for errors
	// For now, we'll assume success if we get a 200 status code

	n.logger.Info("Successfully sent bookmarks to Notado",
		slog.Int("highlight_count", len(allHighlights)),
	)

	return len(allHighlights), nil
}

func BuildPayload(bookmarks []Bookmark, contentIndex map[string]Content, logger *slog.Logger) ([]Response, error) {
	var payloads []Response
	var currentBatch Response
	for count, entry := range bookmarks {
		// If max payload size is reached, start building another batch which will be sent separately
		if count > 0 && (count%HIGHLIGHT_REQUEST_BATCH_MAX == 0) {
			fmt.Println(count / HIGHLIGHT_REQUEST_BATCH_MAX)
			payloads = append(payloads, currentBatch)
			currentBatch = Response{}
		}
		source := contentIndex[entry.VolumeID]
		logger.Debug("Parsing highlight",
			slog.String("title", source.Title),
		)
		var createdAt string
		if entry.DateCreated == "" {
			logger.Warn("No date created for bookmark. Defaulting to date last modified.",
				slog.String("title", source.Title),
				slog.String("volume_id", entry.VolumeID),
			)
			if entry.DateModified == "" {
				logger.Warn("No date modified for bookmark. Default to current date.",
					slog.String("title", source.Title),
					slog.String("volume_id", entry.VolumeID),
				)
				createdAt = time.Now().Format("2006-01-02T15:04:05-07:00")
			} else {
				t, err := time.Parse("2006-01-02T15:04:05Z", entry.DateModified)
				if err != nil {
					logger.Error("Failed to parse a valid timestamp from date modified field",
						slog.String("error", err.Error()),
						slog.String("title", source.Title),
						slog.String("volume_id", entry.VolumeID),
						slog.String("date_modified", entry.DateModified),
					)
					return []Response{}, err
				}
				createdAt = t.Format("2006-01-02T15:04:05-07:00")
			}
		} else {
			t, err := time.Parse("2006-01-02T15:04:05.000", entry.DateCreated)
			if err != nil {
				logger.Error("Failed to parse a valid timestamp from date created field",
					slog.String("error", err.Error()),
					slog.String("title", source.Title),
					slog.String("volume_id", entry.VolumeID),
					slog.String("date_modified", entry.DateModified),
				)
				return []Response{}, err
			}
			createdAt = t.Format("2006-01-02T15:04:05-07:00")
		}
		text := NormaliseText(entry.Text)
		if entry.Annotation != "" && text == "" {
			// I feel like this state probably shouldn't be possible but we'll handle it anyway
			// since it's useful to surface annotations, regardless of highlights. We put a
			// glaring placeholder here because the text field is required by the Notado API.
			text = "Placeholder for attached annotation"
		}
		if entry.Annotation == "" && text == "" {
			// This state should be impossible but stranger things have happened so worth a sanity check
			logger.Warn("Found an entry with neither highlighted text nor an annotation so skipping entry",
				slog.String("title", source.Title),
				slog.String("volume_id", entry.VolumeID),
			)
			continue
		}
		if source.Title == "" {
			// While Kepubs have a title in the Kobo database, the same can't be guaranteed for epubs at all.
			// In that event, we just fall back to using the filename
			sourceFile, err := url.Parse(entry.VolumeID)
			if err != nil {
				// While extremely unlikely, we should handle the case where a VolumeID doesn't have a suffix. This condition is only
				// triggered for completely busted names such as control codes given url.Parse will happen take URLs without a protocol
				// or even just arbitrary strings. Given we don't set a title here, we will use the Notado fallback which is to add
				// these highlights to a book called "Quotes" and let the user figure out their metadata situation. That reminds me though:
				// TODO: Test exports with non-epub files
				logger.Warn("Failed to retrieve epub title. This is not a hard requirement so sending with a dummy title.",
					slog.String("error", err.Error()),
					slog.String("title", source.Title),
					slog.String("volume_id", entry.VolumeID),
				)
				goto sendhighlight
			}
			filename := path.Base(sourceFile.Path)
			logger.Debug("No source title. Constructing title from filename",
				slog.String("filename", filename),
			)
			source.Title = strings.TrimSuffix(filename, ".epub")
		}
	sendhighlight:
		highlightChunks := splitHighlight(text, MaxHighlightLen)
		for _, chunk := range highlightChunks {
			tags := []string{}

			for _, possibleTag := range strings.Split(entry.Annotation, " ") {
				if strings.HasPrefix(possibleTag, ".") {
					tags = append(tags, strings.TrimPrefix(possibleTag, "."))
				}
			}

			highlight := Highlight{
				Content: chunk,
				URL:     entry.VolumeID,
				Title:   fmt.Sprintf("%s - %s", source.Title, source.Attribution),
				Created: createdAt,
				Tags:    tags,
				Author:  source.Attribution,
			}
			currentBatch.Highlights = append(currentBatch.Highlights, highlight)
		}
		logger.Debug("Successfully compiled highlights for book",
			slog.String("title", source.Title),
			slog.String("volume_id", entry.VolumeID),
			slog.Int("chunks", len(highlightChunks)),
		)
	}
	payloads = append(payloads, currentBatch)
	logger.Info("Succcessfully parsed highlights",
		slog.Int("highlight_count", len(currentBatch.Highlights)),
		slog.Int("batch_count", len(payloads)),
	)
	return payloads, nil
}

func NormaliseText(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}
