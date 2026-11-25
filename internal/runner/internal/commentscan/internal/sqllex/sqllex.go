package sqllex

import (
	"path/filepath"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
)

const sqlExtension = ".sql"

func Sqllex() (language lexeme.Language) {
	return lexeme.Language{
		Matches: isSQL,
		Spans:   sqlCode,
	}
}

func sqlCode(
	text []byte,
) (spans []lexeme.Span) {
	return lexeme.Without(
		text,
		sqlSpans(text),
		isMigrationMarker,
	)
}

func isSQL(
	name string,
) (matched bool) {
	return strings.EqualFold(
		filepath.Ext(name),
		sqlExtension,
	)
}
