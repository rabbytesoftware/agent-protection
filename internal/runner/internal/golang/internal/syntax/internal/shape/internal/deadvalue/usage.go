package deadvalue

import (
	"go/ast"
	"path/filepath"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

type usage struct {
	names    map[string]map[string]bool
	selected map[string]bool
}

func usages(
	files []source.File,
) (used usage) {
	used = usage{
		names:    map[string]map[string]bool{},
		selected: map[string]bool{},
	}
	for _, file := range files {
		record(
			used,
			directoryOf(file),
			file.Syntax,
		)
	}

	return used
}

func record(
	used usage,
	directory string,
	root ast.Node,
) {
	if used.names[directory] == nil {
		used.names[directory] = map[string]bool{}
	}

	skipped := map[*ast.Ident]bool{}
	for node := range ast.Preorder(root) {
		recordNode(
			used,
			directory,
			skipped,
			node,
		)
	}
}

func recordNode(
	used usage,
	directory string,
	skipped map[*ast.Ident]bool,
	node ast.Node,
) {
	switch typed := node.(type) {
	case *ast.SelectorExpr:
		used.selected[typed.Sel.Name] = used.selected[typed.Sel.Name] || !skipped[typed.Sel]
		skipped[typed.Sel] = true
	case *ast.ValueSpec:
		skipBinding(
			skipped,
			identifierExpressions(typed.Names),
			typed.Values,
		)
		skipAll(
			skipped,
			identifierExpressions(typed.Names),
		)
	case *ast.AssignStmt:
		skipBinding(
			skipped,
			typed.Lhs,
			typed.Rhs,
		)
	case *ast.Ident:
		used.names[directory][typed.Name] = used.names[directory][typed.Name] || !skipped[typed]
	}
}

func skipBinding(
	skipped map[*ast.Ident]bool,
	targets []ast.Expr,
	values []ast.Expr,
) {
	if slices.ContainsFunc(
		targets,
		isNamed,
	) {
		return
	}

	skipAll(
		skipped,
		values,
	)
}

func skipAll(
	skipped map[*ast.Ident]bool,
	expressions []ast.Expr,
) {
	for _, expression := range expressions {
		skipIdentifiers(
			skipped,
			expression,
		)
	}
}

func skipIdentifiers(
	skipped map[*ast.Ident]bool,
	expression ast.Expr,
) {
	for _, identifier := range source.Nodes[*ast.Ident](expression) {
		skipped[identifier] = true
	}
}

func isBlank(
	expression ast.Expr,
) (blank bool) {
	identifier, isIdentifier := expression.(*ast.Ident)
	return isIdentifier && identifier.Name == blankName
}

func isNamed(
	expression ast.Expr,
) (named bool) {
	return !isBlank(expression)
}

func identifierExpressions(
	identifiers []*ast.Ident,
) (expressions []ast.Expr) {
	for _, identifier := range identifiers {
		expressions = append(
			expressions,
			identifier,
		)
	}

	return expressions
}

func directoryOf(
	file source.File,
) (directory string) {
	return filepath.ToSlash(filepath.Dir(file.Rel))
}
