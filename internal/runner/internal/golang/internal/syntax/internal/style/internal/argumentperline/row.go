package argumentperline

import (
	"go/ast"
	"go/token"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func testRows(
	file source.File,
) (rows []*ast.CompositeLit) {
	if !source.IsTest(file.Path) {
		return nil
	}

	for _, table := range source.Nodes[*ast.CompositeLit](file.Syntax) {
		rows = append(
			rows,
			tableRows(
				file,
				table,
			)...,
		)
	}

	return rows
}

func tableRows(
	file source.File,
	table *ast.CompositeLit,
) (rows []*ast.CompositeLit) {
	if !isCollection(table.Type) {
		return nil
	}

	for _, element := range table.Elts {
		rows = append(
			rows,
			inlineRow(
				file,
				element,
			)...,
		)
	}

	return rows
}

func inlineRow(
	file source.File,
	element ast.Expr,
) (rows []*ast.CompositeLit) {
	row, isLiteral := ast.Unparen(rowValue(element)).(*ast.CompositeLit)
	isInline := isLiteral && isSingleLine(
		file,
		row,
	)
	if !isInline {
		return nil
	}

	return []*ast.CompositeLit{row}
}

func rowValue(
	element ast.Expr,
) (value ast.Expr) {
	pair, isPair := element.(*ast.KeyValueExpr)
	if isPair {
		return rowValue(pair.Value)
	}

	pointer, isPointer := element.(*ast.UnaryExpr)
	isAddress := isPointer &&
		pointer.Op == token.AND
	if isAddress {
		return pointer.X
	}

	return element
}

func isCollection(
	kind ast.Expr,
) (collection bool) {
	_, isArray := kind.(*ast.ArrayType)
	_, isMap := kind.(*ast.MapType)
	return isArray || isMap
}

func isInRow(
	literal *ast.CompositeLit,
	rows []*ast.CompositeLit,
) (inside bool) {
	for _, row := range rows {
		inside = inside || encloses(
			row,
			literal,
		)
	}

	return inside
}

func encloses(
	row *ast.CompositeLit,
	literal *ast.CompositeLit,
) (contains bool) {
	return row.Pos() <= literal.Pos() && literal.End() <= row.End()
}
