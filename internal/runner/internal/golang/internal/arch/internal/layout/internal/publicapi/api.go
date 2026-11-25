package publicapi

import (
	"fmt"
	"go/ast"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	containerFile  = "container.go"
	apiFormat      = "%s %s in %s: %s"
	behaviorFormat = "type %s in %s has %d exported func fields (%s): a struct of funcs is behavior, not API; declare an interface with these operations here, implement it with methods under %s and return it from container.go"
	fieldSeparator = ", "
	maxFunctions   = 3
)

func apiFindings(
	file source.File,
	found scope,
	declaration ast.Decl,
) (findings []report.Finding) {
	if !found.api {
		return nil
	}

	function, isFunction := declaration.(*ast.FuncDecl)
	isEntry := isFunction && isEntryFunction(
		file,
		function,
	)
	if isEntry {
		return nil
	}

	return slices.Concat(
		violationFindings(
			file,
			found,
			declaration,
		),
		aliasFindings(
			file,
			found,
			declaration,
		),
		behaviorFindings(
			file,
			found,
			declaration,
		),
	)
}

func behaviorFindings(
	file source.File,
	found scope,
	declaration ast.Decl,
) (findings []report.Finding) {
	for _, spec := range typeSpecs(declaration) {
		findings = append(
			findings,
			behaviorFinding(
				file,
				found,
				spec,
			)...,
		)
	}

	return findings
}

func behaviorFinding(
	file source.File,
	found scope,
	spec *ast.TypeSpec,
) (findings []report.Finding) {
	structure, isStruct := spec.Type.(*ast.StructType)
	if !isStruct {
		return nil
	}

	functions := exportedFunctions(
		found,
		structure,
	)
	if len(functions) <= maxFunctions {
		return nil
	}

	return []report.Finding{{
		Rule:     apiRule,
		Location: file.Location(spec.Name),
		Snippet: fmt.Sprintf(
			behaviorFormat,
			spec.Name.Name,
			found.label,
			len(functions),
			strings.Join(
				functions,
				fieldSeparator,
			),
			found.internal,
		),
	}}
}

func exportedFunctions(
	found scope,
	structure *ast.StructType,
) (functions []string) {
	for _, field := range structure.Fields.List {
		functions = append(
			functions,
			functionNames(
				found,
				field,
				exportedNames(field.Names),
			)...,
		)
	}

	return functions
}

func functionNames(
	found scope,
	field *ast.Field,
	names []string,
) (functions []string) {
	_, isLiteral := ast.Unparen(field.Type).(*ast.FuncType)
	named, isNamed := ast.Unparen(field.Type).(*ast.Ident)
	isFunction := isLiteral || isNamed && slices.Contains(
		found.functions,
		named.Name,
	)
	if !isFunction {
		return nil
	}

	return names
}

func exportedNames(
	identifiers []*ast.Ident,
) (names []string) {
	for _, identifier := range identifiers {
		names = appendExported(
			names,
			identifier,
		)
	}

	return names
}

func appendExported(
	names []string,
	identifier *ast.Ident,
) (extended []string) {
	if !identifier.IsExported() {
		return names
	}

	return append(
		names,
		identifier.Name,
	)
}

func packageFunctions(
	files []source.File,
) (functions map[string][]string) {
	functions = map[string][]string{}
	for _, file := range files {
		directory := filepath.ToSlash(filepath.Dir(file.Rel))
		functions[directory] = append(
			functions[directory],
			functionTypes(file.Syntax)...,
		)
	}

	return functions
}

func functionTypes(
	syntax *ast.File,
) (names []string) {
	for _, spec := range source.Nodes[*ast.TypeSpec](syntax) {
		names = append(
			names,
			functionType(spec)...,
		)
	}

	return names
}

func functionType(
	spec *ast.TypeSpec,
) (names []string) {
	_, isFunction := spec.Type.(*ast.FuncType)
	if !isFunction {
		return nil
	}

	return []string{spec.Name.Name}
}

func violationFindings(
	file source.File,
	found scope,
	declaration ast.Decl,
) (findings []report.Finding) {
	for _, flagged := range violations(declaration) {
		findings = append(
			findings,
			report.Finding{
				Rule:     apiRule,
				Location: file.Location(flagged.node),
				Snippet: fmt.Sprintf(
					apiFormat,
					flagged.kind,
					flagged.name,
					found.label,
					apiRemedy(
						flagged,
						found,
					),
				),
			},
		)
	}

	return findings
}

func isEntryFunction(
	file source.File,
	function *ast.FuncDecl,
) (isEntry bool) {
	return filepath.Base(file.Path) == containerFile && function.Recv == nil && function.Name.IsExported()
}

func apiRemedy(
	flagged violation,
	found scope,
) (remedy string) {
	switch flagged.kind {
	case methodKind:
		return fmt.Sprintf(
			"implementation lives under %s; expose an interface here and construct the implementation in container.go",
			found.internal,
		)
	case funcKind:
		return functionRemedy(
			flagged,
			found,
		)
	case errorMethodKind:
		return fmt.Sprintf(
			"Error, Unwrap, Is, As and String hold a single return of a field, a literal, a concatenation or one fmt.Sprintf, fmt.Errorf or errors call; move the logic under %s",
			found.internal,
		)
	}

	return fmt.Sprintf(
		"the API holds only exported types, constants and Err sentinels; move it under %s",
		found.internal,
	)
}

func functionRemedy(
	flagged violation,
	found scope,
) (remedy string) {
	if flagged.exported {
		return fmt.Sprintf(
			"entry functions live in container.go and only delegate; move it there or move its logic under %s",
			found.internal,
		)
	}

	return fmt.Sprintf(
		"move it under %s",
		found.internal,
	)
}
