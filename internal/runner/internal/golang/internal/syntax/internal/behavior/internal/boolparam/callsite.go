package boolparam

import (
	"go/ast"
	"go/token"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func builtAtCall(
	members []source.File,
	function string,
	position int,
	structName string,
) (built bool) {
	literals := literalNames(
		members,
		structName,
	)
	arguments := []ast.Expr{}
	for _, file := range members {
		arguments = append(
			arguments,
			callArguments(
				file,
				function,
				position,
			)...,
		)
	}

	for _, argument := range arguments {
		built = built || isBuiltArgument(
			argument,
			structName,
			literals,
		)
	}

	return built
}

func isBuiltArgument(
	argument ast.Expr,
	structName string,
	literals []string,
) (isLiteral bool) {
	return isStructLiteral(
		argument,
		structName,
	) || slices.Contains(
		literals,
		identifierName(argument),
	)
}

func callArguments(
	file source.File,
	function string,
	position int,
) (arguments []ast.Expr) {
	for _, call := range source.Nodes[*ast.CallExpr](file.Syntax) {
		arguments = append(
			arguments,
			callArgument(
				call,
				function,
				position,
			)...,
		)
	}

	return arguments
}

func callArgument(
	call *ast.CallExpr,
	function string,
	position int,
) (arguments []ast.Expr) {
	isCall := calleeName(call.Fun) == function &&
		position < len(call.Args)
	if !isCall {
		return nil
	}

	return []ast.Expr{call.Args[position]}
}

func calleeName(
	callee ast.Expr,
) (name string) {
	selector, isSelector := callee.(*ast.SelectorExpr)
	if isSelector {
		return selector.Sel.Name
	}

	return identifierName(callee)
}

func literalNames(
	members []source.File,
	structName string,
) (names []string) {
	for _, file := range members {
		names = append(
			names,
			fileLiteralNames(
				file,
				structName,
			)...,
		)
	}

	return names
}

func fileLiteralNames(
	file source.File,
	structName string,
) (names []string) {
	for _, assignment := range source.Nodes[*ast.AssignStmt](file.Syntax) {
		names = append(
			names,
			assignedLiterals(
				assignment.Lhs,
				assignment.Rhs,
				structName,
			)...,
		)
	}

	for _, spec := range source.Nodes[*ast.ValueSpec](file.Syntax) {
		names = append(
			names,
			assignedLiterals(
				identExpressions(spec.Names),
				spec.Values,
				structName,
			)...,
		)
	}

	return names
}

func assignedLiterals(
	targets []ast.Expr,
	values []ast.Expr,
	structName string,
) (names []string) {
	for index, value := range values {
		names = append(
			names,
			assignedLiteral(
				targets,
				index,
				value,
				structName,
			)...,
		)
	}

	return names
}

func assignedLiteral(
	targets []ast.Expr,
	index int,
	value ast.Expr,
	structName string,
) (names []string) {
	isAssigned := index < len(targets) && isStructLiteral(
		value,
		structName,
	)
	if !isAssigned {
		return nil
	}

	return []string{identifierName(targets[index])}
}

func identExpressions(
	idents []*ast.Ident,
) (expressions []ast.Expr) {
	for _, ident := range idents {
		expressions = append(
			expressions,
			ident,
		)
	}

	return expressions
}

func isStructLiteral(
	expression ast.Expr,
	structName string,
) (literal bool) {
	switch typed := expression.(type) {
	case *ast.UnaryExpr:
		return typed.Op == token.AND && isStructLiteral(
			typed.X,
			structName,
		)
	case *ast.CompositeLit:
		return identifierName(typed.Type) == structName
	}

	return false
}

func parameterList(
	parameters *ast.FieldList,
) (names []*ast.Ident, kinds []ast.Expr) {
	for _, field := range parameters.List {
		names = append(
			names,
			field.Names...,
		)
		kinds = append(
			kinds,
			slices.Repeat(
				[]ast.Expr{field.Type},
				len(field.Names),
			)...,
		)
	}

	return names,
		kinds
}
