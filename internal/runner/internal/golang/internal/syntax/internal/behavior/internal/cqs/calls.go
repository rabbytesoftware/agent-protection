package cqs

import (
	"go/ast"
	"go/token"
	"go/types"
	"slices"
	"strings"
)

const (
	cancelSuffix = "cancel"
	logName      = "log"
	loggerSuffix = "logger"
)

func mutatingCalls(
	caller callable,
	effects map[string][]int,
) (calls []*ast.CallExpr) {
	for _, call := range caller.calls {
		calls = appendMutating(
			calls,
			caller,
			call,
			effects,
		)
	}

	return calls
}

func appendMutating(
	calls []*ast.CallExpr,
	caller callable,
	call *ast.CallExpr,
	effects map[string][]int,
) (extended []*ast.CallExpr) {
	mutations := callMutations(
		caller,
		call,
		effects,
	)
	if len(mutations) == 0 {
		return calls
	}

	return append(
		calls,
		call,
	)
}

func callPositions(
	caller callable,
	effects map[string][]int,
) (positions []int) {
	for _, call := range caller.calls {
		positions = append(
			positions,
			callMutations(
				caller,
				call,
				effects,
			)...,
		)
	}

	return positions
}

func callMutations(
	caller callable,
	call *ast.CallExpr,
	effects map[string][]int,
) (positions []int) {
	for _, calleePosition := range effects[calleeKey(
		caller,
		call,
	)] {
		positions = appendPosition(
			positions,
			everySlot(caller),
			passedName(
				argumentAt(
					call,
					calleePosition,
				),
				calleePosition,
			),
		)
	}

	return positions
}

func calleeKey(
	caller callable,
	call *ast.CallExpr,
) (key string) {
	selector, isSelector := call.Fun.(*ast.SelectorExpr)
	if !isSelector {
		return identifierName(call.Fun)
	}

	receiverType := typeName(resolvedType(
		caller,
		selector.X,
	))
	if receiverType == "" {
		return ""
	}

	return methodKey(
		receiverType,
		selector.Sel.Name,
	)
}

func argumentAt(
	call *ast.CallExpr,
	position int,
) (argument ast.Expr) {
	arguments := slices.Concat(
		[]ast.Expr{receiverOf(call.Fun)},
		call.Args,
	)
	index := position - receiverPosition
	if index >= len(arguments) {
		return nil
	}

	return arguments[index]
}

func receiverOf(
	function ast.Expr,
) (receiver ast.Expr) {
	selector, isSelector := function.(*ast.SelectorExpr)
	if !isSelector {
		return nil
	}

	return selector.X
}

func passedName(
	argument ast.Expr,
	position int,
) (name string) {
	name = passedPointer(argument)
	isSettled := name != "" || position != receiverPosition
	if isSettled {
		return name
	}

	return rootName(argument)
}

func passedPointer(
	argument ast.Expr,
) (name string) {
	switch typed := argument.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.ParenExpr:
		return passedPointer(typed.X)
	case *ast.UnaryExpr:
		return addressedRoot(typed)
	}

	return ""
}

func addressedRoot(
	address *ast.UnaryExpr,
) (name string) {
	if address.Op != token.AND {
		return ""
	}

	return rootName(address.X)
}

func calleeName(
	call *ast.CallExpr,
) (name string) {
	return types.ExprString(call.Fun)
}

func commandCalls(
	described callable,
) (calls []*ast.CallExpr) {
	for _, call := range described.calls {
		calls = appendCommand(
			calls,
			described,
			call,
		)
	}

	return calls
}

func appendCommand(
	calls []*ast.CallExpr,
	described callable,
	call *ast.CallExpr,
) (extended []*ast.CallExpr) {
	if !isCommand(
		described,
		call,
	) {
		return calls
	}

	return append(
		calls,
		call,
	)
}

func isCommand(
	described callable,
	call *ast.CallExpr,
) (command bool) {
	signature, isSignature := resolvedType(
		described,
		call.Fun,
	).(*ast.FuncType)
	_, isSlot := everySlot(described)[rootName(receiverOf(call.Fun))]
	return isSignature &&
		isSlot &&
		signature.Results.NumFields() == 0 &&
		!isExemptCommand(
			described,
			call.Fun,
		)
}

func commandRoots(
	described callable,
) (names []string) {
	for _, call := range commandCalls(described) {
		names = append(
			names,
			rootName(receiverOf(call.Fun)),
		)
	}

	return names
}

func isExemptCommand(
	described callable,
	function ast.Expr,
) (exempt bool) {
	name := strings.ToLower(lastName(function))
	isLifecycle := slices.Contains(
		lockMethods(),
		name,
	) || strings.HasSuffix(
		name,
		cancelSuffix,
	)
	if isLifecycle {
		return true
	}

	owner := receiverOf(function)
	if owner == nil {
		return false
	}

	return isLoggerName(lastName(owner)) || isLoggerName(ownerTypeName(resolvedType(
		described,
		owner,
	)))
}

func isLoggerName(
	name string,
) (logger bool) {
	lowered := strings.ToLower(name)
	return lowered == logName || strings.HasSuffix(
		lowered,
		loggerSuffix,
	)
}

func ownerTypeName(
	expression ast.Expr,
) (name string) {
	selector, isSelector := elementType(expression).(*ast.SelectorExpr)
	if isSelector {
		return selector.Sel.Name
	}

	return typeName(expression)
}

func lastName(
	expression ast.Expr,
) (name string) {
	switch typed := ast.Unparen(expression).(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.SelectorExpr:
		return typed.Sel.Name
	}

	return ""
}

func lockMethods() (names []string) {
	return []string{
		"lock",
		"unlock",
		"rlock",
		"runlock",
	}
}
