package clex

import (
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
)

const schemeColon = ':'

type dialect struct {
	lineOpeners []string
	quotes      string
}

func (d dialect) spans(
	text []byte,
) (spans []lexeme.Span) {
	cursor := 0
	for cursor < len(text) {
		next, found := d.tokenEnd(
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

func (d dialect) tokenEnd(
	text []byte,
	at int,
) (next int, spans []lexeme.Span) {
	switch {
	case d.opensLine(
		text,
		at,
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
	case strings.IndexByte(
		d.quotes,
		text[at],
	) >= 0:
		return lexeme.QuotedEnd(
				text,
				at,
				lexeme.LineStops,
			),
			nil
	}

	return at + 1,
		nil
}

func (d dialect) opensLine(
	text []byte,
	at int,
) (opens bool) {
	isSchemeSeparator := at > 0 && text[at-1] == schemeColon
	return !isSchemeSeparator && slices.ContainsFunc(
		d.lineOpeners,
		opensAt(
			text,
			at,
		),
	)
}

func opensAt(
	text []byte,
	at int,
) (matches func(string) bool) {
	return func(
		opener string,
	) (opens bool) {
		return lexeme.HasPrefix(
			text,
			at,
			opener,
		)
	}
}
