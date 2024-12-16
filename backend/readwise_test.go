package backend

import (
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildPayload_NoBookmarks(t *testing.T) {
	var expected Response
	var contentIndex map[string]Content
	var bookmarks []Bookmark
	var actual, _ = BuildPayload(bookmarks, contentIndex, slog.New(&discardHandler{}))
	assert.Equal(t, expected, actual[0])
}

func TestBuildPayload_BookmarksPresent(t *testing.T) {
	highlights := []Highlight{{
		Text:          "Hello World",
		Title:         "A Book",
		Author:        "Computer",
		SourceURL:     "mnt://kobo/blah/Good Book - An Author.epub",
		SourceType:    SourceType,
		Category:      SourceCategory,
		Note:          "Making a note here",
		HighlightedAt: "2006-01-02T15:04:05+00:00",
	}}
	expected := Response{Highlights: highlights}
	contentIndex := map[string]Content{"mnt://kobo/blah/Good Book - An Author.epub": {ContentID: "mnt://kobo/blah/Good Book - An Author.epub", Title: "A Book", Attribution: "Computer"}}
	bookmarks := []Bookmark{
		{VolumeID: "mnt://kobo/blah/Good Book - An Author.epub", Text: "Hello World", DateCreated: "2006-01-02T15:04:05.000", Annotation: "Making a note here"},
	}
	var actual, _ = BuildPayload(bookmarks, contentIndex, slog.New(&discardHandler{}))
	assert.Equal(t, expected, actual[0])
}

func TestBuildPayload_HandleAnnotationOnly(t *testing.T) {
	highlights := []Highlight{{
		Text:          "Placeholder for attached annotation",
		Title:         "A Book",
		Author:        "Computer",
		SourceURL:     "mnt://kobo/blah/Good Book - An Author.epub",
		SourceType:    SourceType,
		Category:      SourceCategory,
		Note:          "Making a note here",
		HighlightedAt: "2006-01-02T15:04:05+00:00",
	}}
	expected := Response{Highlights: highlights}
	contentIndex := map[string]Content{"mnt://kobo/blah/Good Book - An Author.epub": {ContentID: "mnt://kobo/blah/Good Book - An Author.epub", Title: "A Book", Attribution: "Computer"}}
	bookmarks := []Bookmark{
		{VolumeID: "mnt://kobo/blah/Good Book - An Author.epub", DateCreated: "2006-01-02T15:04:05.000", Annotation: "Making a note here"},
	}
	var actual, _ = BuildPayload(bookmarks, contentIndex, slog.New(&discardHandler{}))
	assert.Equal(t, expected, actual[0])
}

func TestBuildPayload_TitleFallback(t *testing.T) {
	highlights := []Highlight{{
		Text:          "Hello World",
		Title:         "Good Book - An Author",
		Author:        "",
		SourceURL:     "mnt://kobo/blah/Good Book - An Author.epub",
		SourceType:    SourceType,
		Category:      SourceCategory,
		Note:          "Making a note here",
		HighlightedAt: "2006-01-02T15:04:05+00:00",
	}}
	expected := Response{Highlights: highlights}
	contentIndex := map[string]Content{"mnt://kobo/blah/Good Book - An Author.epub": {ContentID: "mnt://kobo/blah/Good Book - An Author.epub"}}
	bookmarks := []Bookmark{
		{VolumeID: "mnt://kobo/blah/Good Book - An Author.epub", Text: "Hello World", DateCreated: "2006-01-02T15:04:05.000", Annotation: "Making a note here"},
	}
	var actual, _ = BuildPayload(bookmarks, contentIndex, slog.New(&discardHandler{}))
	assert.Equal(t, expected, actual[0])
}

func TestBuildPayload_TitleFallbackFailure(t *testing.T) {
	highlights := []Highlight{{
		Text:          "Hello World",
		SourceURL:     "\t",
		SourceType:    SourceType,
		Category:      SourceCategory,
		Note:          "Making a note here",
		HighlightedAt: "2006-01-02T15:04:05+00:00",
	}}
	expected := Response{Highlights: highlights}
	contentIndex := map[string]Content{"\t": {ContentID: "\t"}}
	bookmarks := []Bookmark{
		{VolumeID: "\t", Text: "Hello World", DateCreated: "2006-01-02T15:04:05.000", Annotation: "Making a note here"},
	}
	var actual, _ = BuildPayload(bookmarks, contentIndex, slog.New(&discardHandler{}))
	assert.Equal(t, expected, actual[0])
}

func TestBuildPayload_NoHighlightDateCreated(t *testing.T) {
	highlights := []Highlight{{
		Text:          "Hello World",
		SourceURL:     "\t",
		SourceType:    SourceType,
		Category:      SourceCategory,
		Note:          "Making a note here",
		HighlightedAt: "2006-01-02T15:04:05+00:00",
	}}
	expected := Response{Highlights: highlights}
	contentIndex := map[string]Content{"\t": {ContentID: "\t"}}
	bookmarks := []Bookmark{
		{VolumeID: "\t", Text: "Hello World", DateCreated: "", Annotation: "Making a note here", DateModified: "2006-01-02T15:04:05Z"},
	}
	var actual, _ = BuildPayload(bookmarks, contentIndex, slog.New(&discardHandler{}))
	assert.Equal(t, expected, actual[0])
}

func TestBuildPayload_NoHighlightDateAtAll(t *testing.T) {
	contentIndex := map[string]Content{"\t": {ContentID: "\t"}}
	bookmarks := []Bookmark{
		{VolumeID: "abc123", Text: "Hello World", Annotation: "Making a note here"},
	}
	var actual, _ = BuildPayload(bookmarks, contentIndex, slog.New(&discardHandler{}))
	assert.Equal(t, actual[0].Highlights[0].SourceURL, bookmarks[0].VolumeID)
	assert.Equal(t, actual[0].Highlights[0].Text, bookmarks[0].Text)
	assert.Equal(t, actual[0].Highlights[0].Note, bookmarks[0].Annotation)
	assert.NotEmpty(t, actual[0].Highlights[0].HighlightedAt)
}

func TestBuildPayload_SkipMalformedBookmarks(t *testing.T) {
	var expected Response
	contentIndex := map[string]Content{"mnt://kobo/blah/Good Book - An Author.epub": {ContentID: "mnt://kobo/blah/Good Book - An Author.epub"}}
	bookmarks := []Bookmark{
		{VolumeID: "mnt://kobo/blah/Good Book - An Author.epub", DateCreated: "2006-01-02T15:04:05.000"},
	}
	var actual, _ = BuildPayload(bookmarks, contentIndex, slog.New(&discardHandler{}))
	assert.Equal(t, expected, actual[0])
}

func TestBuildPayload_LongHighlightChunks(t *testing.T) {
	var b strings.Builder
	for i := 0; i < 20000; i++ {
		fmt.Fprintf(&b, "%s", "a")
	}
	text := b.String()
	highlights := []Highlight{
		{
			Text:          text[0:MaxHighlightLen],
			Title:         "A Book",
			Author:        "Computer",
			SourceURL:     "mnt://kobo/blah/Good Book - An Author.epub",
			SourceType:    SourceType,
			Category:      SourceCategory,
			Note:          "Making a note here",
			HighlightedAt: "2006-01-02T15:04:05+00:00",
		},
		{
			Text:          text[MaxHighlightLen : MaxHighlightLen*2],
			Title:         "A Book",
			Author:        "Computer",
			SourceURL:     "mnt://kobo/blah/Good Book - An Author.epub",
			SourceType:    SourceType,
			Category:      SourceCategory,
			Note:          "Making a note here",
			HighlightedAt: "2006-01-02T15:04:05+00:00",
		},
		{
			Text:          text[MaxHighlightLen*2 : 20000],
			Title:         "A Book",
			Author:        "Computer",
			SourceURL:     "mnt://kobo/blah/Good Book - An Author.epub",
			SourceType:    SourceType,
			Category:      SourceCategory,
			Note:          "Making a note here",
			HighlightedAt: "2006-01-02T15:04:05+00:00",
		}}
	expected := Response{Highlights: highlights}
	contentIndex := map[string]Content{"mnt://kobo/blah/Good Book - An Author.epub": {ContentID: "mnt://kobo/blah/Good Book - An Author.epub", Title: "A Book", Attribution: "Computer"}}
	bookmarks := []Bookmark{
		{VolumeID: "mnt://kobo/blah/Good Book - An Author.epub", Text: text, DateCreated: "2006-01-02T15:04:05.000", Annotation: "Making a note here"},
	}
	actual, err := BuildPayload(bookmarks, contentIndex, slog.New(&discardHandler{}))
	assert.NoError(t, err)
	assert.Equal(t, expected, actual[0])
}

func TestExcessiveHighlightAmounts(t *testing.T) {
	expected := 3
	contentIndex := map[string]Content{"mnt://kobo/blah/Good Book - An Author.epub": {ContentID: "mnt://kobo/blah/Good Book - An Author.epub", Title: "A Book", Attribution: "Computer"}}
	bookmarks := []Bookmark{}
	for i := 0; i < 4002; i++ {
		bookmarks = append(bookmarks, Bookmark{VolumeID: "mnt://kobo/blah/Good Book - An Author.epub", Text: fmt.Sprintf("hi%d", i), DateCreated: "2006-01-02T15:04:05.000", Annotation: "Making a note here"})
	}
	actual, err := BuildPayload(bookmarks, contentIndex, slog.New(&discardHandler{}))
	assert.NoError(t, err)
	assert.Equal(t, expected, len(actual))
}
