package pylex

import (
	"bytes"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
)

const (
	hashMark    = '#'
	quotes      = "'\""
	tripleWidth = 3
	blockBegin  = "=begin"
	blockFinish = "\n=end"
)

func hashSpans(
	text []byte,
) (spans []lexeme.Span) {
	cursor := lexeme.ShebangEnd(text)
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
	case text[at] == hashMark:
		return lexeme.LineComment(
			text,
			at,
		)
	case isBlockBegin(
		text,
		at,
	):
		next = blockCommentEnd(
			text,
			at,
		)
		return next,
			[]lexeme.Span{{Start: at, End: next}}
	case strings.IndexByte(
		quotes,
		text[at],
	) >= 0:
		return stringEnd(
				text,
				at,
			),
			nil
	}

	return lexeme.EscapedNext(
			text,
			at,
		),
		nil
}

func stringEnd(
	text []byte,
	at int,
) (end int) {
	triple := strings.Repeat(
		string(text[at]),
		tripleWidth,
	)
	if !lexeme.HasPrefix(
		text,
		at,
		triple,
	) {
		return lexeme.QuotedEnd(
			text,
			at,
			lexeme.LineStops,
		)
	}

	cursor := at + tripleWidth
	for cursor < len(text) && !lexeme.HasPrefix(
		text,
		cursor,
		triple,
	) {
		cursor = lexeme.EscapedNext(
			text,
			cursor,
		)
	}

	return min(
		cursor+tripleWidth,
		len(text),
	)
}

func isBlockBegin(
	text []byte,
	at int,
) (begins bool) {
	return lexeme.LineStart(
		text,
		at,
	) == at && lexeme.HasPrefix(
		text,
		at,
		blockBegin,
	)
}

func blockCommentEnd(
	text []byte,
	at int,
) (end int) {
	offset := bytes.Index(
		text[at:],
		[]byte(blockFinish),
	)
	if offset < 0 {
		return len(text)
	}

	return lexeme.LineEnd(
		text,
		at+offset+1,
	)
}
