package cqs

import (
	"go/ast"
	"go/types"
	"slices"
	"strings"

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

const (
	signatureSeparator = ","
	resultArrow        = "->"
)

func isStandardMethod(
	function *ast.FuncDecl,
) (standard bool) {
	shape, isKnown := standardMethods()[function.Name.Name]
	return isKnown && function.Recv != nil && slices.Contains(
		shape,
		signatureShape(function.Type),
	)
}

func signatureShape(
	signature *ast.FuncType,
) (shape string) {
	return strings.Join(
		fieldTypes(signature.Params),
		signatureSeparator,
	) + resultArrow + strings.Join(
		fieldTypes(signature.Results),
		signatureSeparator,
	)
}

func fieldTypes(
	list *ast.FieldList,
) (names []string) {
	for _, field := range flatParameters(list) {
		names = append(
			names,
			types.ExprString(field.Type),
		)
	}

	return names
}

func standardMethods() (shapes map[string][]string) {
	return map[string][]string{
		"Read":        {"[]byte->int,error"},
		"Write":       {"[]byte->int,error"},
		"ReadByte":    {"->byte,error"},
		"ReadRune":    {"->rune,int,error"},
		"WriteByte":   {"byte->error"},
		"WriteString": {"string->int,error"},
		"Seek":        {"int64,int->int64,error"},
		"ReadFrom":    {"io.Reader->int64,error"},
		"WriteTo":     {"io.Writer->int64,error"},
		"Pop": {
			"->any",
			"->interface{}",
		},
		"Push": {
			"any->",
			"interface{}->",
		},
		"Set":  {"string->error"},
		"Next": {"->bool"},
		"Scan": {"->bool"},
	}
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
