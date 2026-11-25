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

const blankName = "_"

func blankFindings(
	file source.File,
	callees map[string]int,
) (findings []report.Finding) {
	for _, assignment := range source.Nodes[*ast.AssignStmt](file.Syntax) {
		findings = append(
			findings,
			assignmentFindings(
				file,
				assignment,
				callees,
			)...,
		)
	}

	return findings
}

func assignmentFindings(
	file source.File,
	assignment *ast.AssignStmt,
	callees map[string]int,
) (findings []report.Finding) {
	if isBlankedCall(
		assignment,
		callees,
	) {
		return []report.Finding{finding(
			file,
			assignment.Rhs[0],
			fmt.Sprintf(
				"the error from %s is assigned to _: return it wrapped with %%w, or handle it",
				types.ExprString(assignment.Rhs[0]),
			),
		)}
	}

	if len(assignment.Lhs) != len(assignment.Rhs) {
		return nil
	}

	for index, value := range assignment.Rhs {
		findings = append(
			findings,
			blankFinding(
				file,
				assignment.Lhs[index],
				value,
				callees,
			)...,
		)
	}

	return findings
}

func blankFinding(
	file source.File,
	target ast.Expr,
	value ast.Expr,
	callees map[string]int,
) (findings []report.Finding) {
	isDiscarded := identifierName(target) == blankName && isDroppable(
		value,
		callees,
	)
	if !isDiscarded {
		return nil
	}

	return []report.Finding{finding(
		file,
		value,
		fmt.Sprintf(
			"%s is assigned to _: return it wrapped with %%w, or handle it",
			types.ExprString(value),
		),
	)}
}

func isBlankedCall(
	assignment *ast.AssignStmt,
	callees map[string]int,
) (blanked bool) {
	isSingleCall := len(assignment.Rhs) == 1 && calleeName(assignment.Rhs[0]) != ""
	if !isSingleCall {
		return false
	}

	last := assignment.Lhs[len(assignment.Lhs)-1]
	return identifierName(last) == blankName && resultsOf(
		assignment.Rhs[0],
		callees,
	) == len(assignment.Lhs)
}

func isDroppable(
	value ast.Expr,
	callees map[string]int,
) (droppable bool) {
	name, isIdent := value.(*ast.Ident)
	return isRecoverCall(value) || isIdent && isFailure(
		name,
		callees,
	)
}

func recoverFindings(
	file source.File,
) (findings []report.Finding) {
	statements := slices.DeleteFunc(
		source.Nodes[*ast.ExprStmt](file.Syntax),
		isNotRecoverStatement,
	)
	for _, statement := range statements {
		findings = append(
			findings,
			finding(
				file,
				statement,
				"recover() result is discarded: hand the recovered value to a handler or return it as an error",
			),
		)
	}

	comparisons := slices.DeleteFunc(
		source.Nodes[*ast.BinaryExpr](file.Syntax),
		isNotRecoverCheck,
	)
	for _, comparison := range comparisons {
		findings = append(
			findings,
			finding(
				file,
				comparison,
				"recover() is only compared with nil: hand the recovered value to a handler or return it as an error",
			),
		)
	}

	return findings
}

func isNotRecoverStatement(
	statement *ast.ExprStmt,
) (excluded bool) {
	return !isRecoverCall(statement.X)
}

func isNotRecoverCheck(
	comparison *ast.BinaryExpr,
) (excluded bool) {
	return comparison.Op != token.NEQ || !isRecoverCall(comparedWithNil(comparison))
}

func inlineFindings(
	file source.File,
	callees map[string]int,
) (findings []report.Finding) {
	statements := source.Nodes[*ast.IfStmt](file.Syntax)
	for _, comparison := range source.Nodes[*ast.BinaryExpr](file.Syntax) {
		findings = append(
			findings,
			inlineFinding(
				file,
				comparison,
				statements,
				callees,
			)...,
		)
	}

	return findings
}

func inlineFinding(
	file source.File,
	comparison *ast.BinaryExpr,
	statements []*ast.IfStmt,
	callees map[string]int,
) (findings []report.Finding) {
	call := ast.Unparen(comparedWithNil(comparison))
	isDrop := calleeName(call) != "" && resultsOf(
		call,
		callees,
	) == errorOnly && !isHandledInline(
		file,
		comparison,
		enclosingIf(
			comparison,
			statements,
		),
	)
	if !isDrop {
		return nil
	}

	return []report.Finding{finding(
		file,
		comparison,
		fmt.Sprintf(
			"the error from %s is compared with nil and then dropped: bind it with err := %s, then return it wrapped with %%w, or handle it",
			types.ExprString(call),
			types.ExprString(call),
		),
	)}
}

func isHandledInline(
	file source.File,
	comparison *ast.BinaryExpr,
	statement *ast.IfStmt,
) (handled bool) {
	isChecked := statement != nil && comparison.Op == token.NEQ
	if isChecked {
		return !isExiting(lastStatement(statement.Body))
	}

	return source.IsTest(file.Path)
}

func isExiting(
	last ast.Stmt,
) (exiting bool) {
	switch exit := last.(type) {
	case *ast.ReturnStmt:
		return true
	case *ast.BranchStmt:
		return exit.Tok != token.GOTO
	}

	return false
}

func enclosingIf(
	comparison *ast.BinaryExpr,
	statements []*ast.IfStmt,
) (enclosing *ast.IfStmt) {
	for _, statement := range statements {
		enclosing = innermost(
			enclosing,
			statement,
			comparison,
		)
	}

	return enclosing
}

func innermost(
	enclosing *ast.IfStmt,
	statement *ast.IfStmt,
	comparison *ast.BinaryExpr,
) (inner *ast.IfStmt) {
	isInside := statement.Cond.Pos() <= comparison.Pos() && comparison.End() <= statement.Cond.End()
	if !isInside {
		return enclosing
	}

	return statement
}
