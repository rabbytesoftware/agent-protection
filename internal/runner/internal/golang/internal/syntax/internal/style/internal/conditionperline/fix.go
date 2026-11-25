package conditionperline

import (
	"cmp"
	"go/ast"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const lineBreak = "\n"

func Edits(
	file source.File,
) (edits []source.Edit) {
	for _, join := range fileJoins(file) {
		offset := file.Offset(join.Y.Pos())
		edits = append(
			edits,
			source.Edit{
				Start: offset,
				End:   offset,
				Text:  lineBreak,
			},
		)
	}

	return edits
}

func fileJoins(
	file source.File,
) (joins []*ast.BinaryExpr) {
	for _, statement := range source.Nodes[*ast.IfStmt](file.Syntax) {
		joins = append(
			joins,
			collapsedJoins(
				file,
				statement,
			)...,
		)
	}

	slices.SortFunc(
		joins,
		byPosition,
	)
	return slices.Compact(joins)
}

func byPosition(
	first *ast.BinaryExpr,
	second *ast.BinaryExpr,
) (order int) {
	return cmp.Compare(
		first.Pos(),
		second.Pos(),
	)
}
