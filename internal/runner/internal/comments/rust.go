package comments

import (
	"bytes"
	"strings"
	"unicode/utf8"
)

const (
	innerAttribute = "#!["
	stringQuote    = '"'
	charQuote      = '\''
	rawHash        = "#"
	rawMarker      = "r"
	stringKinds    = "bc"
	noStops        = ""
	opensComment   = 1
	closesComment  = -1
)

func rustSpans(
	text []byte,
) (spans []span) {
	cursor := rustShebangEnd(text)
	for cursor < len(text) {
		next, found := rustTokenEnd(
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

func rustShebangEnd(
	text []byte,
) (end int) {
	if bytes.HasPrefix(
		text,
		[]byte(innerAttribute),
	) {
		return 0
	}

	return shebangEnd(text)
}

func rustTokenEnd(
	text []byte,
	at int,
) (next int, spans []span) {
	if hasPrefix(
		text,
		at,
		lineCommentOpener,
	) {
		return lineComment(
			text,
			at,
		)
	}

	if hasPrefix(
		text,
		at,
		blockCommentOpener,
	) {
		next = nestedBlockEnd(
			text,
			at,
		)
		return next,
			[]span{{start: at, end: next}}
	}

	return rustLiteralEnd(
			text,
			at,
		),
		nil
}

func rustLiteralEnd(
	text []byte,
	at int,
) (end int) {
	switch text[at] {
	case stringQuote:
		return quotedEnd(
			text,
			at,
			noStops,
		)
	case charQuote:
		return charOrLifetimeEnd(
			text,
			at,
		)
	}

	return rawStringEnd(
		text,
		at,
	)
}

func nestedBlockEnd(
	text []byte,
	at int,
) (end int) {
	depth, end := blockDelimiter(
		text,
		at,
	)
	for depth > 0 && end < len(text) {
		delta, next := blockDelimiter(
			text,
			end,
		)
		depth += delta
		end = next
	}

	return end
}

func blockDelimiter(
	text []byte,
	at int,
) (delta int, next int) {
	if hasPrefix(
		text,
		at,
		blockCommentOpener,
	) {
		return opensComment,
			at + len(blockCommentOpener)
	}

	if hasPrefix(
		text,
		at,
		blockCommentCloser,
	) {
		return closesComment,
			at + len(blockCommentCloser)
	}

	return 0,
		at + 1
}

func charOrLifetimeEnd(
	text []byte,
	at int,
) (end int) {
	rest := text[at+1:]
	if bytes.HasPrefix(
		rest,
		[]byte{escape},
	) {
		return escapedCharEnd(
			text,
			at,
		)
	}

	_, width := utf8.DecodeRune(rest)
	isChar := width > 0 && bytes.HasPrefix(
		rest[width:],
		[]byte{charQuote},
	)
	if isChar {
		return at + 1 + width + 1
	}

	return at + 1
}

func escapedCharEnd(
	text []byte,
	at int,
) (end int) {
	escapeEnd := min(
		at+1+escapeWidth,
		len(text),
	)
	closing := bytes.IndexByte(
		text[escapeEnd:lineEnd(
			text,
			escapeEnd,
		)],
		charQuote,
	)
	if closing < 0 {
		return at + 1
	}

	return escapeEnd + closing + 1
}

func rawStringEnd(
	text []byte,
	at int,
) (end int) {
	opening := rawStringOpening(
		text,
		at,
	)
	if opening == 0 {
		return at + 1
	}

	closer := string(stringQuote) + strings.Repeat(
		rawHash,
		bytes.Count(
			text[at:at+opening],
			[]byte(rawHash),
		),
	)
	offset := bytes.Index(
		text[at+opening:],
		[]byte(closer),
	)
	if offset < 0 {
		return len(text)
	}

	return at + opening + offset + len(closer)
}

func rawStringOpening(
	text []byte,
	at int,
) (width int) {
	isInsideIdentifier := at > 0 &&
		isIdentifierByte(text[at-1])
	if isInsideIdentifier {
		return 0
	}

	rest := text[at:]
	afterKind := rest
	if strings.IndexByte(
		stringKinds,
		rest[0],
	) >= 0 {
		afterKind = rest[1:]
	}

	afterRaw, isRaw := bytes.CutPrefix(
		afterKind,
		[]byte(rawMarker),
	)
	if !isRaw {
		return 0
	}

	return rawDelimiterWidth(
		rest,
		len(rest)-len(afterRaw),
	)
}

func rawDelimiterWidth(
	rest []byte,
	prefixWidth int,
) (width int) {
	afterHashes := bytes.TrimLeft(
		rest[prefixWidth:],
		rawHash,
	)
	if !bytes.HasPrefix(
		afterHashes,
		[]byte{stringQuote},
	) {
		return 0
	}

	return len(rest) - len(afterHashes) + 1
}
