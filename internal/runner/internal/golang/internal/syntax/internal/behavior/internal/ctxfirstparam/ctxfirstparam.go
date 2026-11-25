package ctxfirstparam

import (
	"go/ast"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule           = "R25"
	contextPackage = "context"
	contextType    = "Context"
)

func Ctxfirstparam(
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
	for _, signature := range source.Nodes[*ast.FuncType](file.Syntax) {
		findings = append(
			findings,
			parameterFindings(
				file,
				signature,
			)...,
		)
	}

	for _, structure := range source.Nodes[*ast.StructType](file.Syntax) {
		findings = append(
			findings,
			fieldFindings(
				file,
				structure,
			)...,
		)
	}

	return append(
		findings,
		storedFindings(file)...,
	)
}

func parameterFindings(
	file source.File,
	signature *ast.FuncType,
) (findings []report.Finding) {
	parameters := signature.Params.List
	if len(parameters) == 0 {
		return nil
	}

	misplaced := slices.DeleteFunc(
		slices.Clone(parameters[1:]),
		isNotContext,
	)
	for _, parameter := range misplaced {
		findings = append(
			findings,
			report.Finding{
				Rule:     rule,
				Location: file.Location(parameter),
				Snippet:  "context.Context is not the first parameter: move ctx to the front",
			},
		)
	}

	return findings
}

func fieldFindings(
	file source.File,
	structure *ast.StructType,
) (findings []report.Finding) {
	stored := slices.DeleteFunc(
		slices.Clone(structure.Fields.List),
		isNotContext,
	)
	for _, field := range stored {
		findings = append(
			findings,
			report.Finding{
				Rule:     rule,
				Location: file.Location(field),
				Snippet:  "struct field stores a context.Context: pass ctx as the first parameter instead",
			},
		)
	}

	return findings
}

func isNotContext(
	field *ast.Field,
) (notContext bool) {
	selector, isSelector := field.Type.(*ast.SelectorExpr)
	if !isSelector {
		return true
	}

	qualifier, isIdent := selector.X.(*ast.Ident)
	isContext := isIdent && qualifier.Name == contextPackage && selector.Sel.Name == contextType
	return !isContext
}
