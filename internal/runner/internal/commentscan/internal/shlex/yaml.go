package shlex

import (
	"bytes"
	"regexp"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
)

const (
	blockHeaderPattern = `^(\s*(?:-\s+)*)(?:([^\s#:-][^#:]*?)\s*:\s+)?[|>][-+0-9]*\s*(?:#.*)?$`
	sequenceDash       = "-"
	keyQuotes          = "'\""
	maskByte           = ' '
	prefixGroup        = 1
	keyGroup           = 2
)

func yamlSpans(
	text []byte,
) (spans []lexeme.Span) {
	return shellSpans(masked(
		text,
		dataBlocks(text),
	))
}

func dataBlocks(
	text []byte,
) (blocks []lexeme.Span) {
	header := regexp.MustCompile(blockHeaderPattern)
	cursor := 0
	for cursor < len(text) {
		next, found := blockAt(
			text,
			cursor,
			header,
		)
		blocks = append(
			blocks,
			found...,
		)
		cursor = next
	}

	return blocks
}

func blockAt(
	text []byte,
	at int,
	header *regexp.Regexp,
) (next int, blocks []lexeme.Span) {
	end := lexeme.LineEnd(
		text,
		at,
	)
	next = min(
		end+1,
		len(text),
	)
	match := header.FindSubmatch(text[at:end])
	if match == nil {
		return next,
			nil
	}

	column, isData := blockOwner(
		match[prefixGroup],
		match[keyGroup],
	)
	if !isData {
		return next,
			nil
	}

	contentEnd := blockContentEnd(
		text,
		next,
		column,
	)
	return contentEnd,
		[]lexeme.Span{{Start: next, End: contentEnd}}
}

func blockOwner(
	prefix []byte,
	key []byte,
) (column int, isData bool) {
	if key != nil {
		return len(prefix),
			!slices.Contains(
				shellKeys(),
				strings.Trim(
					string(key),
					keyQuotes,
				),
			)
	}

	dash := bytes.LastIndex(
		prefix,
		[]byte(sequenceDash),
	)
	return dash,
		dash >= 0
}

func blockContentEnd(
	text []byte,
	at int,
	column int,
) (end int) {
	end = at
	for end < len(text) &&
		isBlockContent(
			text[end:lexeme.LineEnd(
				text,
				end,
			)],
			column,
		) {
		end = lexeme.NextLine(
			text,
			end,
		)
	}

	return end
}

func isBlockContent(
	line []byte,
	column int,
) (content bool) {
	indent := len(line) - len(bytes.TrimLeft(
		line,
		horizontalSpace,
	))
	return len(bytes.TrimSpace(line)) == 0 ||
		indent > column
}

func masked(
	text []byte,
	blocks []lexeme.Span,
) (maskedText []byte) {
	previous := 0
	for _, block := range blocks {
		maskedText = append(
			maskedText,
			text[previous:block.Start]...,
		)
		maskedText = append(
			maskedText,
			blanked(text[block.Start:block.End])...,
		)
		previous = block.End
	}

	return append(
		maskedText,
		text[previous:]...,
	)
}

func blanked(
	segment []byte,
) (blank []byte) {
	lines := bytes.Split(
		segment,
		[]byte{lexeme.Newline},
	)
	for index, line := range lines {
		lines[index] = bytes.Repeat(
			[]byte{maskByte},
			len(line),
		)
	}

	return bytes.Join(
		lines,
		[]byte{lexeme.Newline},
	)
}

func shellKeys() (keys []string) {
	return []string{
		"run",
		"script",
		"before_script",
		"after_script",
		"command",
		"commands",
		"entrypoint",
	}
}
