package golex

import (
	"path/filepath"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
)

const goExtension = ".go"

func Gomod() (language lexeme.Language) {
	return lexeme.Language{
		Matches: isModuleFile,
		Spans:   moduleCode,
	}
}

func Gosource() (language lexeme.Language) {
	return lexeme.Language{
		Matches: isSource,
		Spans:   sourceCode,
	}
}

func moduleCode(
	text []byte,
) (spans []lexeme.Span) {
	return lexeme.Without(
		text,
		slashSpans(text),
		isIndirectMarker,
	)
}

func sourceCode(
	text []byte,
) (spans []lexeme.Span) {
	return lexeme.Without(
		text,
		slashSpans(text),
		isCompilerDirective,
	)
}

func isModuleFile(
	name string,
) (matched bool) {
	return slices.Contains(
		[]string{"go.mod", "go.work"},
		name,
	)
}

func isSource(
	name string,
) (matched bool) {
	return filepath.Ext(name) == goExtension
}
