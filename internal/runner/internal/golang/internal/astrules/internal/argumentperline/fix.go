package argumentperline

import (
	"go/ast"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func Edits(
	file source.File,
) (edits []source.Edit) {
	for _, call := range source.Nodes[*ast.CallExpr](file.Syntax) {
		edits = append(
			edits,
			callEdits(
				file,
				call,
			)...,
		)
	}

	return edits
}

func callEdits(
	file source.File,
	call *ast.CallExpr,
) (edits []source.Edit) {
	if len(call.Args) < minimumArguments {
		return nil
	}

	return source.Breaks(
		file,
		call.Lparen,
		argumentExtents(call),
		call.Rparen,
	)
}

func argumentExtents(
	call *ast.CallExpr,
) (extents []source.Extent) {
	for _, argument := range call.Args {
		extents = append(
			extents,
			source.Extent{Start: argument.Pos(), End: argument.End()},
		)
	}

	return extents
}
