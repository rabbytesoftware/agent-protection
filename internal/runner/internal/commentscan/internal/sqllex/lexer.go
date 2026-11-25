package sqllex

import (
	"bytes"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
)

const (
	lineCommentOpener = "--"
	quotes            = "'\"`"
	dollar            = '$'
	tagBytes          = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	tagStarters       = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_$"
)

func sqlSpans(
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
		lineCommentOpener,
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
	if strings.IndexByte(
		quotes,
		text[at],
	) >= 0 {
		return lexeme.RawEnd(
			text,
			at,
		)
	}

	return dollarQuotedEnd(
		text,
		at,
	)
}

func dollarQuotedEnd(
	text []byte,
	at int,
) (end int) {
	if text[at] != dollar {
		return at + 1
	}

	tag := dollarTag(
		text,
		at,
	)
	if tag == "" {
		return at + 1
	}

	offset := bytes.Index(
		text[at+len(tag):],
		[]byte(tag),
	)
	if offset < 0 {
		return len(text)
	}

	return at + len(tag) + offset + len(tag)
}

func dollarTag(
	text []byte,
	at int,
) (tag string) {
	isPrecededByWord := at > 0 && strings.IndexByte(
		tagBytes,
		text[at-1],
	) >= 0
	isTagStart := at+1 < len(text) && strings.IndexByte(
		tagStarters,
		text[at+1],
	) >= 0
	if isPrecededByWord ||
		!isTagStart {
		return ""
	}

	closing := bytes.IndexByte(
		text[at+1:],
		dollar,
	)
	if closing < 0 {
		return ""
	}

	hasForeignByte := len(bytes.TrimLeft(
		text[at+1:at+1+closing],
		tagBytes,
	)) > 0
	if hasForeignByte {
		return ""
	}

	return string(text[at : at+closing+2])
}
