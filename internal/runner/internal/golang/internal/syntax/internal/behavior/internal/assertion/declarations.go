package assertion

import (
	"go/ast"
	"path/filepath"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	isFunction      = "function"
	isMethod        = "method"
	rootDirectory   = "."
	benchmarkPrefix = "Benchmark"
)

func fileDeclarations(
	file source.File,
) (found []declaration) {
	imports := importedPaths(file.Syntax)
	for _, function := range source.Nodes[*ast.FuncDecl](file.Syntax) {
		found = appendFunction(
			found,
			file,
			imports,
			function,
		)
	}

	for _, spec := range source.Nodes[*ast.ValueSpec](file.Syntax) {
		found = append(
			found,
			literalDeclarations(
				file,
				imports,
				spec,
			)...,
		)
	}

	return found
}

func appendFunction(
	found []declaration,
	file source.File,
	imports map[string]string,
	function *ast.FuncDecl,
) (extended []declaration) {
	if function.Body == nil {
		return found
	}

	return append(
		found,
		declaration{
			file:     file,
			imports:  imports,
			name:     function.Name.Name,
			receiver: function.Recv,
			function: function.Type,
			body:     function.Body,
		},
	)
}

func literalDeclarations(
	file source.File,
	imports map[string]string,
	spec *ast.ValueSpec,
) (found []declaration) {
	for index, value := range spec.Values {
		found = append(
			found,
			literalDeclaration(
				file,
				imports,
				spec,
				index,
				value,
			)...,
		)
	}

	return found
}

func literalDeclaration(
	file source.File,
	imports map[string]string,
	spec *ast.ValueSpec,
	index int,
	value ast.Expr,
) (found []declaration) {
	literal, isLiteral := value.(*ast.FuncLit)
	if !isLiteral ||
		index >= len(spec.Names) {
		return nil
	}

	return []declaration{{
		file:     file,
		imports:  imports,
		name:     spec.Names[index].Name,
		function: literal.Type,
		body:     literal.Body,
	}}
}

func valueNames(
	files []source.File,
) (values map[string][]string) {
	values = map[string][]string{}
	for _, file := range files {
		directory := filepath.Dir(file.Rel)
		values[directory] = append(
			values[directory],
			fileValueNames(file.Syntax)...,
		)
	}

	return values
}

func fileValueNames(
	syntax *ast.File,
) (names []string) {
	called := calledNames(syntax)
	for _, identifier := range source.Nodes[*ast.Ident](syntax) {
		names = appendValueName(
			names,
			identifier,
			called,
		)
	}

	return names
}

func appendValueName(
	names []string,
	identifier *ast.Ident,
	called []*ast.Ident,
) (extended []string) {
	if slices.Contains(
		called,
		identifier,
	) {
		return names
	}

	return append(
		names,
		identifier.Name,
	)
}

func calledNames(
	syntax ast.Node,
) (called []*ast.Ident) {
	for _, call := range source.Nodes[*ast.CallExpr](syntax) {
		called = appendCalled(
			called,
			unwrapped(call.Fun),
		)
	}

	for _, function := range source.Nodes[*ast.FuncDecl](syntax) {
		called = append(
			called,
			function.Name,
		)
	}

	return called
}

func appendCalled(
	called []*ast.Ident,
	function ast.Expr,
) (extended []*ast.Ident) {
	switch typed := function.(type) {
	case *ast.Ident:
		return append(
			called,
			typed,
		)
	case *ast.SelectorExpr:
		return append(
			called,
			typed.Sel,
		)
	}

	return called
}

func isTestName(
	name string,
) (test bool) {
	for _, prefix := range testPrefixes() {
		test = test || isNamedTest(
			name,
			prefix,
		)
	}

	return test
}

func testPrefixes() (prefixes []string) {
	return []string{
		testPrefix,
		fuzzPrefix,
		benchmarkPrefix,
	}
}

func kindOf(
	candidate declaration,
) (kind string) {
	if candidate.receiver != nil {
		return isMethod
	}

	return isFunction
}

func unwrapped(
	expression ast.Expr,
) (inner ast.Expr) {
	switch typed := ast.Unparen(expression).(type) {
	case *ast.IndexExpr:
		return unwrapped(typed.X)
	case *ast.IndexListExpr:
		return unwrapped(typed.X)
	}

	return ast.Unparen(expression)
}

func identifierName(
	identifier *ast.Ident,
) (name string) {
	if identifier == nil {
		return ""
	}

	return identifier.Name
}
