package cqs

import (
	"go/ast"
	"maps"
	"slices"
)

const appendName = "append"

func directPositions(
	described callable,
) (positions []int) {
	for _, name := range writtenRoots(described.declaration.Body) {
		positions = appendPosition(
			positions,
			described.pointers,
			name,
		)
	}

	for _, name := range sharedRoots(described.declaration.Body) {
		positions = appendPosition(
			positions,
			everySlot(described),
			name,
		)
	}

	return positions
}

func appendPosition(
	positions []int,
	slots map[string]int,
	name string,
) (extended []int) {
	position, isSlot := slots[name]
	if !isSlot {
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

func everySlot(
	described callable,
) (slots map[string]int) {
	slots = maps.Clone(described.pointers)
	maps.Copy(
		slots,
		described.references,
	)
	return slots
}

func sharedRoots(
	body *ast.BlockStmt,
) (names []string) {
	for _, assignment := range liveNodes[*ast.AssignStmt](body) {
		names = append(
			names,
			sharedAssignmentRoots(assignment)...,
		)
	}

	for _, increment := range liveNodes[*ast.IncDecStmt](body) {
		names = appendShared(
			names,
			increment.X,
			nil,
		)
	}

	for _, call := range slices.DeleteFunc(
		liveNodes[*ast.CallExpr](body),
		isNotBuiltinWrite,
	) {
		names = append(
			names,
			rootName(call.Args[0]),
		)
	}

	for _, send := range liveNodes[*ast.SendStmt](body) {
		names = append(
			names,
			rootName(send.Chan),
		)
	}

	return names
}

func sharedAssignmentRoots(
	assignment *ast.AssignStmt,
) (names []string) {
	for position, target := range assignment.Lhs {
		names = appendShared(
			names,
			target,
			valueFor(
				assignment.Rhs,
				position,
			),
		)
	}

	return names
}

func appendShared(
	names []string,
	target ast.Expr,
	value ast.Expr,
) (extended []string) {
	if !isSharedWrite(
		target,
		value,
	) {
		return names
	}

	return append(
		names,
		rootName(target),
	)
}

func isSharedWrite(
	target ast.Expr,
	value ast.Expr,
) (shared bool) {
	_, isField := ast.Unparen(target).(*ast.SelectorExpr)
	return throughReference(target) || isField && isAppendCall(value)
}

func throughReference(
	target ast.Expr,
) (through bool) {
	switch typed := target.(type) {
	case *ast.SelectorExpr:
		return throughReference(typed.X)
	case *ast.ParenExpr:
		return throughReference(typed.X)
	case *ast.StarExpr, *ast.IndexExpr:
		return true
	}

	return false
}

func isAppendCall(
	value ast.Expr,
) (appends bool) {
	call, isCall := value.(*ast.CallExpr)
	return isCall && identifierName(call.Fun) == appendName
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
