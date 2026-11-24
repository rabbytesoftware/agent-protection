package cqs

import (
	"fmt"
	"go/ast"
	"path/filepath"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule           = "R33"
	errorType      = "error"
	directFormat   = "func %s(...) returns a value and mutates its receiver or a pointer parameter: split the query from the command"
	indirectFormat = "func %s(...) returns a value and mutates its receiver or a pointer parameter through %s(...): split the query from the command"
)

// Cqs reports every function or method that both returns a non-error value
// and mutates its pointer receiver or a pointer parameter (R33): a call
// must either answer a question or change state, never both. Mutation is a
// write through the pointer (assignment, increment, delete, clear or copy)
// or handing the pointer to a function or method of the same package that
// mutates it, so moving the write into a helper hides nothing. A chainable
// method, whose every return hands back its own receiver, is a command:
// the receiver it returns is no answer, only the next link. Mocks are
// skipped: a test double keeps the signature of the interface it fakes and
// exists to record the calls made on it.
func Cqs(
	files []source.File,
) (findings []report.Finding) {
	checked := slices.DeleteFunc(
		slices.Clone(files),
		isMock,
	)
	packages := map[string][]source.File{}
	for _, file := range checked {
		packages[packageKey(file)] = append(
			packages[packageKey(file)],
			file,
		)
	}

	effects := map[string]map[string][]int{}
	for key, members := range packages {
		effects[key] = packageEffects(members)
	}

	for _, file := range checked {
		findings = append(
			findings,
			fileFindings(
				file,
				effects[packageKey(file)],
			)...,
		)
	}

	return findings
}

func isMock(
	file source.File,
) (excluded bool) {
	return source.IsMock(file.Rel)
}

func packageKey(
	file source.File,
) (key string) {
	return filepath.Join(
		filepath.Dir(file.Path),
		file.Syntax.Name.Name,
	)
}

func fileFindings(
	file source.File,
	effects map[string][]int,
) (findings []report.Finding) {
	functions := slices.DeleteFunc(
		source.Nodes[*ast.FuncDecl](file.Syntax),
		isNotQuery,
	)
	for _, function := range functions {
		findings = appendFinding(
			findings,
			file,
			function,
			effects,
		)
	}

	return findings
}

func appendFinding(
	findings []report.Finding,
	file source.File,
	function *ast.FuncDecl,
	effects map[string][]int,
) (extended []report.Finding) {
	snippet := mixedSnippet(
		callableOf(function),
		effects,
	)
	if snippet == "" {
		return findings
	}

	return append(
		findings,
		report.Finding{Rule: rule, Location: file.Location(function), Snippet: snippet},
	)
}

func isNotQuery(
	function *ast.FuncDecl,
) (excluded bool) {
	return function.Body == nil || !returnsValue(function) || isChainable(function)
}

func mixedSnippet(
	mixed callable,
	effects map[string][]int,
) (snippet string) {
	name := mixed.declaration.Name.Name
	if len(directPositions(mixed)) > 0 {
		return fmt.Sprintf(
			directFormat,
			name,
		)
	}

	calls := mutatingCalls(
		mixed,
		effects,
	)
	if len(calls) == 0 {
		return ""
	}

	return fmt.Sprintf(
		indirectFormat,
		name,
		calleeName(calls[0]),
	)
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

func identifierName(
	expression ast.Expr,
) (name string) {
	identifier, isIdentifier := expression.(*ast.Ident)
	if !isIdentifier {
		return ""
	}

	return identifier.Name
}
