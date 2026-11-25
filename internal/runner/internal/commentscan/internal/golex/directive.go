package golex

import (
	"bytes"
	"regexp"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
)

const (
	indirectMarker   = "// indirect"
	directivePattern = `^//(?:go:[a-z_]+(?:\s.*)?|line \S+|export \w+|extern \w+| \+build [\w!.,\s]+)$`
)

func isIndirectMarker(
	text []byte,
	span lexeme.Span,
) (directive bool) {
	return !lexeme.IsLineLeading(
		text,
		span,
	) && string(bytes.TrimSpace(text[span.Start:span.End])) == indirectMarker
}

func isCompilerDirective(
	text []byte,
	span lexeme.Span,
) (directive bool) {
	return lexeme.IsLineLeading(
		text,
		span,
	) && regexp.MustCompile(directivePattern).Match(bytes.TrimRight(
		text[span.Start:span.End],
		"\r",
	))
}
