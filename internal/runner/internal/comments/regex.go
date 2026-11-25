package comments

import (
	"bytes"
	"slices"
	"strings"
	"unicode/utf8"
)

const (
	regexDelimiter  = '/'
	regexStops      = "/\n"
	regexPrecursors = "(,=:[!&|?{};"
	classOpener     = '['
	classCloser     = ']'
	classStops      = "]\n"
)

func startsRegex(
	text []byte,
	at int,
) (starts bool) {
	if text[at] != regexDelimiter {
		return false
	}

	before := bytes.TrimRight(
		text[:at],
		horizontalSpace,
	)
	isLineStart := len(before) == 0 ||
		before[len(before)-1] == newline
	if isLineStart {
		return true
	}

	return strings.IndexByte(
		regexPrecursors,
		before[len(before)-1],
	) >= 0 || slices.Contains(
		regexKeywords(),
		string(trailingWord(before)),
	)
}

func regexKeywords() (keywords []string) {
	return []string{"return", "typeof"}
}

func trailingWord(
	text []byte,
) (word []byte) {
	boundary := bytes.LastIndexFunc(
		text,
		isNotIdentifierRune,
	)
	return text[boundary+1:]
}

func isNotIdentifierRune(
	character rune,
) (separator bool) {
	return character < utf8.RuneSelf && !isIdentifierByte(byte(character))
}

func regexEnd(
	text []byte,
	at int,
) (end int) {
	cursor := at + 1
	for isOpen(
		text,
		cursor,
		regexStops,
	) {
		cursor = regexSegmentEnd(
			text,
			cursor,
		)
	}

	return closedEnd(
		text,
		cursor,
		regexDelimiter,
	)
}

func regexSegmentEnd(
	text []byte,
	at int,
) (next int) {
	if text[at] != classOpener {
		return escapedNext(
			text,
			at,
		)
	}

	return closedEnd(
		text,
		scanUntil(
			text,
			at+1,
			classStops,
		),
		classCloser,
	)
}
