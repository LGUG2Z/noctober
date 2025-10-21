package backend

var (
	configFilename  = "october/config.json"
	MaxHighlightLen = 8096 // It's actually 8191 but we'll go under the limit anyway
	UserAgentFmt    = "noctober/%s <https://github.com/LGUG2Z/noctober>"
)
