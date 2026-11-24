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
