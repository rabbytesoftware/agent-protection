package eslint

// FileResult is eslint's JSON report for one linted file: its path and
// every message raised in it.
type FileResult struct {
	FilePath string    `json:"filePath"`
	Messages []Message `json:"messages"`
}
