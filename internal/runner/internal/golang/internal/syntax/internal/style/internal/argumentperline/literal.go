package argumentperline

import (
	"go/ast"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	maxInlineElements = 3
	literalSnippet    = "composite literal with several elements: put every element on its own line, after an opening brace that ends the line and before a closing brace on a line of its own; only up to 3 plain identifiers or basic literals, or a table-driven test row, may stay on one line"
)

func literalFindings(
	file source.File,
) (findings []report.Finding) {
	for _, literal := range brokenLiterals(file) {
		findings = append(
			findings,
			report.Finding{
				Rule:     rule,
				Location: file.Location(literal),
				Snippet:  literalSnippet,
			},
		)
	}

	return findings
}

func brokenLiterals(
	file source.File,
) (broken []*ast.CompositeLit) {
	rows := testRows(file)
	for _, literal := range source.Nodes[*ast.CompositeLit](file.Syntax) {
		broken = append(
			broken,
			brokenLiteral(
				file,
				literal,
				rows,
			)...,
		)
	}

	return broken
}

func brokenLiteral(
	file source.File,
	literal *ast.CompositeLit,
	rows []*ast.CompositeLit,
) (broken []*ast.CompositeLit) {
	isCompliant := len(literal.Elts) < minimumArguments ||
		isInRow(
			literal,
			rows,
		) ||
		isShortInline(
			file,
			literal,
		) ||
		isOnePerLine(
			file,
			literal.Lbrace,
			elementExtents(literal),
			literal.Rbrace,
		)
	if isCompliant {
		return nil
	}

	return []*ast.CompositeLit{literal}
}

func isShortInline(
	file source.File,
	literal *ast.CompositeLit,
) (inline bool) {
	return isSingleLine(
		file,
		literal,
	) &&
		len(literal.Elts) <= maxInlineElements &&
		!slices.ContainsFunc(
			literal.Elts,
			isCompound,
		)
}

func isCompound(
	element ast.Expr,
) (compound bool) {
	pair, isPair := element.(*ast.KeyValueExpr)
	if isPair {
		return isCompound(pair.Key) || isCompound(pair.Value)
	}

	return !isPlain(element)
}

func isPlain(
	element ast.Expr,
) (plain bool) {
	_, isIdent := element.(*ast.Ident)
	_, isBasic := element.(*ast.BasicLit)
	return isIdent || isBasic
}

func isSingleLine(
	file source.File,
	literal *ast.CompositeLit,
) (single bool) {
	return lineOf(
		file,
		literal.Lbrace,
	) == lineOf(
		file,
		literal.Rbrace,
	)
}

func elementExtents(
	literal *ast.CompositeLit,
) (extents []ast.Node) {
	for _, element := range literal.Elts {
		extents = append(
			extents,
			element,
		)
	}

	return extents
}
