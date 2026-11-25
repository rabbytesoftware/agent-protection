package lexeme

import "bytes"

func Lines(
	text []byte,
	spans []Span,
) (lines []int) {
	line := 1
	counted := 0
	merged := runs(
		text,
		spans,
	)
	for _, current := range merged {
		line += bytes.Count(
			text[counted:current.Start],
			[]byte{Newline},
		)
		counted = current.Start
		lines = append(
			lines,
			line,
		)
	}

	return lines
}

func LineOf(
	text []byte,
	at int,
) (line int) {
	return 1 + bytes.Count(
		text[:at],
		[]byte{Newline},
	)
}

func LineStart(
	text []byte,
	at int,
) (start int) {
	return bytes.LastIndexByte(
		text[:at],
		Newline,
	) + 1
}

func IsLineLeading(
	text []byte,
	span Span,
) (leading bool) {
	return len(bytes.TrimSpace(text[LineStart(
		text,
		span.Start,
	):span.Start])) == 0
}

func runs(
	text []byte,
	spans []Span,
) (merged []Span) {
	for _, current := range spans {
		merged = appendToRun(
			text,
			merged,
			current,
		)
	}

	return merged
}

func appendToRun(
	text []byte,
	merged []Span,
	current Span,
) (extended []Span) {
	if len(merged) == 0 {
		return []Span{current}
	}

	last := merged[len(merged)-1]
	if !continuesRun(
		text,
		last,
		current,
	) {
		return append(
			merged,
			current,
		)
	}

	return append(
		merged[:len(merged)-1],
		Span{
			Start:  last.Start,
			End:    current.End,
			IsLine: true,
		},
	)
}

func continuesRun(
	text []byte,
	previous Span,
	current Span,
) (continues bool) {
	gap := text[previous.End:current.Start]
	return previous.IsLine &&
		current.IsLine &&
		bytes.Count(
			gap,
			[]byte{Newline},
		) == 1 &&
		len(bytes.TrimSpace(gap)) == 0
}
