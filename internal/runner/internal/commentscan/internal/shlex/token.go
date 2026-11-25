package shlex

import "github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"

type token struct {
	next       int
	spans      []lexeme.Span
	delimiters []string
}
