package paragraph

import (
	"go/ast"
	"go/token"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func blankBetween(
	file source.File,
	from token.Pos,
	to token.Pos,
) (blank bool) {
	return len(blankLines(
		file,
		from,
		to,
	)) > 0
}

func blankLines(
	file source.File,
	from token.Pos,
	to token.Pos,
) (lines []int) {
	commented := map[int]bool{}
	for _, group := range file.Syntax.Comments {
		markLines(
			commented,
			groupLines(
				file,
				group,
				from,
				to,
			),
		)
	}

	first := lineOf(
		file,
		from,
	) + 1
	last := lineOf(
		file,
		to,
	) - 1
	for line := first; line <= last; line++ {
		lines = appendUnmarked(
			lines,
			commented,
			line,
		)
	}

	return lines
}

func groupLines(
	file source.File,
	group *ast.CommentGroup,
	from token.Pos,
	to token.Pos,
) (lines []int) {
	first := max(
		lineOf(
			file,
			group.Pos(),
		),
		lineOf(
			file,
			from,
		)+1,
	)
	last := min(
		lineOf(
			file,
			group.End(),
		),
		lineOf(
			file,
			to,
		)-1,
	)
	for line := first; line <= last; line++ {
		lines = append(
			lines,
			line,
		)
	}

	return lines
}

func markLines(
	marked map[int]bool,
	lines []int,
) {
	for _, line := range lines {
		marked[line] = true
	}
}

func appendUnmarked(
	lines []int,
	marked map[int]bool,
	line int,
) (unmarked []int) {
	if marked[line] {
		return lines
	}

	return append(
		lines,
		line,
	)
}

func lineOf(
	file source.File,
	position token.Pos,
) (line int) {
	return file.FileSet.Position(position).Line
}
