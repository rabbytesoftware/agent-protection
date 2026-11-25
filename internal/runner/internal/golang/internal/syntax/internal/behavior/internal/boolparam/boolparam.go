package boolparam

import (
	"fmt"
	"go/ast"
	"path/filepath"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule         = "R32"
	boolType     = "bool"
	plainFormat  = "func %s(...) has a bool parameter: split it into two named functions or take a typed enum"
	namedFormat  = "func %s(...) has a parameter of type %s, whose underlying type is bool: split it into two named functions or take a typed enum with named constants"
	notFoundName = ""
)

func Boolparam(
	files []source.File,
) (findings []report.Finding) {
	packages := map[string][]source.File{}
	for _, file := range files {
		packages[packageKey(file)] = append(
			packages[packageKey(file)],
			file,
		)
	}

	named := map[string][]string{}
	for key, members := range packages {
		named[key] = boolTypes(members)
	}

	for _, file := range files {
		findings = slices.Concat(
			findings,
			fileFindings(
				file,
				named[packageKey(file)],
			),
			flagFieldFindings(
				file,
				packages[packageKey(file)],
			),
			modeFindings(
				file,
				packages[packageKey(file)],
			),
		)
	}

	return findings
}

func packageKey(
	file source.File,
) (key string) {
	return filepath.Join(
		filepath.Dir(file.Path),
		file.Syntax.Name.Name,
	)
}

func boolTypes(
	files []source.File,
) (names []string) {
	specs := []*ast.TypeSpec{}
	for _, file := range files {
		specs = append(
			specs,
			source.Nodes[*ast.TypeSpec](file.Syntax)...,
		)
	}

	underlying := map[string]string{}
	for _, spec := range specs {
		underlying[spec.Name.Name] = identifierName(spec.Type)
	}

	return widened(
		underlying,
		[]string{boolType},
	)
}

func widened(
	underlying map[string]string,
	names []string,
) (closure []string) {
	closure = slices.Clone(names)
	for name, base := range underlying {
		closure = appendDerived(
			closure,
			name,
			base,
		)
	}

	if len(closure) == len(names) {
		return closure
	}

	return widened(
		underlying,
		closure,
	)
}

func appendDerived(
	names []string,
	name string,
	base string,
) (extended []string) {
	isDerived := slices.Contains(
		names,
		base,
	) && !slices.Contains(
		names,
		name,
	)
	if !isDerived {
		return names
	}

	return append(
		names,
		name,
	)
}

func fileFindings(
	file source.File,
	boolNames []string,
) (findings []report.Finding) {
	for _, function := range source.Nodes[*ast.FuncDecl](file.Syntax) {
		findings = appendFinding(
			findings,
			file,
			function,
			boolNames,
		)
	}

	return findings
}

func appendFinding(
	findings []report.Finding,
	file source.File,
	function *ast.FuncDecl,
	boolNames []string,
) (extended []report.Finding) {
	name := boolParamType(
		function,
		boolNames,
	)
	if name == notFoundName {
		return findings
	}

	return append(
		findings,
		report.Finding{
			Rule:     rule,
			Location: file.Location(function),
			Snippet: snippet(
				function.Name.Name,
				name,
			),
		},
	)
}

func snippet(
	function string,
	typeName string,
) (text string) {
	if typeName == boolType {
		return fmt.Sprintf(
			plainFormat,
			function,
		)
	}

	return fmt.Sprintf(
		namedFormat,
		function,
		typeName,
	)
}

func boolParamType(
	function *ast.FuncDecl,
	boolNames []string,
) (name string) {
	for _, field := range function.Type.Params.List {
		name = firstBool(
			name,
			parameterType(field.Type),
			boolNames,
		)
	}

	return name
}

func firstBool(
	found string,
	candidate string,
	boolNames []string,
) (name string) {
	isNew := found == notFoundName && slices.Contains(
		boolNames,
		candidate,
	)
	if !isNew {
		return found
	}

	return candidate
}

func parameterType(
	expression ast.Expr,
) (name string) {
	variadic, isVariadic := expression.(*ast.Ellipsis)
	if isVariadic {
		return identifierName(variadic.Elt)
	}

	return identifierName(expression)
}

func identifierName(
	expression ast.Expr,
) (name string) {
	identifier, isIdentifier := expression.(*ast.Ident)
	if !isIdentifier {
		return notFoundName
	}

	return identifier.Name
}
