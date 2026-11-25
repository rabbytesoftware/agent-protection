package linelex

import (
	"bytes"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
)

const (
	dashRule        = "--"
	horizontalSpace = " \t\r"
)

type markers struct {
	openers  []string
	bareRule string
}

func (m markers) spans(
	text []byte,
) (spans []lexeme.Span) {
	cursor := 0
	for cursor < len(text) {
		spans = append(
			spans,
			m.lineSpan(
				text,
				cursor,
			)...,
		)
		cursor = lexeme.NextLine(
			text,
			cursor,
		)
	}

	return spans
}

func (m markers) lineSpan(
	text []byte,
	at int,
) (spans []lexeme.Span) {
	end := lexeme.LineEnd(
		text,
		at,
	)
	line := bytes.TrimRight(
		text[at:end],
		horizontalSpace,
	)
	content := bytes.TrimLeft(
		line,
		horizontalSpace,
	)
	isComment := len(m.bareRule) > 0 && string(content) == m.bareRule
	isComment = isComment || slices.ContainsFunc(
		m.openers,
		opens(content),
	)
	if !isComment {
		return nil
	}

	start := end - len(bytes.TrimLeft(
		text[at:end],
		horizontalSpace,
	))
	return []lexeme.Span{{Start: start, End: end, IsLine: true}}
}

func opens(
	content []byte,
) (matches func(string) bool) {
	return func(
		opener string,
	) (opened bool) {
		return bytes.HasPrefix(
			content,
			[]byte(opener),
		)
	}
}
