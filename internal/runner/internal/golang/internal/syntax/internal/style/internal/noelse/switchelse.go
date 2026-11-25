package noelse

import (
	"go/ast"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	pairSnippet    = "switch with one case and a default is an if/else: write the case as a guard clause that returns early and leave the default as the success path"
	booleanSnippet = "switch on a bool compares it against true or false, an if/else in disguise: write a guard clause on the condition that returns early"
	trueName       = "true"
	falseName      = "false"
	pairClauses    = 2
)

func switchFindings(
	file source.File,
) (findings []report.Finding) {
	for _, statement := range source.Nodes[*ast.SwitchStmt](file.Syntax) {
		findings = append(
			findings,
			disguisedElse(
				file,
				statement,
			)...,
		)
	}

	return findings
}

func disguisedElse(
	file source.File,
	statement *ast.SwitchStmt,
) (findings []report.Finding) {
	clauses := caseClauses(statement.Body)
	if isBooleanSwitch(
		statement,
		clauses,
	) {
		return []report.Finding{{
			Rule:     rule,
			Location: file.Location(statement),
			Snippet:  booleanSnippet,
		}}
	}

	isPair := isTagless(statement.Tag) &&
		len(clauses) == pairClauses &&
		slices.ContainsFunc(
			clauses,
			isDefault,
		)
	if !isPair {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: file.Location(statement),
		Snippet:  pairSnippet,
	}}
}

func caseClauses(
	body *ast.BlockStmt,
) (clauses []*ast.CaseClause) {
	for _, statement := range body.List {
		clauses = append(
			clauses,
			asClause(statement)...,
		)
	}

	return clauses
}

func asClause(
	statement ast.Stmt,
) (clauses []*ast.CaseClause) {
	clause, isClause := statement.(*ast.CaseClause)
	if !isClause {
		return nil
	}

	return []*ast.CaseClause{clause}
}

func isTagless(
	tag ast.Expr,
) (tagless bool) {
	return tag == nil || isIdentNamed(
		tag,
		trueName,
	)
}

func isBooleanSwitch(
	statement *ast.SwitchStmt,
	clauses []*ast.CaseClause,
) (boolean bool) {
	values := []ast.Expr{}
	for _, clause := range clauses {
		values = append(
			values,
			clause.List...,
		)
	}

	return statement.Tag != nil &&
		len(values) > 0 &&
		!slices.ContainsFunc(
			values,
			isNotBooleanLiteral,
		)
}

func isNotBooleanLiteral(
	value ast.Expr,
) (notBoolean bool) {
	return !isIdentNamed(
		value,
		trueName,
	) && !isIdentNamed(
		value,
		falseName,
	)
}

func isIdentNamed(
	expression ast.Expr,
	name string,
) (named bool) {
	ident, isIdent := expression.(*ast.Ident)
	return isIdent && ident.Name == name
}

func isDefault(
	clause *ast.CaseClause,
) (fallback bool) {
	return clause.List == nil
}
