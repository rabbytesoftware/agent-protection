package lexeme

import (
	"bytes"
	"strings"
	"unicode/utf8"
)

const (
	LineCommentOpener  = "//"
	BlockCommentOpener = "/*"
	BlockCommentCloser = "*/"
	Shebang            = "#!"
	Escape             = '\\'
	Newline            = '\n'
	EscapeWidth        = 2
	quotePairWidth     = 2
	NoStops            = ""
	LineStops          = "\n"
	identifierBytes    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_$"
)

func HasPrefix(
	text []byte,
	at int,
	prefix string,
) (matches bool) {
	return bytes.HasPrefix(
		text[at:],
		[]byte(prefix),
	)
}

func LineEnd(
	text []byte,
	at int,
) (end int) {
	offset := bytes.IndexByte(
		text[at:],
		Newline,
	)
	if offset < 0 {
		return len(text)
	}

	return at + offset
}

func ShebangEnd(
	text []byte,
) (end int) {
	if !bytes.HasPrefix(
		text,
		[]byte(Shebang),
	) {
		return 0
	}

	return LineEnd(
		text,
		0,
	)
}

func EscapedNext(
	text []byte,
	at int,
) (next int) {
	if text[at] != Escape {
		return at + 1
	}

	return min(
		at+EscapeWidth,
		len(text),
	)
}

func QuotedEnd(
	text []byte,
	at int,
	stops string,
) (end int) {
	quote := text[at]
	return ClosedEnd(
		text,
		ScanUntil(
			text,
			at+1,
			string(quote)+stops,
		),
		quote,
	)
}

func RawEnd(
	text []byte,
	at int,
) (end int) {
	offset := bytes.IndexByte(
		text[at+1:],
		text[at],
	)
	if offset < 0 {
		return len(text)
	}

	return at + offset + quotePairWidth
}

func ScanUntil(
	text []byte,
	at int,
	stops string,
) (cursor int) {
	cursor = at
	for IsOpen(
		text,
		cursor,
		stops,
	) {
		cursor = EscapedNext(
			text,
			cursor,
		)
	}

	return cursor
}

func IsOpen(
	text []byte,
	cursor int,
	stops string,
) (open bool) {
	return cursor < len(text) && strings.IndexByte(
		stops,
		text[cursor],
	) < 0
}

func ClosedEnd(
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

func LineComment(
	text []byte,
	at int,
) (end int, spans []Span) {
	end = LineEnd(
		text,
		at,
	)
	return end,
		[]Span{{Start: at, End: end, IsLine: true}}
}

func BlockEnd(
	text []byte,
	at int,
) (end int) {
	bodyStart := at + len(BlockCommentOpener)
	offset := bytes.Index(
		text[bodyStart:],
		[]byte(BlockCommentCloser),
	)
	if offset < 0 {
		return len(text)
	}

	return bodyStart + offset + len(BlockCommentCloser)
}

func IsIdentifierByte(
	character byte,
) (identifier bool) {
	return character >= utf8.RuneSelf || strings.IndexByte(
		identifierBytes,
		character,
	) >= 0
}

func Without(
	text []byte,
	spans []Span,
	isDirective func([]byte, Span) bool,
) (kept []Span) {
	for _, span := range spans {
		kept = append(
			kept,
			keptSpan(
				text,
				span,
				isDirective,
			)...,
		)
	}

	return kept
}

func keptSpan(
	text []byte,
	span Span,
	isDirective func([]byte, Span) bool,
) (kept []Span) {
	if isDirective(
		text,
		span,
	) {
		return nil
	}

	return []Span{span}
}

func NextLine(
	text []byte,
	at int,
) (next int) {
	return min(
		LineEnd(
			text,
			at,
		)+1,
		len(text),
	)
}
