package markuplex

import (
	"bytes"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
)

const (
	commentOpener = "<!--"
	commentCloser = "-->"
	cdataOpener   = "<![CDATA["
	cdataCloser   = "]]>"
	actionOpener  = "{{"
	actionCloser  = "}}"
	trimMarker    = "-"
	actionSpace   = " \t\r\n"
	tagCloser     = ">"
	scriptTag     = "script"
	styleTag      = "style"
	tagOpener     = "<"
	endTagOpener  = "</"
	tagNameEnders = " \t\r\n>/"
)

type document struct {
	script lexeme.Language
	style  lexeme.Language
}

func (d document) spans(
	text []byte,
) (spans []lexeme.Span) {
	cursor := 0
	for cursor < len(text) {
		next, found := d.tokenEnd(
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

func (d document) tokenEnd(
	text []byte,
	at int,
) (next int, spans []lexeme.Span) {
	switch {
	case lexeme.HasPrefix(
		text,
		at,
		commentOpener,
	):
		return commentEnd(
			text,
			at,
		)
	case lexeme.HasPrefix(
		text,
		at,
		cdataOpener,
	):
		return closedAfter(
				text,
				at,
				cdataCloser,
			),
			nil
	case lexeme.HasPrefix(
		text,
		at,
		actionOpener,
	):
		return actionEnd(
			text,
			at,
		)
	}

	return d.elementEnd(
		text,
		at,
	)
}

func (d document) elementEnd(
	text []byte,
	at int,
) (next int, spans []lexeme.Span) {
	if opensElement(
		text,
		at,
		scriptTag,
	) {
		return embeddedEnd(
			text,
			at,
			scriptTag,
			d.script,
		)
	}

	if opensElement(
		text,
		at,
		styleTag,
	) {
		return embeddedEnd(
			text,
			at,
			styleTag,
			d.style,
		)
	}

	return at + 1,
		nil
}

func commentEnd(
	text []byte,
	at int,
) (next int, spans []lexeme.Span) {
	next = closedAfter(
		text,
		at+len(commentOpener),
		commentCloser,
	)
	return next,
		[]lexeme.Span{{Start: at, End: next}}
}

func actionEnd(
	text []byte,
	at int,
) (next int, spans []lexeme.Span) {
	bodyStart := at + len(actionOpener)
	inner := bytes.TrimLeft(
		bytes.TrimPrefix(
			text[bodyStart:],
			[]byte(trimMarker),
		),
		actionSpace,
	)
	if !bytes.HasPrefix(
		inner,
		[]byte(lexeme.BlockCommentOpener),
	) {
		return closedAfter(
				text,
				bodyStart,
				actionCloser,
			),
			nil
	}

	next = closedAfter(
		text,
		closedAfter(
			text,
			len(text)-len(inner),
			lexeme.BlockCommentCloser,
		),
		actionCloser,
	)
	return next,
		[]lexeme.Span{{Start: at, End: next}}
}

func embeddedEnd(
	text []byte,
	at int,
	tag string,
	language lexeme.Language,
) (next int, spans []lexeme.Span) {
	bodyStart := closedAfter(
		text,
		at,
		tagCloser,
	)
	bodyEnd := bodyStart + closingTagOffset(
		text[bodyStart:],
		tag,
	)
	for _, span := range language.Spans(text[bodyStart:bodyEnd]) {
		spans = append(
			spans,
			lexeme.Span{
				Start:  bodyStart + span.Start,
				End:    bodyStart + span.End,
				IsLine: span.IsLine,
			},
		)
	}

	return bodyEnd,
		spans
}

func closingTagOffset(
	body []byte,
	tag string,
) (offset int) {
	offset = bytes.Index(
		bytes.ToLower(body),
		[]byte(endTagOpener+tag),
	)
	if offset < 0 {
		return len(body)
	}

	return offset
}

func opensElement(
	text []byte,
	at int,
	tag string,
) (opens bool) {
	nameEnd := at + len(tagOpener) + len(tag)
	isNamed := nameEnd < len(text) && bytes.EqualFold(
		text[at:nameEnd],
		[]byte(tagOpener+tag),
	)
	return isNamed && bytes.IndexByte(
		[]byte(tagNameEnders),
		text[nameEnd],
	) >= 0
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
