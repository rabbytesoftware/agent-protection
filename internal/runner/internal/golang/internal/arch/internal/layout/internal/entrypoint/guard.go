package entrypoint

import (
	"go/ast"
	"go/token"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	nilName     = "nil"
	errorName   = "err"
	errorSuffix = "Err"
	pairedCount = 2
)

func guardConditions(
	body *ast.BlockStmt,
	role string,
) (conditions []ast.Expr) {
	errorNames := errorVariables(body)
	for _, statement := range source.Nodes[*ast.IfStmt](body) {
		conditions = append(
			conditions,
			guardCondition(
				statement,
				errorNames,
				role,
			)...,
		)
	}

	return conditions
}

func guardCondition(
	statement *ast.IfStmt,
	errorNames []string,
	role string,
) (conditions []ast.Expr) {
	condition, isBinary := ast.Unparen(statement.Cond).(*ast.BinaryExpr)
	isGuard := isBinary && isErrorCheck(
		condition,
		errorNames,
	) && statement.Else == nil && isExitBody(
		statement.Body,
		role,
	)
	if !isGuard {
		return nil
	}

	return []ast.Expr{condition}
}

func isErrorCheck(
	condition *ast.BinaryExpr,
	errorNames []string,
) (matches bool) {
	if condition.Op != token.NEQ {
		return false
	}

	return isNil(condition.X) && isErrorVariable(
		condition.Y,
		errorNames,
	) || isNil(condition.Y) && isErrorVariable(
		condition.X,
		errorNames,
	)
}

func isExitBody(
	body *ast.BlockStmt,
	role string,
) (matches bool) {
	if role == mainRole {
		return len(body.List) > 0 && !slices.ContainsFunc(
			body.List,
			isNotExitStatement,
		)
	}

	return len(body.List) == 1 && isReturn(body.List[0])
}

func isNotExitStatement(
	statement ast.Stmt,
) (excluded bool) {
	_, isExpression := statement.(*ast.ExprStmt)
	return !isExpression && !isReturn(statement)
}

func isReturn(
	statement ast.Stmt,
) (matches bool) {
	_, matches = statement.(*ast.ReturnStmt)
	return matches
}

func isNil(
	expression ast.Expr,
) (matches bool) {
	identifier, isIdentifier := ast.Unparen(expression).(*ast.Ident)
	return isIdentifier && identifier.Name == nilName
}

func isErrorVariable(
	expression ast.Expr,
	errorNames []string,
) (matches bool) {
	identifier, isIdentifier := ast.Unparen(expression).(*ast.Ident)
	return isIdentifier && (identifier.Name == errorName || strings.HasSuffix(
		identifier.Name,
		errorSuffix,
	) || slices.Contains(
		errorNames,
		identifier.Name,
	))
}

func errorVariables(
	body *ast.BlockStmt,
) (names []string) {
	for _, statement := range source.Nodes[*ast.AssignStmt](body) {
		names = append(
			names,
			lastCallResult(
				statement.Lhs,
				statement.Rhs,
			)...,
		)
	}

	for _, spec := range source.Nodes[*ast.ValueSpec](body) {
		names = append(
			names,
			lastCallResult(
				identifierExpressions(spec.Names),
				spec.Values,
			)...,
		)
	}

	return names
}

func lastCallResult(
	targets []ast.Expr,
	values []ast.Expr,
) (names []string) {
	isPairedCall := len(targets) >= pairedCount && len(values) == 1 && isCall(values[0])
	if !isPairedCall {
		return nil
	}

	identifier, isIdentifier := targets[len(targets)-1].(*ast.Ident)
	if !isIdentifier {
		return nil
	}

	return []string{identifier.Name}
}

func isCall(
	expression ast.Expr,
) (matches bool) {
	_, matches = ast.Unparen(expression).(*ast.CallExpr)
	return matches
}

func identifierExpressions(
	identifiers []*ast.Ident,
) (expressions []ast.Expr) {
	for _, identifier := range identifiers {
		expressions = append(
			expressions,
			identifier,
		)
	}

	return expressions
}
