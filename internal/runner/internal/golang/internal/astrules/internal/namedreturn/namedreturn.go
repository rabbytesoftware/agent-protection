package namedreturn

import (
	"fmt"
	"go/ast"
	"go/token"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const rule = "R04"

// Namedreturn reports two ways a function hides what it returns: a bare
// return in a function with named results, and a var declared in the body
// only to be returned by name, which belongs in the result list instead.
// Each function and function literal is judged on its own statements.
func Namedreturn(
	files []source.File,
) (findings []report.Finding) {
	for _, file := range files {
		findings = append(
			findings,
			fileFindings(file)...,
		)
	}

	return findings
}

func fileFindings(
	file source.File,
) (findings []report.Finding) {
	for _, function := range source.Nodes[*ast.FuncDecl](file.Syntax) {
		findings = append(
			findings,
			functionFindings(
				file,
				function.Type,
				function.Body,
			)...,
		)
	}

	for _, literal := range source.Nodes[*ast.FuncLit](file.Syntax) {
		findings = append(
			findings,
			functionFindings(
				file,
				literal.Type,
				literal.Body,
			)...,
		)
	}

	return findings
}

func functionFindings(
	file source.File,
	signature *ast.FuncType,
	body *ast.BlockStmt,
) (findings []report.Finding) {
	if body == nil {
		return nil
	}

	literals := source.Nodes[*ast.FuncLit](body)
	returns := ownReturns(
		body,
		literals,
	)
	return slices.Concat(
		nakedFindings(
			file,
			signature,
			returns,
		),
		variableFindings(
			file,
			body,
			literals,
			returns,
		),
	)
}

func nakedFindings(
	file source.File,
	signature *ast.FuncType,
	returns []*ast.ReturnStmt,
) (findings []report.Finding) {
	if !hasNamedResults(signature) {
		return nil
	}

	naked := slices.DeleteFunc(
		slices.Clone(returns),
		hasResults,
	)
	for _, statement := range naked {
		findings = append(
			findings,
			report.Finding{
				Rule:     rule,
				Location: file.Location(statement),
				Snippet:  "naked return: write the returned values explicitly",
			},
		)
	}

	return findings
}

func variableFindings(
	file source.File,
	body *ast.BlockStmt,
	literals []*ast.FuncLit,
	returns []*ast.ReturnStmt,
) (findings []report.Finding) {
	returned := returnedNames(returns)
	for _, declaration := range source.Nodes[*ast.GenDecl](body) {
		findings = append(
			findings,
			declarationFindings(
				file,
				declaration,
				literals,
				returned,
			)...,
		)
	}

	return findings
}

func declarationFindings(
	file source.File,
	declaration *ast.GenDecl,
	literals []*ast.FuncLit,
	returned []string,
) (findings []report.Finding) {
	isOwnVariable := declaration.Tok == token.VAR && !isEnclosed(
		declaration,
		literals,
	)
	if !isOwnVariable {
		return nil
	}

	for _, spec := range source.Nodes[*ast.ValueSpec](declaration) {
		findings = append(
			findings,
			specFindings(
				file,
				spec,
				returned,
			)...,
		)
	}

	return findings
}

func specFindings(
	file source.File,
	spec *ast.ValueSpec,
	returned []string,
) (findings []report.Finding) {
	for _, name := range spec.Names {
		findings = append(
			findings,
			nameFindings(
				file,
				name,
				returned,
			)...,
		)
	}

	return findings
}

func nameFindings(
	file source.File,
	name *ast.Ident,
	returned []string,
) (findings []report.Finding) {
	if !slices.Contains(
		returned,
		name.Name,
	) {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: file.Location(name),
		Snippet: fmt.Sprintf(
			"var %s is declared only to be returned: name it in the result list instead",
			name.Name,
		),
	}}
}

func returnedNames(
	returns []*ast.ReturnStmt,
) (names []string) {
	for _, statement := range returns {
		names = append(
			names,
			identNames(statement.Results)...,
		)
	}

	return names
}

func identNames(
	expressions []ast.Expr,
) (names []string) {
	for _, expression := range expressions {
		names = append(
			names,
			identName(expression)...,
		)
	}

	return names
}

func identName(
	expression ast.Expr,
) (names []string) {
	ident, isIdent := expression.(*ast.Ident)
	if !isIdent {
		return nil
	}

	return []string{ident.Name}
}

func ownReturns(
	body *ast.BlockStmt,
	literals []*ast.FuncLit,
) (returns []*ast.ReturnStmt) {
	for _, statement := range source.Nodes[*ast.ReturnStmt](body) {
		returns = append(
			returns,
			ownReturn(
				statement,
				literals,
			)...,
		)
	}

	return returns
}

func ownReturn(
	statement *ast.ReturnStmt,
	literals []*ast.FuncLit,
) (returns []*ast.ReturnStmt) {
	if isEnclosed(
		statement,
		literals,
	) {
		return nil
	}

	return []*ast.ReturnStmt{statement}
}

func isEnclosed(
	node ast.Node,
	literals []*ast.FuncLit,
) (enclosed bool) {
	for _, literal := range literals {
		enclosed = enclosed || encloses(
			literal,
			node,
		)
	}

	return enclosed
}

func encloses(
	literal *ast.FuncLit,
	node ast.Node,
) (contains bool) {
	return literal.Pos() <= node.Pos() && node.End() <= literal.End()
}

func hasNamedResults(
	signature *ast.FuncType,
) (named bool) {
	return signature.Results != nil && slices.ContainsFunc(
		signature.Results.List,
		isNamed,
	)
}

func isNamed(
	field *ast.Field,
) (named bool) {
	return len(field.Names) > 0
}

func hasResults(
	statement *ast.ReturnStmt,
) (results bool) {
	return len(statement.Results) > 0
}
