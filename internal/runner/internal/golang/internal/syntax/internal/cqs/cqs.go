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
	directFormat   = "func %s(...) returns a value and mutates its receiver or a pointer, map, slice or channel parameter: split the query from the command"
	indirectFormat = "func %s(...) returns a value and mutates its receiver or a pointer, map, slice or channel parameter through %s(...): split the query from the command"
)

func Cqs(
	files []source.File,
) (findings []report.Finding) {
	checked := slices.DeleteFunc(
		slices.Clone(files),
		isMock,
	)
	effects := map[string]map[string][]int{}
	fields := map[string]map[string]ast.Expr{}
	for key, members := range packagesOf(checked) {
		fields[key] = packageFields(members)
		effects[key] = packageEffects(
			members,
			fields[key],
		)
	}

	for _, file := range checked {
		findings = append(
			findings,
			fileFindings(
				file,
				effects[packageKey(file)],
				fields[packageKey(file)],
			)...,
		)
	}

	return findings
}

func packagesOf(
	files []source.File,
) (packages map[string][]source.File) {
	packages = map[string][]source.File{}
	for _, file := range files {
		packages[packageKey(file)] = append(
			packages[packageKey(file)],
			file,
		)
	}

	return packages
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
	fields map[string]ast.Expr,
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
			fields,
		)
	}

	return findings
}

func appendFinding(
	findings []report.Finding,
	file source.File,
	function *ast.FuncDecl,
	effects map[string][]int,
	fields map[string]ast.Expr,
) (extended []report.Finding) {
	snippet := mixedSnippet(
		callableOf(
			function,
			fields,
		),
		effects,
	)
	if snippet == "" {
		return findings
	}

	return append(
		findings,
		report.Finding{
			Rule:     rule,
			Location: file.Location(function),
			Snippet:  snippet,
		},
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
	if returnsCreatedValue(
		mixed,
		effects,
	) {
		return ""
	}

	name := mixed.declaration.Name.Name
	writes := stdlibCalls(mixed)
	if len(writes) > 0 {
		return fmt.Sprintf(
			indirectFormat,
			name,
			calleeName(writes[0]),
		)
	}

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
