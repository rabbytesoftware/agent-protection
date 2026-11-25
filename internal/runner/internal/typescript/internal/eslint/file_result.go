package eslint

type FileResult struct {
	FilePath string    `json:"filePath"`
	Messages []Message `json:"messages"`
}
