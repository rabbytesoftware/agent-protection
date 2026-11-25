package realinfra

import (
	"go/ast"
	"maps"
	"slices"
	"strings"
)

type scope struct {
	imports map[string]string
	owned   []string
	kinds   map[string]string
	fields  map[string]string
	makers  []string
}

func (s scope) qualified(
	expression ast.Expr,
) (importPath string, name string) {
	selector, isSelector := ast.Unparen(expression).(*ast.SelectorExpr)
	if !isSelector {
		return "",
			""
	}

	qualifier, isIdent := selector.X.(*ast.Ident)
	if !isIdent ||
		qualifier.Obj != nil {
		return "",
			""
	}

	return normalizedPath(s.imports[qualifier.Name]),
		selector.Sel.Name
}

func (s scope) typeKind(
	expression ast.Expr,
) (kind string) {
	star, isStar := expression.(*ast.StarExpr)
	if isStar {
		return s.typeKind(star.X)
	}

	importPath, name := s.qualified(expression)
	return typeKinds()[importPath][name]
}

func (s scope) valueKind(
	expression ast.Expr,
) (kind string) {
	switch typed := ast.Unparen(expression).(type) {
	case *ast.UnaryExpr:
		return s.valueKind(typed.X)
	case *ast.CompositeLit:
		return s.typeKind(typed.Type)
	case *ast.CallExpr:
		return s.constructedKind(typed)
	case *ast.Ident:
		return s.kinds[typed.Name]
	case *ast.SelectorExpr:
		return s.selectedKind(typed)
	}

	return ""
}

func (s scope) constructedKind(
	call *ast.CallExpr,
) (kind string) {
	builtin, isIdent := call.Fun.(*ast.Ident)
	if !isIdent ||
		builtin.Name != newBuiltin ||
		len(call.Args) != 1 {
		return ""
	}

	return s.typeKind(call.Args[0])
}

func (s scope) selectedKind(
	selector *ast.SelectorExpr,
) (kind string) {
	importPath, name := s.qualified(selector)
	if importPath != "" {
		return packageValues()[importPath][name]
	}

	return s.fields[selector.Sel.Name]
}

func (s scope) liveTarget(
	call *ast.CallExpr,
) (live bool) {
	selector, isSelector := ast.Unparen(call.Fun).(*ast.SelectorExpr)
	if !isSelector {
		return false
	}

	importPath, name := s.qualified(selector)
	if importPath != "" {
		return slices.Contains(
			liveFunctions()[importPath],
			name,
		) || hasAnyPrefix(
			name,
			livePrefixes()[importPath],
		)
	}

	return slices.Contains(
		liveMethods()[s.valueKind(selector.X)],
		selector.Sel.Name,
	)
}

func (s scope) isExempt(
	call *ast.CallExpr,
) (exempt bool) {
	selector, isSelector := ast.Unparen(call.Fun).(*ast.SelectorExpr)
	exempt = isSelector && s.mentionsOwned(selector.X)
	for _, argument := range call.Args {
		exempt = exempt || s.mentionsOwned(argument)
	}

	return exempt || s.isEphemeralListen(call) || slices.ContainsFunc(
		call.Args,
		isInMemoryDatabase,
	)
}

func (s scope) isEphemeralListen(
	call *ast.CallExpr,
) (ephemeral bool) {
	importPath, name := s.qualified(call.Fun)
	if importPath != netPath ||
		!strings.HasPrefix(
			name,
			listenPrefix,
		) {
		return false
	}

	return slices.ContainsFunc(
		call.Args,
		isEphemeralAddress,
	)
}

func (s scope) mentionsOwned(
	expression ast.Expr,
) (mentions bool) {
	selected := selectedNames(expression)
	for _, identifier := range nodes[*ast.Ident](expression) {
		mentions = mentions || !slices.Contains(
			selected,
			identifier,
		) && slices.Contains(
			s.owned,
			identifier.Name,
		)
	}

	for _, selector := range nodes[*ast.SelectorExpr](expression) {
		mentions = mentions || s.fields[selector.Sel.Name] == kindServer
	}

	for _, call := range nodes[*ast.CallExpr](expression) {
		mentions = mentions || s.isOwnedSeed(call)
	}

	return mentions
}

func (s scope) isOwnedSeed(
	call *ast.CallExpr,
) (seed bool) {
	importPath, name := s.qualified(call.Fun)
	if importPath != "" {
		return slices.Contains(
			ownedConstructors()[importPath],
			name,
		) || s.isEphemeralListen(call) || slices.ContainsFunc(
			call.Args,
			isInMemoryDatabase,
		)
	}

	switch function := ast.Unparen(call.Fun).(type) {
	case *ast.Ident:
		return slices.Contains(
			s.makers,
			function.Name,
		)
	case *ast.SelectorExpr:
		return function.Sel.Name == testDirectoryMethod || isLocalAddress(function)
	}

	return false
}

func (s scope) bound(
	bindings []binding,
) (grown scope, changed bool) {
	grown = s
	grown.owned = slices.Clone(s.owned)
	grown.kinds = maps.Clone(s.kinds)
	for _, current := range bindings {
		grown = grown.withBinding(current)
	}

	return grown,
		len(grown.owned) != len(s.owned) || len(grown.kinds) != len(s.kinds)
}

func (s scope) withBinding(
	current binding,
) (grown scope) {
	grown = s.withKinds(
		current.names,
		s.bindingKind(current),
	)
	if !s.isOwnedBinding(current) {
		return grown
	}

	for _, name := range current.names {
		grown.owned = appendMissing(
			grown.owned,
			name,
		)
	}

	return grown
}

func (s scope) withKinds(
	names []string,
	kind string,
) (grown scope) {
	grown = s
	if kind == "" {
		return grown
	}

	grown.kinds = maps.Clone(s.kinds)
	for _, name := range names {
		grown.kinds[name] = kind
	}

	return grown
}

func (s scope) bindingKind(
	current binding,
) (kind string) {
	return firstKind(
		s.typeKind(current.declared),
		s.valueKind(current.value),
	)
}

func (s scope) isOwnedBinding(
	current binding,
) (owned bool) {
	return s.typeKind(current.declared) == kindServer || current.value != nil && s.mentionsOwned(current.value)
}

func (s scope) returnsServer(
	function *ast.FuncType,
) (returns bool) {
	if function.Results == nil {
		return false
	}

	for _, result := range function.Results.List {
		returns = returns || s.typeKind(result.Type) == kindServer
	}

	return returns
}
