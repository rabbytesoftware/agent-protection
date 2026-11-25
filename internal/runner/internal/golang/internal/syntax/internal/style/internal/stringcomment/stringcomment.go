package stringcomment

import (
	"go/ast"
	"go/token"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule           = "R07"
	bindingSnippet = "string literal bound to _ is a comment moved into a string: delete it and make the code say it through names and structure"
	blankName      = "_"
)

func Stringcomment(
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
	for _, spec := range source.Nodes[*ast.ValueSpec](file.Syntax) {
		findings = append(
			findings,
			specFindings(
				file,
				spec,
			)...,
		)
	}

	for _, assignment := range source.Nodes[*ast.AssignStmt](file.Syntax) {
		findings = append(
			findings,
			assignmentFindings(
				file,
				assignment,
			)...,
		)
	}

	return slices.Concat(
		findings,
		tagFindings(file),
		logFindings(file),
	)
}

func specFindings(
	file source.File,
	spec *ast.ValueSpec,
) (findings []report.Finding) {
	isBlankString := !slices.ContainsFunc(
		spec.Names,
		isNamed,
	) && slices.ContainsFunc(
		spec.Values,
		isStringLiteral,
	)
	if !isBlankString {
		return nil
	}

	return bindingFinding(
		file,
		spec,
	)
}

func assignmentFindings(
	file source.File,
	assignment *ast.AssignStmt,
) (findings []report.Finding) {
	isBlankString := !slices.ContainsFunc(
		assignment.Lhs,
		isNotBlank,
	) && slices.ContainsFunc(
		assignment.Rhs,
		isStringLiteral,
	)
	if !isBlankString {
		return nil
	}

	return bindingFinding(
		file,
		assignment,
	)
}

func bindingFinding(
	file source.File,
	node ast.Node,
) (findings []report.Finding) {
	return []report.Finding{{
		Rule:     rule,
		Location: file.Location(node),
		Snippet:  bindingSnippet,
	}}
}

func isNamed(
	name *ast.Ident,
) (named bool) {
	return name.Name != blankName
}

func isNotBlank(
	expression ast.Expr,
) (notBlank bool) {
	name, isIdent := expression.(*ast.Ident)
	return !isIdent ||
		isNamed(name)
}

func isStringLiteral(
	expression ast.Expr,
) (literal bool) {
	switch value := expression.(type) {
	case *ast.BasicLit:
		return value.Kind == token.STRING
	case *ast.ParenExpr:
		return isStringLiteral(value.X)
	case *ast.BinaryExpr:
		return isStringLiteral(value.X) &&
			isStringLiteral(value.Y)
	}

	return false
}
