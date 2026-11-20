package cqs

import (
	"fmt"
	"go/ast"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule      = "R33"
	errorType = "error"
)

// Cqs reports every function or method that both returns a non-error value
// and assigns through its pointer receiver or a pointer parameter (R33): a
// call must either answer a question or change state, never both.
func Cqs(
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
		isNotMixed,
	)
	for _, function := range functions {
		findings = append(
			findings,
			report.Finding{
				Rule:     rule,
				Location: file.Location(function),
				Snippet: fmt.Sprintf(
					"func %s(...) returns a value and mutates its receiver or a pointer parameter: split the query from the command",
					function.Name.Name,
				),
			},
		)
	}
	return findings
}

func isNotMixed(
	function *ast.FuncDecl,
) (excluded bool) {
	return function.Body == nil || !returnsValue(function) || !mutatesPointer(function)
}

func returnsValue(
	function *ast.FuncDecl,
) (returns bool) {
	return function.Type.Results != nil && slices.ContainsFunc(
		function.Type.Results.List,
		isNonErrorResult,
	)
}

func isNonErrorResult(
	field *ast.Field,
) (nonError bool) {
	identifier, isIdentifier := field.Type.(*ast.Ident)
	return !isIdentifier || identifier.Name != errorType
}

func mutatesPointer(
	function *ast.FuncDecl,
) (mutates bool) {
	pointers := pointerNames(function)
	for _, name := range assignedNames(function.Body) {
		mutates = mutates || slices.Contains(
			pointers,
			name,
		)
	}
	return mutates
}

func pointerNames(
	function *ast.FuncDecl,
) (names []string) {
	fields := slices.Concat(
		receiverFields(function),
		function.Type.Params.List,
	)
	for _, field := range slices.DeleteFunc(
		fields,
		isNotPointer,
	) {
		names = append(
			names,
			identifierNames(field.Names)...,
		)
	}
	return names
}

func receiverFields(
	function *ast.FuncDecl,
) (fields []*ast.Field) {
	if function.Recv == nil {
		return nil
	}
	return function.Recv.List
}

func isNotPointer(
	field *ast.Field,
) (excluded bool) {
	_, isPointer := field.Type.(*ast.StarExpr)
	return !isPointer
}

func identifierNames(
	identifiers []*ast.Ident,
) (names []string) {
	for _, identifier := range identifiers {
		names = append(
			names,
			identifier.Name,
		)
	}
	return names
}

func assignedNames(
	body *ast.BlockStmt,
) (names []string) {
	for _, assignment := range source.Nodes[*ast.AssignStmt](body) {
		names = append(
			names,
			mutatedRoots(assignment.Lhs)...,
		)
	}
	return names
}

func mutatedRoots(
	targets []ast.Expr,
) (names []string) {
	for _, target := range targets {
		names = append(
			names,
			mutatedRoot(target),
		)
	}
	return names
}

func mutatedRoot(
	target ast.Expr,
) (name string) {
	switch typed := target.(type) {
	case *ast.SelectorExpr:
		return identifierName(typed.X)
	case *ast.StarExpr:
		return identifierName(typed.X)
	case *ast.IndexExpr:
		return identifierName(typed.X)
	}
	return ""
}

func identifierName(
	expression ast.Expr,
) (name string) {
	identifier, isIdentifier := expression.(*ast.Ident)
	if !isIdentifier {
		return ""
	}
	return identifier.Name
}
