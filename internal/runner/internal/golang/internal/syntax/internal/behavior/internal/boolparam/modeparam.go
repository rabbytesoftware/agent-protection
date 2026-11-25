package boolparam

import (
	"go/ast"
	"go/token"
	"go/types"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const modeFormat = "func %s(...) uses its %s parameter only to compare it against %s and branch: a mode %s that means true or false is a flag; split it into two named functions or take a typed enum with named constants"

func modeFindings(
	file source.File,
	members []source.File,
) (findings []report.Finding) {
	constants := constantNames(members)
	for _, function := range source.Nodes[*ast.FuncDecl](file.Syntax) {
		findings = append(
			findings,
			functionModeFindings(
				file,
				function,
				constants,
			)...,
		)
	}

	return findings
}

func constantNames(
	members []source.File,
) (names []string) {
	declarations := []*ast.GenDecl{}
	for _, file := range members {
		declarations = append(
			declarations,
			source.Nodes[*ast.GenDecl](file.Syntax)...,
		)
	}

	for _, declaration := range declarations {
		names = append(
			names,
			declaredConstants(declaration)...,
		)
	}

	return names
}

func declaredConstants(
	declaration *ast.GenDecl,
) (names []string) {
	if declaration.Tok != token.CONST {
		return nil
	}

	for _, spec := range source.Nodes[*ast.ValueSpec](declaration) {
		names = append(
			names,
			specNames(spec)...,
		)
	}

	return names
}

func specNames(
	spec *ast.ValueSpec,
) (names []string) {
	for _, name := range spec.Names {
		names = append(
			names,
			name.Name,
		)
	}

	return names
}

func functionModeFindings(
	file source.File,
	function *ast.FuncDecl,
	constants []string,
) (findings []report.Finding) {
	if function.Body == nil {
		return nil
	}

	for _, parameter := range function.Type.Params.List {
		findings = append(
			findings,
			parameterModeFindings(
				file,
				function,
				parameter,
				constants,
			)...,
		)
	}

	return findings
}

func parameterModeFindings(
	file source.File,
	function *ast.FuncDecl,
	parameter *ast.Field,
	constants []string,
) (findings []report.Finding) {
	kind := identifierName(parameter.Type)
	names, _ := parameterList(function.Type.Params)
	isCandidate := len(names) > 1 && slices.Contains(
		modeTypes(),
		kind,
	)
	if !isCandidate {
		return nil
	}

	for _, name := range parameter.Names {
		findings = appendFieldFinding(
			findings,
			file,
			function,
			modeFormat,
			name.Name,
			onlyComparedValue(
				function.Body,
				name.Name,
				constants,
			),
			kind,
		)
	}

	return findings
}

func onlyComparedValue(
	body *ast.BlockStmt,
	parameter string,
	constants []string,
) (literal string) {
	compared := comparedValues(
		branchComparisons(body),
		parameter,
		constants,
	)
	uses := identUses(
		body,
		parameter,
	)
	distinct := slices.Compact(slices.Sorted(slices.Values(compared)))
	isMode := len(compared) > 0 &&
		len(compared) == uses &&
		len(distinct) == 1
	if !isMode {
		return notFoundName
	}

	return compared[0]
}

func identUses(
	body *ast.BlockStmt,
	name string,
) (uses int) {
	for _, ident := range source.Nodes[*ast.Ident](body) {
		uses += countedUse(
			ident,
			name,
		)
	}

	return uses
}

func countedUse(
	ident *ast.Ident,
	name string,
) (count int) {
	if ident.Name != name {
		return 0
	}

	return 1
}

func branchComparisons(
	body *ast.BlockStmt,
) (comparisons []ast.Expr) {
	for _, statement := range source.Nodes[*ast.IfStmt](body) {
		comparisons = append(
			comparisons,
			conjuncts(statement.Cond)...,
		)
	}

	for _, statement := range source.Nodes[*ast.SwitchStmt](body) {
		comparisons = append(
			comparisons,
			switchComparisons(statement)...,
		)
	}

	return comparisons
}

func conjuncts(
	condition ast.Expr,
) (leaves []ast.Expr) {
	switch typed := condition.(type) {
	case *ast.ParenExpr:
		return conjuncts(typed.X)
	case *ast.UnaryExpr:
		return conjuncts(typed.X)
	case *ast.BinaryExpr:
		return binaryConjuncts(typed)
	}

	return []ast.Expr{condition}
}

func conditionLeaves(
	conditions []ast.Expr,
) (leaves []ast.Expr) {
	for _, condition := range conditions {
		leaves = append(
			leaves,
			conjuncts(condition)...,
		)
	}

	return leaves
}

func binaryConjuncts(
	condition *ast.BinaryExpr,
) (leaves []ast.Expr) {
	isLogical := condition.Op == token.LAND || condition.Op == token.LOR
	if !isLogical {
		return []ast.Expr{condition}
	}

	return slices.Concat(
		conjuncts(condition.X),
		conjuncts(condition.Y),
	)
}

func switchComparisons(
	statement *ast.SwitchStmt,
) (comparisons []ast.Expr) {
	if statement.Tag == nil {
		return conditionLeaves(switchConditions(statement))
	}

	for _, clause := range caseClauses(statement) {
		comparisons = append(
			comparisons,
			taggedComparisons(
				statement.Tag,
				clause.List,
			)...,
		)
	}

	return comparisons
}

func caseClauses(
	statement *ast.SwitchStmt,
) (clauses []*ast.CaseClause) {
	for _, member := range statement.Body.List {
		clauses = append(
			clauses,
			asClause(member)...,
		)
	}

	return clauses
}

func asClause(
	member ast.Stmt,
) (clauses []*ast.CaseClause) {
	clause, isClause := member.(*ast.CaseClause)
	if !isClause {
		return nil
	}

	return []*ast.CaseClause{clause}
}

func taggedComparisons(
	tag ast.Expr,
	values []ast.Expr,
) (comparisons []ast.Expr) {
	for _, value := range values {
		comparisons = append(
			comparisons,
			&ast.BinaryExpr{
				X:  tag,
				Op: token.EQL,
				Y:  value,
			},
		)
	}

	return comparisons
}

func comparedValues(
	comparisons []ast.Expr,
	parameter string,
	constants []string,
) (values []string) {
	for _, comparison := range comparisons {
		values = append(
			values,
			comparedValue(
				comparison,
				parameter,
				constants,
			)...,
		)
	}

	return values
}

func comparedValue(
	comparison ast.Expr,
	parameter string,
	constants []string,
) (values []string) {
	binary, isBinary := comparison.(*ast.BinaryExpr)
	isEquality := isBinary &&
		(binary.Op == token.EQL || binary.Op == token.NEQ)
	if !isEquality {
		return nil
	}

	other := otherOperand(
		binary,
		parameter,
	)
	isLiteral := other != nil && isConstant(
		other,
		constants,
	)
	if !isLiteral {
		return nil
	}

	return []string{types.ExprString(other)}
}

func otherOperand(
	binary *ast.BinaryExpr,
	parameter string,
) (other ast.Expr) {
	if identifierName(binary.X) == parameter {
		return binary.Y
	}

	if identifierName(binary.Y) == parameter {
		return binary.X
	}

	return nil
}

func isConstant(
	expression ast.Expr,
	constants []string,
) (constant bool) {
	switch typed := expression.(type) {
	case *ast.BasicLit:
		return true
	case *ast.Ident:
		return slices.Contains(
			constants,
			typed.Name,
		)
	}

	return false
}

func modeTypes() (names []string) {
	return []string{
		"string",
		"int",
		"int8",
		"int16",
		"int32",
		"int64",
		"uint",
		"uint8",
		"uint16",
		"uint32",
		"uint64",
	}
}
