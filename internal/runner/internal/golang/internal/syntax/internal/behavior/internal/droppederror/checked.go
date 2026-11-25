package droppederror

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	errorsPackage = "errors"
	partialFormat = "%s handles only the error it matches and drops every other %s: return, wrap or check %s again on the path where it does not match"
)

func checkedFindings(
	file source.File,
	function *ast.FuncDecl,
	callees map[string]int,
) (findings []report.Finding) {
	for _, statement := range source.Nodes[*ast.IfStmt](function) {
		findings = slices.Concat(
			findings,
			ifFindings(
				file,
				function,
				statement,
				callees,
			),
			matchFindings(
				file,
				function,
				statement,
				callees,
			),
		)
	}

	return findings
}

func ifFindings(
	file source.File,
	function *ast.FuncDecl,
	statement *ast.IfStmt,
	callees map[string]int,
) (findings []report.Finding) {
	for _, checked := range nilChecked(statement.Cond) {
		findings = append(
			findings,
			checkFindings(
				file,
				function,
				statement,
				checked,
				callees,
			)...,
		)
	}

	return findings
}

func checkFindings(
	file source.File,
	function *ast.FuncDecl,
	statement *ast.IfStmt,
	checked *ast.Ident,
	callees map[string]int,
) (findings []report.Finding) {
	isDrop := isFailure(
		checked,
		callees,
	) && isDropped(
		function,
		statement,
		checked,
	)
	if !isDrop {
		return nil
	}

	return []report.Finding{finding(
		file,
		checked,
		fmt.Sprintf(
			"%s is checked against nil and then dropped: return it wrapped with %%w, or handle it",
			checked.Name,
		),
	)}
}

func matchFindings(
	file source.File,
	function *ast.FuncDecl,
	statement *ast.IfStmt,
	callees map[string]int,
) (findings []report.Finding) {
	for _, call := range source.Nodes[*ast.CallExpr](statement.Cond) {
		findings = append(
			findings,
			matchFinding(
				file,
				function,
				statement,
				call,
				callees,
			)...,
		)
	}

	return findings
}

func matchFinding(
	file source.File,
	function *ast.FuncDecl,
	statement *ast.IfStmt,
	call *ast.CallExpr,
	callees map[string]int,
) (findings []report.Finding) {
	checked := matchedError(call)
	isDrop := checked != nil &&
		isFailure(
			checked,
			callees,
		) &&
		!isParameter(
			function,
			checked,
		) &&
		!isNamedResult(
			function,
			checked,
		) &&
		!reachesMismatch(
			function,
			statement,
			call,
			checked,
		)
	if !isDrop {
		return nil
	}

	return []report.Finding{finding(
		file,
		call,
		fmt.Sprintf(
			partialFormat,
			types.ExprString(call.Fun),
			checked.Name,
			checked.Name,
		),
	)}
}

func reachesMismatch(
	function *ast.FuncDecl,
	statement *ast.IfStmt,
	call *ast.CallExpr,
	checked *ast.Ident,
) (reaches bool) {
	uses, bindings := occurrences(
		function,
		checked,
	)
	reaches = anyBetween(
		uses,
		call.End(),
		statement.Body.Lbrace,
	) || anyBetween(
		uses,
		statement.End(),
		earliestAfter(
			bindings,
			statement.End(),
			function.End(),
		),
	)
	if statement.Else != nil {
		reaches = reaches || anyBetween(
			uses,
			statement.Else.Pos(),
			statement.Else.End(),
		)
	}

	return reaches || isNegated(
		statement.Cond,
		call,
	) && (anyBetween(
		uses,
		statement.Body.Lbrace,
		statement.Body.Rbrace,
	) || len(source.Nodes[*ast.CallExpr](statement.Body)) > 0)
}

