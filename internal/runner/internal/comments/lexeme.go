package comments

import (
	"bytes"
	"strings"
	"unicode/utf8"
)

const (
	lineCommentOpener  = "//"
	blockCommentOpener = "/*"
	blockCommentCloser = "*/"
	shebang            = "#!"
	escape             = '\\'
	newline            = '\n'
	escapeWidth        = 2
	identifierBytes    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_$"
)

func hasPrefix(
	text []byte,
	at int,
	prefix string,
) (matches bool) {
	return bytes.HasPrefix(
		text[at:],
		[]byte(prefix),
	)
}

func lineEnd(
	text []byte,
	at int,
) (end int) {
	offset := bytes.IndexByte(
		text[at:],
		newline,
	)
	if offset < 0 {
		return len(text)
	}

	return at + offset
}

func shebangEnd(
	text []byte,
) (end int) {
	if !bytes.HasPrefix(
		text,
		[]byte(shebang),
	) {
		return 0
	}

	return lineEnd(
		text,
		0,
	)
}

func escapedNext(
	text []byte,
	at int,
) (next int) {
	if text[at] != escape {
		return at + 1
	}

	return min(
		at+escapeWidth,
		len(text),
	)
}

func quotedEnd(
	text []byte,
	at int,
	stops string,
) (end int) {
	quote := text[at]
	return closedEnd(
		text,
		scanUntil(
			text,
			at+1,
			string(quote)+stops,
		),
		quote,
	)
}

func scanUntil(
	text []byte,
	at int,
	stops string,
) (cursor int) {
	cursor = at
	for isOpen(
		text,
		cursor,
		stops,
	) {
		cursor = escapedNext(
			text,
			cursor,
		)
	}

	return cursor
}

func isOpen(
	text []byte,
	cursor int,
	stops string,
) (open bool) {
	return cursor < len(text) && strings.IndexByte(
		stops,
		text[cursor],
	) < 0
}

func closedEnd(
	text []byte,
	cursor int,
	closer byte,
) (end int) {
	isClosed := cursor < len(text) &&
		text[cursor] == closer
	if isClosed {
		return cursor + 1
	}

	return min(
		cursor,
		len(text),
	)
}

func lineComment(
	text []byte,
	at int,
) (end int, spans []span) {
	end = lineEnd(
		text,
		at,
	)
	return end,
		[]span{{start: at, end: end, isLine: true}}
}

func blockEnd(
	text []byte,
	at int,
) (end int) {
	bodyStart := at + len(blockCommentOpener)
	offset := bytes.Index(
		text[bodyStart:],
		[]byte(blockCommentCloser),
	)
	if offset < 0 {
		return len(text)
	}

	return bodyStart + offset + len(blockCommentCloser)
}

func isIdentifierByte(
	character byte,
) (identifier bool) {
	return character >= utf8.RuneSelf || strings.IndexByte(
		identifierBytes,
		character,
	) >= 0
}
