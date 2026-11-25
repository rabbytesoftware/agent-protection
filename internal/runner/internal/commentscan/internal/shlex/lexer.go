package shlex

import (
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
)

const (
	hashMark     = '#'
	singleQuote  = '\''
	doubleQuote  = '"'
	hashOpeners  = " \t\n;(|&"
	quoteOpeners = " \t\n=([{,:$!|&;<>"
)

func shellSpans(
	text []byte,
) (spans []lexeme.Span) {
	cursor := lexeme.ShebangEnd(text)
	for cursor < len(text) {
		next, found := statementEnd(
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

func statementEnd(
	text []byte,
	at int,
) (next int, spans []lexeme.Span) {
	cursor := at
	delimiters := []string{}
	for lexeme.IsOpen(
		text,
		cursor,
		lexeme.LineStops,
	) {
		current := tokenAt(
			text,
			cursor,
		)
		spans = append(
			spans,
			current.spans...,
		)
		delimiters = append(
			delimiters,
			current.delimiters...,
		)
		cursor = current.next
	}

	return heredocBodiesEnd(
			text,
			cursor+1,
			delimiters,
		),
		spans
}

func tokenAt(
	text []byte,
	at int,
) (current token) {
	switch {
	case text[at] == lexeme.Escape:
		return token{next: lexeme.EscapedNext(
			text,
			at,
		)}
	case text[at] == hashMark && opensAt(
		text,
		at,
		hashOpeners,
	):
		end, found := lexeme.LineComment(
			text,
			at,
		)
		return token{next: end, spans: found}
	case text[at] == doubleQuote:
		return token{next: lexeme.QuotedEnd(
			text,
			at,
			lexeme.NoStops,
		)}
	}

	return unquotedToken(
		text,
		at,
	)
}

func unquotedToken(
	text []byte,
	at int,
) (current token) {
	isSingleQuoted := text[at] == singleQuote && opensAt(
		text,
		at,
		quoteOpeners,
	)
	if isSingleQuoted {
		return token{next: lexeme.RawEnd(
			text,
			at,
		)}
	}

	if lexeme.HasPrefix(
		text,
		at,
		heredocOperator,
	) {
		return heredocAt(
			text,
			at,
		)
	}

	return token{next: at + 1}
}

func opensAt(
	text []byte,
	at int,
	openers string,
) (opens bool) {
	return at == 0 || strings.IndexByte(
		openers,
		text[at-1],
	) >= 0
}
