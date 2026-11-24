package cqs

import (
	"go/ast"
	"slices"
)

func directPositions(
	described callable,
) (positions []int) {
	for _, name := range writtenRoots(described.declaration.Body) {
		positions = appendPosition(
			positions,
			described,
			name,
		)
	}

	return positions
}

func appendPosition(
	positions []int,
	described callable,
	name string,
) (extended []int) {
	position, isPointer := described.pointers[name]
	if !isPointer {
		return positions
	}

	return append(
		positions,
		position,
	)
}

func writtenRoots(
	body *ast.BlockStmt,
) (names []string) {
	targets := []ast.Expr{}
	for _, assignment := range liveNodes[*ast.AssignStmt](body) {
		targets = append(
			targets,
			assignment.Lhs...,
		)
	}

	for _, increment := range liveNodes[*ast.IncDecStmt](body) {
		targets = append(
			targets,
			increment.X,
		)
	}

	for _, call := range slices.DeleteFunc(
		liveNodes[*ast.CallExpr](body),
		isNotBuiltinWrite,
	) {
		targets = append(
			targets,
			call.Args[0],
		)
	}

	for _, target := range targets {
		names = append(
			names,
			mutatedRoot(target),
		)
	}

	return names
}

func isNotBuiltinWrite(
	call *ast.CallExpr,
) (excluded bool) {
	return len(call.Args) == 0 || !slices.Contains(
		builtinWrites(),
		identifierName(call.Fun),
	)
}

func mutatedRoot(
	target ast.Expr,
) (name string) {
	_, isIdentifier := ast.Unparen(target).(*ast.Ident)
	if isIdentifier {
		return ""
	}

	return rootName(target)
}

func rootName(
	expression ast.Expr,
) (name string) {
	switch typed := expression.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.SelectorExpr:
		return rootName(typed.X)
	case *ast.StarExpr:
		return rootName(typed.X)
	case *ast.IndexExpr:
		return rootName(typed.X)
	case *ast.ParenExpr:
		return rootName(typed.X)
	}

	return ""
}

func builtinWrites() (names []string) {
	return []string{"delete", "clear", "copy"}
}
