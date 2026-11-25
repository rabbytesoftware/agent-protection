package shlex

import (
	"bytes"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
)

const (
	heredocOperator   = "<<"
	quotePairWidth    = 2
	hereStringMarker  = '<'
	stripTabsMarker   = '-'
	horizontalSpace   = " \t"
	delimiterQuotes   = "'\""
	delimiterStarters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_"
	delimiterBytes    = delimiterStarters + "0123456789-"
)

func heredocAt(
	text []byte,
	at int,
) (current token) {
	cursor := at + len(heredocOperator)
	if lexeme.HasPrefix(
		text,
		cursor,
		string(hereStringMarker),
	) {
		return token{next: cursor + 1}
	}

	cursor = skipSpace(
		text,
		skipStripTabs(
			text,
			cursor,
		),
	)
	delimiter, end := delimiterAt(
		text,
		cursor,
	)
	if delimiter == "" {
		return token{next: at + len(heredocOperator)}
	}

	return token{
		next:       end,
		delimiters: []string{delimiter},
	}
}

func skipStripTabs(
	text []byte,
	at int,
) (cursor int) {
	if lexeme.HasPrefix(
		text,
		at,
		string(stripTabsMarker),
	) {
		return at + 1
	}

	return at
}

func skipSpace(
	text []byte,
	at int,
) (cursor int) {
	cursor = at
	for lexeme.IsOpen(
		text,
		cursor,
		lexeme.NoStops,
	) && strings.IndexByte(
		horizontalSpace,
		text[cursor],
	) >= 0 {
		cursor++
	}

	return cursor
}

func delimiterAt(
	text []byte,
	at int,
) (delimiter string, end int) {
	if at >= len(text) {
		return "",
			at
	}

	if strings.IndexByte(
		delimiterQuotes,
		text[at],
	) >= 0 {
		return quotedDelimiter(
			text,
			at,
		)
	}

	cursor := at
	if text[cursor] == lexeme.Escape {
		cursor++
	}

	return wordAt(
		text,
		cursor,
	)
}

func quotedDelimiter(
	text []byte,
	at int,
) (delimiter string, end int) {
	closing := bytes.IndexByte(
		text[at+1:lexeme.LineEnd(
			text,
			at,
		)],
		text[at],
	)
	if closing < 0 {
		return "",
			at + 1
	}

	return string(text[at+1 : at+1+closing]),
		at + closing + quotePairWidth
}

func wordAt(
	text []byte,
	at int,
) (word string, end int) {
	isWordStart := at < len(text) && strings.IndexByte(
		delimiterStarters,
		text[at],
	) >= 0
	if !isWordStart {
		return "",
			at
	}

	end = at
	for lexeme.IsOpen(
		text,
		end,
		lexeme.NoStops,
	) && strings.IndexByte(
		delimiterBytes,
		text[end],
	) >= 0 {
		end++
	}

	return string(text[at:end]),
		end
}

func heredocBodiesEnd(
	text []byte,
	at int,
	delimiters []string,
) (end int) {
	end = min(
		at,
		len(text),
	)
	for _, delimiter := range delimiters {
		end = bodyEnd(
			text,
			end,
			delimiter,
		)
	}

	return end
}

func bodyEnd(
	text []byte,
	at int,
	delimiter string,
) (end int) {
	end = at
	for end < len(text) &&
		!isDelimiterLine(
			text,
			end,
			delimiter,
		) {
		end = lexeme.NextLine(
			text,
			end,
		)
	}

	return lexeme.NextLine(
		text,
		end,
	)
}

func isDelimiterLine(
	text []byte,
	at int,
	delimiter string,
) (matches bool) {
	return strings.TrimSpace(string(text[at:lexeme.LineEnd(
		text,
		at,
	)])) == delimiter
}
