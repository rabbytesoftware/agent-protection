package jslex

import (
	"bytes"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
)

const (
	templateQuote    = '`'
	templateStops    = "`"
	expressionOpener = "${"
	blockOpener      = '{'
	blockCloser      = '}'
	scriptQuotes     = "'\""

	tripleSlash        = "///"
	schemeColon        = ":"
	referenceDirective = "<reference"
	horizontalSpace    = " \t"
	opensBlock         = 1
	closesBlock        = -1
)

func scriptSpans(
	text []byte,
) (spans []lexeme.Span) {
	cursor := lexeme.ShebangEnd(text)
	for cursor < len(text) {
		closing, found := scriptBlockEnd(
			text,
			cursor,
		)
		spans = append(
			spans,
			found...,
		)
		cursor = closing + 1
	}

	return spans
}

func scriptBlockEnd(
	text []byte,
	at int,
) (closing int, spans []lexeme.Span) {
	depth := 0
	closing = at
	for !closesScriptBlock(
		text,
		closing,
		depth,
	) {
		depth += braceDelta(text[closing])
		next, found := scriptTokenEnd(
			text,
			closing,
		)
		spans = append(
			spans,
			found...,
		)
		closing = next
	}

	return closing,
		spans
}

func closesScriptBlock(
	text []byte,
	cursor int,
	depth int,
) (closes bool) {
	return cursor >= len(text) ||
		(text[cursor] == blockCloser &&
			depth == 0)
}

func braceDelta(
	character byte,
) (delta int) {
	switch character {
	case blockOpener:
		return opensBlock
	case blockCloser:
		return closesBlock
	}

	return 0
}

func scriptTokenEnd(
	text []byte,
	at int,
) (next int, spans []lexeme.Span) {
	switch {
	case lexeme.HasPrefix(
		text,
		at,
		lexeme.LineCommentOpener,
	):
		return scriptLineComment(
			text,
			at,
		)
	case lexeme.HasPrefix(
		text,
		at,
		lexeme.BlockCommentOpener,
	):
		return scriptBlockComment(
			text,
			at,
		)
	case text[at] == templateQuote:
		return templateEnd(
			text,
			at+1,
		)
	}

	return scriptLiteralEnd(
			text,
			at,
		),
		nil
}

func scriptLiteralEnd(
	text []byte,
	at int,
) (end int) {
	if strings.IndexByte(
		scriptQuotes,
		text[at],
	) >= 0 {
		return lexeme.QuotedEnd(
			text,
			at,
			lexeme.LineStops,
		)
	}

	if startsRegex(
		text,
		at,
	) {
		return regexEnd(
			text,
			at,
		)
	}

	return at + 1
}

func scriptLineComment(
	text []byte,
	at int,
) (end int, spans []lexeme.Span) {
	if isSchemeSeparator(
		text,
		at,
	) {
		return at + len(lexeme.LineCommentOpener),
			nil
	}

	end, spans = lexeme.LineComment(
		text,
		at,
	)
	if isReferenceDirective(text[at:end]) {
		return end,
			nil
	}

	return end,
		spans
}

func isSchemeSeparator(
	text []byte,
	at int,
) (separator bool) {
	scheme, hasColon := bytes.CutSuffix(
		text[:at],
		[]byte(schemeColon),
	)
	start := bytes.LastIndexFunc(
		scheme,
		isNotIdentifierRune,
	) + 1
	return hasColon && slices.Contains(
		urlSchemes(),
		strings.ToLower(string(scheme[start:])),
	)
}

func urlSchemes() (schemes []string) {
	return []string{"http", "https", "ws", "wss", "ftp", "file", "mailto"}
}

func isReferenceDirective(
	comment []byte,
) (directive bool) {
	rest, isTripleSlash := bytes.CutPrefix(
		comment,
		[]byte(tripleSlash),
	)
	return isTripleSlash && bytes.HasPrefix(
		bytes.TrimLeft(
			rest,
			horizontalSpace,
		),
		[]byte(referenceDirective),
	)
}

func scriptBlockComment(
	text []byte,
	at int,
) (end int, spans []lexeme.Span) {
	end = lexeme.BlockEnd(
		text,
		at,
	)
	if isPureAnnotation(text[at:end]) {
		return end,
			nil
	}

	return end,
		[]lexeme.Span{{Start: at, End: end}}
}

func isPureAnnotation(
	comment []byte,
) (annotation bool) {
	inner := bytes.TrimSpace(bytes.TrimSuffix(
		bytes.TrimPrefix(
			comment,
			[]byte(lexeme.BlockCommentOpener),
		),
		[]byte(lexeme.BlockCommentCloser),
	))
	return slices.Contains(
		pureAnnotations(),
		string(inner),
	)
}

func pureAnnotations() (annotations []string) {
	return []string{"#__PURE__", "@__PURE__"}
}

func templateEnd(
	text []byte,
	at int,
) (end int, spans []lexeme.Span) {
	cursor := at
	for lexeme.IsOpen(
		text,
		cursor,
		templateStops,
	) {
		next, found := templateSegmentEnd(
			text,
			cursor,
		)
		spans = append(
			spans,
			found...,
		)
		cursor = next
	}

	return lexeme.ClosedEnd(
			text,
			cursor,
			templateQuote,
		),
		spans
}

func templateSegmentEnd(
	text []byte,
	at int,
) (next int, spans []lexeme.Span) {
	if !lexeme.HasPrefix(
		text,
		at,
		expressionOpener,
	) {
		return lexeme.EscapedNext(
				text,
				at,
			),
			nil
	}

	closing, spans := scriptBlockEnd(
		text,
		at+len(expressionOpener),
	)
	return min(
			closing+1,
			len(text),
		),
		spans
}

func Jslex() (language lexeme.Language) {
	return lexeme.Language{
		Matches: matches,
		Spans:   scriptSpans,
	}
}

func matches(
	name string,
) (matched bool) {
	return slices.Contains(
		extensions(),
		filepath.Ext(name),
	)
}

func extensions() (extensions []string) {
	return []string{".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs", ".mts", ".cts"}
}
