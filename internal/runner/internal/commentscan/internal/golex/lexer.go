package golex

import (
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
)

const (
	rawQuote        = '`'
	interpretQuotes = "'\""
)

func slashSpans(
	text []byte,
) (spans []lexeme.Span) {
	cursor := 0
	for cursor < len(text) {
		next, found := tokenEnd(
			text,
			cursor,
		)
		spans = append(
			spans,
			found...,
		)
		cursor = next
	}

	return spans
}

func tokenEnd(
	text []byte,
	at int,
) (next int, spans []lexeme.Span) {
	switch {
	case lexeme.HasPrefix(
		text,
		at,
		lexeme.LineCommentOpener,
	):
		return lexeme.LineComment(
			text,
			at,
		)
	case lexeme.HasPrefix(
		text,
		at,
		lexeme.BlockCommentOpener,
	):
		next = lexeme.BlockEnd(
			text,
			at,
		)
		return next,
			[]lexeme.Span{{Start: at, End: next}}
	}

	return literalEnd(
			text,
			at,
		),
		nil
}

func literalEnd(
	text []byte,
	at int,
) (end int) {
	if text[at] == rawQuote {
		return lexeme.RawEnd(
			text,
			at,
		)
	}

	if strings.IndexByte(
		interpretQuotes,
		text[at],
	) >= 0 {
		return lexeme.QuotedEnd(
			text,
			at,
			lexeme.LineStops,
		)
	}

	return at + 1
}
