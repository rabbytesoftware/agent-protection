package scan

import "github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"

type source struct {
	path     string
	language lexeme.Language
}
