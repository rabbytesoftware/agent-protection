package cqs

import (
	"go/ast"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func isChainable(
	function *ast.FuncDecl,
) (chainable bool) {
	receiver := receiverName(function)
	returns := ownReturns(function.Body)
	isCandidate := receiver != "" &&
		function.Type.Results.NumFields() == 1 &&
		len(returns) > 0
	if !isCandidate {
		return false
	}

	chainable = true
	for _, statement := range returns {
		chainable = chainable && returnsOnly(
			statement,
			receiver,
		)
	}

	return chainable
}

func receiverName(
	function *ast.FuncDecl,
) (name string) {
	receivers := flatParameters(function.Recv)
	if len(receivers) != 1 {
		return ""
	}

	return receivers[0].Names[0].Name
}

func ownReturns(
	body *ast.BlockStmt,
) (returns []*ast.ReturnStmt) {
	nested := []*ast.ReturnStmt{}
	for _, literal := range source.Nodes[*ast.FuncLit](body) {
		nested = append(
			nested,
			source.Nodes[*ast.ReturnStmt](literal)...,
		)
	}

	for _, statement := range source.Nodes[*ast.ReturnStmt](body) {
		returns = appendOwn(
			returns,
			statement,
			nested,
		)
	}

	return returns
}

func appendOwn(
	returns []*ast.ReturnStmt,
	statement *ast.ReturnStmt,
	nested []*ast.ReturnStmt,
) (own []*ast.ReturnStmt) {
	if slices.Contains(
		nested,
		statement,
	) {
		return returns
	}

	return append(
		returns,
		statement,
	)
}

func returnsOnly(
	statement *ast.ReturnStmt,
	name string,
) (only bool) {
	return len(statement.Results) == 1 && identifierName(statement.Results[0]) == name
}
