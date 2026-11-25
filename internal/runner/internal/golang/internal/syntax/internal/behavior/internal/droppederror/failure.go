package droppederror

import (
	"go/ast"
	"slices"
	"strings"
)

const (
	errorName   = "err"
	errorSuffix = "Err"
	nilName     = "nil"
	recoverName = "recover"
)

func isFailure(
	name *ast.Ident,
	callees map[string]int,
) (failure bool) {
	return name.Name == errorName || strings.HasSuffix(
		name.Name,
		errorSuffix,
	) || isDeclaredFailure(
		name,
		callees,
	)
}

func isDeclaredFailure(
	name *ast.Ident,
	callees map[string]int,
) (failure bool) {
	if name.Obj == nil {
		return false
	}

	switch declaration := name.Obj.Decl.(type) {
	case *ast.Field:
		return isErrorType(declaration.Type)
	case *ast.ValueSpec:
		return isErrorType(declaration.Type) || isBoundToFailure(
			name.Name,
			identifiers(declaration.Names),
			declaration.Values,
			callees,
		)
	case *ast.AssignStmt:
		return isBoundToFailure(
			name.Name,
			declaration.Lhs,
			declaration.Rhs,
			callees,
		)
	}

	return false
}

func isBoundToFailure(
	name string,
	targets []ast.Expr,
	values []ast.Expr,
	callees map[string]int,
) (failure bool) {
	if isRecoveredBinding(
		name,
		targets,
		values,
	) {
		return true
	}

	index := slices.Index(
		identifierNames(targets),
		name,
	)
	isLastResult := index >= 0 && index == len(targets)-1 && len(values) == 1
	return isLastResult && callees[calleeName(values[0])] == len(targets)
}

func isRecoveredBinding(
	name string,
	targets []ast.Expr,
	values []ast.Expr,
) (recovered bool) {
	index := slices.Index(
		identifierNames(targets),
		name,
	)
	return index >= 0 && len(values) == len(targets) && isRecoverCall(values[index])
}

func isCaughtError(
	name *ast.Ident,
	callees map[string]int,
) (caught bool) {
	return isFailure(
		name,
		callees,
	) && !isRecovered(name)
}

func isRecovered(
	name *ast.Ident,
) (recovered bool) {
	if name.Obj == nil {
		return false
	}

	switch declaration := name.Obj.Decl.(type) {
	case *ast.AssignStmt:
		return isRecoveredBinding(
			name.Name,
			declaration.Lhs,
			declaration.Rhs,
		)
	case *ast.ValueSpec:
		return isRecoveredBinding(
			name.Name,
			identifiers(declaration.Names),
			declaration.Values,
		)
	}

	return false
}

func identifiers(
	names []*ast.Ident,
) (expressions []ast.Expr) {
	for _, name := range names {
		expressions = append(
			expressions,
			name,
		)
	}

	return expressions
}

func identifierNames(
	expressions []ast.Expr,
) (names []string) {
	for _, expression := range expressions {
		names = append(
			names,
			identifierName(expression),
		)
	}

	return names
}

func identifierName(
	expression ast.Expr,
) (name string) {
	identifier, isIdent := expression.(*ast.Ident)
	if !isIdent {
		return ""
	}

	return identifier.Name
}

func calleeName(
	expression ast.Expr,
) (name string) {
	call, isCall := expression.(*ast.CallExpr)
	if !isCall {
		return ""
	}

	switch callee := call.Fun.(type) {
	case *ast.Ident:
		return callee.Name
	case *ast.SelectorExpr:
		return callee.Sel.Name
	}

	return ""
}

func isRecoverCall(
	expression ast.Expr,
) (isRecover bool) {
	call, isCall := expression.(*ast.CallExpr)
	return isCall && len(call.Args) == 0 && isBuiltin(
		call.Fun,
		recoverName,
	)
}

func isBuiltin(
	expression ast.Expr,
	name string,
) (builtin bool) {
	identifier, isIdent := expression.(*ast.Ident)
	return isIdent && identifier.Obj == nil && identifier.Name == name
}
