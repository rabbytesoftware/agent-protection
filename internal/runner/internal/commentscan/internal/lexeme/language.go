package lexeme

type Language struct {
	Matches      func(string) bool
	Spans        func([]byte) []Span
	Interpreters []string
}
