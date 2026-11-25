package pylex

import (
	"bytes"
	"regexp"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
)

const (
	encodingPattern = `^#\s*(?:-\*-\s*)?(?:en)?coding[:=]\s*[-\w.]+\s*(?:-\*-)?\s*$|^#\s*vim:\s*set\s+fileencoding=[-\w.]+\s*:\s*$`
	magicPattern    = `^#\s*(?:-\*-\s*)?(?:frozen_string_literal|encoding|coding|warn_indent|shareable_constant_value)\s*:\s*[\w-]+\s*(?:-\*-)?\s*$`
	schemaPattern   = `^#:schema\s+\S+\s*$`
	encodingLines   = 2
	schemaLines     = 1
)

func isEncodingDeclaration(
	text []byte,
	span lexeme.Span,
) (directive bool) {
	return isLeadingMatch(
		text,
		span,
		encodingPattern,
	) && lexeme.LineOf(
		text,
		span.Start,
	) <= encodingLines
}

func isMagicComment(
	text []byte,
	span lexeme.Span,
) (directive bool) {
	return isLeadingMatch(
		text,
		span,
		magicPattern,
	) && isInHeader(
		text,
		span.Start,
	)
}

func isSchemaDirective(
	text []byte,
	span lexeme.Span,
) (directive bool) {
	return isLeadingMatch(
		text,
		span,
		schemaPattern,
	) && lexeme.LineOf(
		text,
		span.Start,
	) <= schemaLines
}

func isLeadingMatch(
	text []byte,
	span lexeme.Span,
	pattern string,
) (matches bool) {
	return lexeme.IsLineLeading(
		text,
		span,
	) && regexp.MustCompile(pattern).Match(text[span.Start:span.End])
}

func isInHeader(
	text []byte,
	at int,
) (header bool) {
	return !slices.ContainsFunc(
		bytes.Split(
			text[:at],
			[]byte{lexeme.Newline},
		),
		isCodeLine,
	)
}

func isCodeLine(
	line []byte,
) (code bool) {
	trimmed := bytes.TrimSpace(line)
	return len(trimmed) > 0 &&
		trimmed[0] != hashMark
}
