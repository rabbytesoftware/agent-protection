package cqs

import (
	"go/ast"
	"go/token"
	"slices"
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
			caller,
			passedPointer(argumentAt(
				call,
				calleePosition,
			)),
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

	receiverType := caller.types[identifierName(selector.X)]
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
	selector, isSelector := call.Fun.(*ast.SelectorExpr)
	if !isSelector {
		return identifierName(call.Fun)
	}

	return identifierName(selector.X) + methodSeparator + selector.Sel.Name
}