func isNegated(
	condition ast.Expr,
	call *ast.CallExpr,
) (negated bool) {
	for _, unary := range source.Nodes[*ast.UnaryExpr](condition) {
		negated = negated || unary.Op == token.NOT && ast.Unparen(unary.X) == call
	}

	return negated
}

func matchedError(
	call *ast.CallExpr,
) (checked *ast.Ident) {
	selector, isSelector := call.Fun.(*ast.SelectorExpr)
	isMatch := isSelector &&
		len(call.Args) == 2 &&
		identifierName(selector.X) == errorsPackage &&
		slices.Contains(
			matchFunctions(),
			selector.Sel.Name,
		)
	if !isMatch {
		return nil
	}

	checked, isIdent := ast.Unparen(call.Args[0]).(*ast.Ident)
	if !isIdent {
		return nil
	}

	return checked
}

func matchFunctions() (names []string) {
	return []string{
		"Is",
		"As",
	}
}

func nilChecked(
	condition ast.Expr,
) (names []*ast.Ident) {
	for _, comparison := range source.Nodes[*ast.BinaryExpr](condition) {
		names = appendNilChecked(
			names,
			comparison,
		)
	}

	return names
}

func appendNilChecked(
	names []*ast.Ident,
	comparison *ast.BinaryExpr,
) (appended []*ast.Ident) {
	name, isIdent := comparedWithNil(comparison).(*ast.Ident)
	isNotNil := isIdent && comparison.Op == token.NEQ
	if !isNotNil {
		return names
	}

	return append(
		names,
		name,
	)
}

func comparedWithNil(
	comparison *ast.BinaryExpr,
) (compared ast.Expr) {
	isEquality := comparison.Op == token.NEQ || comparison.Op == token.EQL
	if !isEquality {
		return nil
	}

	if isBuiltin(
		comparison.Y,
		nilName,
	) {
		return comparison.X
	}

	if isBuiltin(
		comparison.X,
		nilName,
	) {
		return comparison.Y
	}

	return nil
}

func isDropped(
	function *ast.FuncDecl,
	statement *ast.IfStmt,
	checked *ast.Ident,
) (dropped bool) {
	if isParameter(
		function,
		checked,
	) {
		return false
	}

	uses, bindings := occurrences(
		function,
		checked,
	)
	isUsed := anyBetween(
		uses,
		latestBefore(
			bindings,
			statement.Cond.Pos(),
		),
		statement.End(),
	)
	return !isUsed && isLeftWithout(
		function,
		statement,
		checked,
		uses,
	)
}

func isLeftWithout(
	function *ast.FuncDecl,
	statement *ast.IfStmt,
	checked *ast.Ident,
	uses []*ast.Ident,
) (left bool) {
	switch exit := lastStatement(statement.Body).(type) {
	case *ast.ReturnStmt:
		return len(exit.Results) > 0 || !isNamedResult(
			function,
			checked,
		)
	case *ast.BranchStmt:
		return exit.Tok != token.GOTO && !anyBetween(
			uses,
			statement.Body.Rbrace,
			function.End(),
		)
	}

	return false
}

func lastStatement(
	body *ast.BlockStmt,
) (last ast.Stmt) {
	if len(body.List) == 0 {
		return nil
	}

	return body.List[len(body.List)-1]
}

func isParameter(
	function *ast.FuncDecl,
	checked *ast.Ident,
) (parameter bool) {
	if checked.Obj == nil {
		return false
	}

	field, isField := checked.Obj.Decl.(*ast.Field)
	return isField && slices.Contains(
		function.Type.Params.List,
		field,
	)
}

func isNamedResult(
	function *ast.FuncDecl,
	checked *ast.Ident,
) (named bool) {
	if checked.Obj == nil {
		return false
	}

	field, isField := checked.Obj.Decl.(*ast.Field)
	for _, signature := range source.Nodes[*ast.FuncType](function) {
		named = named || isField && signature.Results != nil && slices.Contains(
			signature.Results.List,
			field,
		)
	}

	return named
}
