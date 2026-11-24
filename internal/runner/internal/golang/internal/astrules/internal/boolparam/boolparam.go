package boolparam

import (
	"fmt"
	"go/ast"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule     = "R32"
	boolType = "bool"
)

func Boolparam(
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
	functions := slices.DeleteFunc(
		source.Nodes[*ast.FuncDecl](file.Syntax),
		hasNoBoolParam,
	)
	for _, function := range functions {
		findings = append(
			findings,
			report.Finding{
				Rule:     rule,
				Location: file.Location(function),
				Snippet: fmt.Sprintf(
					"func %s(...) has a bool parameter: split it into two named functions or take a typed enum",
					function.Name.Name,
				),
			},
		)
	}

	return findings
}

func hasNoBoolParam(
	function *ast.FuncDecl,
) (clean bool) {
	return !slices.ContainsFunc(
		function.Type.Params.List,
		isBoolField,
	)
}

func isBoolField(
	field *ast.Field,
) (isBool bool) {
	identifier, isIdentifier := field.Type.(*ast.Ident)
	return isIdentifier && identifier.Name == boolType
}
