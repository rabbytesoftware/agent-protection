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
	gap := lineOf(
		file,
		to,
	) - lineOf(
		file,
		from,
	) - 1
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

	return gap > len(commented)
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

func lineOf(
	file source.File,
	position token.Pos,
) (line int) {
	return file.FileSet.Position(position).Line
}
