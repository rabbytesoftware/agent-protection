package paragraph

import (
	"go/ast"
	"go/token"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func Edits(
	file source.File,
) (edits []source.Edit) {
	for _, statements := range sequences(file.Syntax) {
		edits = append(
			edits,
			sequenceEdits(
				file,
				statements,
			)...,
		)
	}

	for _, block := range source.Nodes[*ast.BlockStmt](file.Syntax) {
		edits = append(
			edits,
			edgeEdits(
				file,
				block,
			)...,
		)
	}

	return edits
}

func sequenceEdits(
	file source.File,
	statements []ast.Stmt,
) (edits []source.Edit) {
	for index := 1; index < len(statements); index++ {
		edits = append(
			edits,
			pairEdits(
				file,
				statements[index-1],
				statements[index],
			)...,
		)
	}

	return edits
}

func pairEdits(
	file source.File,
	previous ast.Stmt,
	current ast.Stmt,
) (edits []source.Edit) {
	switch pairViolation(
		file,
		previous,
		current,
	) {
	case afterBlock:
		return []source.Edit{blankLineAfter(
			file,
			previous,
		)}
	case detachedCheck:
		return blankRemovals(
			file,
			previous.End(),
			current.Pos(),
		)
	}

	return nil
}

func edgeEdits(
	file source.File,
	block *ast.BlockStmt,
) (edits []source.Edit) {
	if len(block.List) == 0 {
		return nil
	}

	return append(
		blankRemovals(
			file,
			block.Lbrace,
			block.List[0].Pos(),
		),
		blankRemovals(
			file,
			block.List[len(block.List)-1].End(),
			block.Rbrace,
		)...,
	)
}

func blankLineAfter(
	file source.File,
	statement ast.Stmt,
) (edit source.Edit) {
	offset := file.LineOffset(lineOf(
		file,
		statement.End(),
	) + 1)
	return source.Edit{Start: offset, End: offset, Text: "\n"}
}

func blankRemovals(
	file source.File,
	from token.Pos,
	to token.Pos,
) (edits []source.Edit) {
	for _, line := range blankLines(
		file,
		from,
		to,
	) {
		edits = append(
			edits,
			source.Edit{Start: file.LineOffset(line), End: file.LineOffset(line + 1)},
		)
	}

	return edits
}
