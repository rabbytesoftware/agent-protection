package lualex

import (
	"bytes"
	"path/filepath"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
)

const (
	luaExtension      = ".lua"
	lineCommentOpener = "--"
	longOpener        = '['
	longCloser        = ']'
	longLevel         = "="
	quotes            = "'\""
)

func Lua() (language lexeme.Language) {
	return lexeme.Language{
		Matches:      isLua,
		Spans:        luaSpans,
		Interpreters: []string{"lua", "luajit"},
	}
}

func isLua(
	name string,
) (matched bool) {
	return strings.EqualFold(
		filepath.Ext(name),
		luaExtension,
	)
}

func luaSpans(
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
	case lexeme.HasPrefix(
		text,
		at,
		lineCommentOpener,
	):
		return commentEnd(
			text,
			at,
		)
	case strings.IndexByte(
		quotes,
		text[at],
	) >= 0:
		return lexeme.QuotedEnd(
				text,
				at,
				lexeme.LineStops,
			),
			nil
	}

	return longStringEnd(
			text,
			at,
		),
		nil
}

func commentEnd(
	text []byte,
	at int,
) (end int, spans []lexeme.Span) {
	bodyStart := at + len(lineCommentOpener)
	closer := longCloserAt(
		text,
		bodyStart,
	)
	if closer == "" {
		return lexeme.LineComment(
			text,
			at,
		)
	}

	end = closedAfter(
		text,
		bodyStart,
		closer,
	)
	return end,
		[]lexeme.Span{{Start: at, End: end}}
}

func longStringEnd(
	text []byte,
	at int,
) (end int) {
	closer := longCloserAt(
		text,
		at,
	)
	if closer == "" {
		return at + 1
	}

	return closedAfter(
		text,
		at,
		closer,
	)
}

func longCloserAt(
	text []byte,
	at int,
) (closer string) {
	if at >= len(text) ||
		text[at] != longOpener {
		return ""
	}

	rest := text[at+1:]
	level := len(rest) - len(bytes.TrimLeft(
		rest,
		longLevel,
	))
	isLong := level < len(rest) && rest[level] == longOpener
	if !isLong {
		return ""
	}

	return string(longCloser) + strings.Repeat(
		longLevel,
		level,
	) + string(longCloser)
}

func closedAfter(
	text []byte,
	at int,
	closer string,
) (end int) {
	offset := bytes.Index(
		text[at:],
		[]byte(closer),
	)
	if offset < 0 {
		return len(text)
	}

	return at + offset + len(closer)
}
