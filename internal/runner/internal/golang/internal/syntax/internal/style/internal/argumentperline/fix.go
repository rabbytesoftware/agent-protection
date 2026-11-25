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

	for _, literal := range brokenLiterals(file) {
		edits = append(
			edits,
			source.Breaks(
				file,
				literal.Lbrace,
				elementExtents(literal),
				literal.Rbrace,
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
) (extents []ast.Node) {
	for _, argument := range call.Args {
		extents = append(
			extents,
			argument,
		)
	}

	return extents
}
