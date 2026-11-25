package sqllex

import (
	"regexp"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
)

const migrationPattern = `^--\s*\+(?:goose\s+(?:Up|Down|StatementBegin|StatementEnd|NO TRANSACTION|ENVSUB ON|ENVSUB OFF)|migrate\s+(?:Up|Down|StatementBegin|StatementEnd)(?:\s+notransaction)?)\s*$|^--\s*name:\s*[A-Za-z_][A-Za-z0-9_]*\s+:(?:one|many|exec|execrows|execresult|execlastid|batchexec|batchmany|batchone|copyfrom)\s*$`

func isMigrationMarker(
	text []byte,
	span lexeme.Span,
) (directive bool) {
	return lexeme.IsLineLeading(
		text,
		span,
	) && regexp.MustCompile(migrationPattern).Match(text[span.Start:span.End])
}
