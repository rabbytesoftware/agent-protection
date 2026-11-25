package comments

import "bytes"

func commentLines(
	text []byte,
	spans []span,
) (lines []int) {
	line := 1
	counted := 0
	merged := runs(
		text,
		spans,
	)
	for _, current := range merged {
		line += bytes.Count(
			text[counted:current.start],
			[]byte{newline},
		)
		counted = current.start
		lines = append(
			lines,
			line,
		)
	}

	return lines
}

func runs(
	text []byte,
	spans []span,
) (merged []span) {
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
	merged []span,
	current span,
) (extended []span) {
	if len(merged) == 0 {
		return []span{current}
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
		span{start: last.start, end: current.end, isLine: true},
	)
}

func continuesRun(
	text []byte,
	previous span,
	current span,
) (continues bool) {
	gap := text[previous.end:current.start]
	return previous.isLine &&
		current.isLine &&
		bytes.Count(
			gap,
			[]byte{newline},
		) == 1 &&
		len(bytes.TrimSpace(gap)) == 0
}
