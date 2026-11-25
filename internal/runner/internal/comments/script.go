package comments

import (
	"bytes"
	"slices"
	"strings"
)

const (
	templateQuote      = '`'
	templateStops      = "`"
	expressionOpener   = "${"
	blockOpener        = '{'
	blockCloser        = '}'
	scriptQuotes       = "'\""
	scriptStringStops  = "\n"
	tripleSlash        = "///"
	schemeColon        = ":"
	referenceDirective = "<reference"
	horizontalSpace    = " \t"
	opensBlock         = 1
	closesBlock        = -1
)

func scriptSpans(
	text []byte,
) (spans []span) {
	cursor := shebangEnd(text)
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
) (closing int, spans []span) {
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
) (next int, spans []span) {
	switch {
	case hasPrefix(
		text,
		at,
		lineCommentOpener,
	):
		return scriptLineComment(
			text,
			at,
		)
	case hasPrefix(
		text,
		at,
		blockCommentOpener,
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
		return quotedEnd(
			text,
			at,
			scriptStringStops,
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
) (end int, spans []span) {
	if isSchemeSeparator(
		text,
		at,
	) {
		return at + len(lineCommentOpener),
			nil
	}

	end, spans = lineComment(
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
	return hasColon &&
		len(scheme) > 0 &&
		isIdentifierByte(scheme[len(scheme)-1])
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
) (end int, spans []span) {
	end = blockEnd(
		text,
		at,
	)
	if isPureAnnotation(text[at:end]) {
		return end,
			nil
	}

	return end,
		[]span{{start: at, end: end}}
}

func isPureAnnotation(
	comment []byte,
) (annotation bool) {
	inner := bytes.TrimSpace(bytes.TrimSuffix(
		bytes.TrimPrefix(
			comment,
			[]byte(blockCommentOpener),
		),
		[]byte(blockCommentCloser),
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
) (end int, spans []span) {
	cursor := at
	for isOpen(
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

	return closedEnd(
			text,
			cursor,
			templateQuote,
		),
		spans
}

func templateSegmentEnd(
	text []byte,
	at int,
) (next int, spans []span) {
	if !hasPrefix(
		text,
		at,
		expressionOpener,
	) {
		return escapedNext(
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
