package entrypoint

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	mainFunction  = "main"
	funcKind      = "func"
	methodKind    = "method"
	typeKind      = "type"
	varKind       = "var"
	commandFormat = "file %s in package main directory %s: only main.go holds code in a command; move the implementation under %s and call its entry function from func main"
	strayFormat   = "%s %s in %s: main.go holds only func main, which only delegates to an entry function; move it under %s"
	nilName       = "nil"
	errorName     = "err"
	errorSuffix   = "Err"
	pairedCount   = 2
)

func mainFindings(
	found unit,
) (findings []report.Finding) {
	for _, declaration := range found.file.Syntax.Decls {
		findings = append(
			findings,
			mainDeclarationFindings(
				found,
				declaration,
			)...,
		)
	}

	return findings
}

func mainDeclarationFindings(
	found unit,
	declaration ast.Decl,
) (findings []report.Finding) {
	switch typed := declaration.(type) {
	case *ast.FuncDecl:
		return mainFunctionFindings(
			found,
			typed,
		)
	case *ast.GenDecl:
		return slices.Concat(
			typeFindings(
				found,
				typed,
			),
			variableFindings(
				found,
				typed,
			),
		)
	}

	return nil
}

func mainFunctionFindings(
	found unit,
	function *ast.FuncDecl,
) (findings []report.Finding) {
	if function.Recv != nil {
		return strayFindings(
			found,
			methodKind,
			function.Name,
		)
	}

	if function.Name.Name != mainFunction {
		return strayFindings(
			found,
			funcKind,
			function.Name,
		)
	}

	return delegationFindings(
		found,
		function,
	)
}

func typeFindings(
	found unit,
	general *ast.GenDecl,
) (findings []report.Finding) {
	if general.Tok != token.TYPE {
		return nil
	}

	for _, spec := range source.Nodes[*ast.TypeSpec](general) {
		findings = append(
			findings,
			strayFindings(
				found,
				typeKind,
				spec.Name,
			)...,
		)
	}

	return findings
}

func variableFindings(
	found unit,
	general *ast.GenDecl,
) (findings []report.Finding) {
	if general.Tok != token.VAR {
		return nil
	}

	for _, spec := range source.Nodes[*ast.ValueSpec](general) {
		findings = append(
			findings,
			runningFindings(
				found,
				spec,
			)...,
		)
	}

	return findings
}

func runningFindings(
	found unit,
	spec *ast.ValueSpec,
) (findings []report.Finding) {
	running := slices.ContainsFunc(
		spec.Values,
		runsCode,
	)
	if !running {
		return nil
	}

	for _, name := range spec.Names {
		findings = append(
			findings,
			strayFindings(
				found,
				varKind,
				name,
			)...,
		)
	}

	return findings
}

func runsCode(
	value ast.Expr,
) (runs bool) {
	return len(source.Nodes[*ast.CallExpr](value))+len(source.Nodes[*ast.FuncLit](value)) > 0
}

func strayFindings(
	found unit,
	kind string,
	name *ast.Ident,
) (findings []report.Finding) {
	return []report.Finding{{
		Rule:     apiRule,
		Location: found.file.Location(name),
		Snippet: fmt.Sprintf(
			strayFormat,
			kind,
			name.Name,
			found.file.Rel,
			internalPath(found.directory),
		),
	}}
}

func commandFindings(
	found unit,
) (findings []report.Finding) {
	return []report.Finding{{
		Rule:     apiRule,
		Location: found.file.Location(found.file.Syntax.Name),
		Snippet: fmt.Sprintf(
			commandFormat,
			filepath.Base(found.file.Rel),
			found.directory,
			internalPath(found.directory),
		),
	}}
}

func guardConditions(
	body *ast.BlockStmt,
	fileRole role,
) (conditions []ast.Expr) {
	errorNames := errorVariables(body)
	for _, statement := range source.Nodes[*ast.IfStmt](body) {
		conditions = append(
			conditions,
			guardCondition(
				statement,
				errorNames,
				fileRole,
			)...,
		)
	}

	return conditions
}

func guardCondition(
	statement *ast.IfStmt,
	errorNames []string,
	fileRole role,
) (conditions []ast.Expr) {
	condition, isBinary := ast.Unparen(statement.Cond).(*ast.BinaryExpr)
	isGuard := isBinary && isErrorCheck(
		condition,
		errorNames,
	) && statement.Else == nil && isExitBody(
		statement.Body,
		fileRole,
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
	fileRole role,
) (matches bool) {
	if fileRole == mainRole {
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
