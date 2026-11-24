package droppederror

import (
	"go/ast"
	"go/token"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func occurrences(
	function *ast.FuncDecl,
	checked *ast.Ident,
) (uses []*ast.Ident, bindings []*ast.Ident) {
	sites := bindingSites(function)
	for _, name := range sameVariable(
		function,
		checked,
	) {
		uses = appendOutside(
			uses,
			name,
			sites,
		)
		bindings = appendInside(
			bindings,
			name,
			sites,
		)
	}

	return uses,
		bindings
}

func appendOutside(
	names []*ast.Ident,
	name *ast.Ident,
	sites []*ast.Ident,
) (appended []*ast.Ident) {
	if slices.Contains(
		sites,
		name,
	) {
		return names
	}

	return append(
		names,
		name,
	)
}

func appendInside(
	names []*ast.Ident,
	name *ast.Ident,
	sites []*ast.Ident,
) (appended []*ast.Ident) {
	if !slices.Contains(
		sites,
		name,
	) {
		return names
	}

	return append(
		names,
		name,
	)
}

func sameVariable(
	function *ast.FuncDecl,
	checked *ast.Ident,
) (names []*ast.Ident) {
	for _, name := range source.Nodes[*ast.Ident](function) {
		names = appendSame(
			names,
			name,
			checked,
		)
	}

	return names
}

func appendSame(
	names []*ast.Ident,
	name *ast.Ident,
	checked *ast.Ident,
) (appended []*ast.Ident) {
	isOther := name != checked && isSameVariable(
		name,
		checked,
	)
	if !isOther {
		return names
	}

	return append(
		names,
		name,
	)
}

func isSameVariable(
	name *ast.Ident,
	checked *ast.Ident,
) (same bool) {
	if checked.Obj != nil {
		return name.Obj == checked.Obj
	}

	return name.Obj == nil && name.Name == checked.Name
}

func bindingSites(
	function *ast.FuncDecl,
) (sites []*ast.Ident) {
	for _, assignment := range source.Nodes[*ast.AssignStmt](function) {
		sites = appendIdentifiers(
			sites,
			assignment.Lhs...,
		)
	}

	for _, spec := range source.Nodes[*ast.ValueSpec](function) {
		sites = append(
			sites,
			spec.Names...,
		)
	}

	for _, field := range source.Nodes[*ast.Field](function) {
		sites = append(
			sites,
			field.Names...,
		)
	}

	for _, loop := range source.Nodes[*ast.RangeStmt](function) {
		sites = appendIdentifiers(
			sites,
			loop.Key,
			loop.Value,
		)
	}

	return sites
}

func appendIdentifiers(
	sites []*ast.Ident,
	expressions ...ast.Expr,
) (appended []*ast.Ident) {
	appended = sites
	for _, expression := range expressions {
		appended = appendIdentifier(
			appended,
			expression,
		)
	}

	return appended
}

func appendIdentifier(
	sites []*ast.Ident,
	expression ast.Expr,
) (appended []*ast.Ident) {
	name, isIdent := expression.(*ast.Ident)
	if !isIdent {
		return sites
	}

	return append(
		sites,
		name,
	)
}

func anyBetween(
	names []*ast.Ident,
	from token.Pos,
	to token.Pos,
) (found bool) {
	for _, name := range names {
		found = found || from < name.Pos() && name.Pos() < to
	}

	return found
}

func latestBefore(
	names []*ast.Ident,
	position token.Pos,
) (latest token.Pos) {
	for _, name := range names {
		latest = max(
			latest,
			earlier(
				name.Pos(),
				position,
			),
		)
	}

	return latest
}

func earlier(
	candidate token.Pos,
	position token.Pos,
) (before token.Pos) {
	if candidate >= position {
		return token.NoPos
	}

	return candidate
}
