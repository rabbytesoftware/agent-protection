package testmodule

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

const (
	testAttribute = "#[cfg(test)]"
	moduleKeyword = "mod "
	openBrace     = "{"
	closeBrace    = "}"
)

func Testmodule(
	path string,
) (lines []int, err error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil,
			fmt.Errorf(
				"read %s: %w",
				path,
				err,
			)
	}

	source := strings.Split(
		string(content),
		"\n",
	)
	for index := range source {
		lines = append(
			lines,
			moduleLines(
				source,
				index,
			)...,
		)
	}

	return lines,
		nil
}

func moduleLines(
	source []string,
	attribute int,
) (lines []int) {
	if !strings.Contains(
		source[attribute],
		testAttribute,
	) {
		return nil
	}

	open := firstNonBlank(
		source,
		attribute+1,
	)
	isModule := open < len(source) && strings.Contains(
		source[open],
		moduleKeyword,
	)
	if !isModule {
		return nil
	}

	return lineRange(
		open+1,
		closeLine(
			source,
			open,
		),
	)
}

func firstNonBlank(
	source []string,
	from int,
) (index int) {
	offset := slices.IndexFunc(
		source[from:],
		isNotBlank,
	)
	if offset < 0 {
		return len(source)
	}

	return from + offset
}

func closeLine(
	source []string,
	open int,
) (line int) {
	offset := slices.IndexFunc(
		runningDepths(source[open:]),
		isClosed,
	)
	if offset < 0 {
		return len(source)
	}

	return open + offset + 1
}

func runningDepths(
	source []string,
) (depths []int) {
	depth := 0
	for _, text := range source {
		depth += strings.Count(
			text,
			openBrace,
		) - strings.Count(
			text,
			closeBrace,
		)
		depths = append(
			depths,
			depth,
		)
	}

	return depths
}

func lineRange(
	first int,
	last int,
) (lines []int) {
	for line := first; line <= last; line++ {
		lines = append(
			lines,
			line,
		)
	}

	return lines
}

func isNotBlank(
	text string,
) (notBlank bool) {
	return strings.TrimSpace(text) != ""
}

func isClosed(
	depth int,
) (closed bool) {
	return depth <= 0
}
