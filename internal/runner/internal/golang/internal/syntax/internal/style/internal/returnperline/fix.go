package returnperline

import (
	"go/ast"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const lineBreak = "\n"

func Edits(
	file source.File,
) (edits []source.Edit) {
	for _, statement := range source.Nodes[*ast.ReturnStmt](file.Syntax) {
		edits = append(
			edits,
			statementEdits(
				file,
				statement,
			)...,
		)
	}

	return edits
}

func statementEdits(
	file source.File,
	statement *ast.ReturnStmt,
) (edits []source.Edit) {
	for index := 1; index < len(statement.Results); index++ {
		edits = append(
			edits,
			breaks(
				file,
				collapsedValue(
					file,
					statement.Results[index-1],
					statement.Results[index],
				),
			)...,
		)
	}

	return edits
}

func breaks(
	file source.File,
	values []ast.Expr,
) (edits []source.Edit) {
	for _, value := range values {
		offset := file.Offset(value.Pos())
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
