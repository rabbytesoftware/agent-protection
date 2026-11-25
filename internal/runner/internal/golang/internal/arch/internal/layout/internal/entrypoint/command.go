package entrypoint

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	mainFunction  = "main"
	funcKind      = "func"
	methodKind    = "method"
	typeKind      = "type"
	varKind       = "var"
	commandFormat = "file %s in package main directory %s: only main.go holds code in a command; move the implementation under %s and call its entry function from func main"
	strayFormat   = "%s %s in %s: main.go holds only func main, which only delegates to an entry function; move it under %s"
)

func mainFindings(
	found unit,
) (findings []report.Finding) {
	for _, declaration := range found.file.Syntax.Decls {
		findings = append(
			findings,
			mainDeclarationFindings(
				found,
				declaration,
			)...,
		)
	}

	return findings
}

func mainDeclarationFindings(
	found unit,
	declaration ast.Decl,
) (findings []report.Finding) {
	switch typed := declaration.(type) {
	case *ast.FuncDecl:
		return mainFunctionFindings(
			found,
			typed,
		)
	case *ast.GenDecl:
		return slices.Concat(
			typeFindings(
				found,
				typed,
			),
			variableFindings(
				found,
				typed,
			),
		)
	}

	return nil
}

func mainFunctionFindings(
	found unit,
	function *ast.FuncDecl,
) (findings []report.Finding) {
	if function.Recv != nil {
		return strayFindings(
			found,
			methodKind,
			function.Name,
		)
	}

	if function.Name.Name != mainFunction {
		return strayFindings(
			found,
			funcKind,
			function.Name,
		)
	}

	return delegationFindings(
		found,
		function,
	)
}

func typeFindings(
	found unit,
	general *ast.GenDecl,
) (findings []report.Finding) {
	if general.Tok != token.TYPE {
		return nil
	}

	for _, spec := range source.Nodes[*ast.TypeSpec](general) {
		findings = append(
			findings,
			strayFindings(
				found,
				typeKind,
				spec.Name,
			)...,
		)
	}

	return findings
}

func variableFindings(
	found unit,
	general *ast.GenDecl,
) (findings []report.Finding) {
	if general.Tok != token.VAR {
		return nil
	}

	for _, spec := range source.Nodes[*ast.ValueSpec](general) {
		findings = append(
			findings,
			runningFindings(
				found,
				spec,
			)...,
		)
	}

	return findings
}

func runningFindings(
	found unit,
	spec *ast.ValueSpec,
) (findings []report.Finding) {
	running := slices.ContainsFunc(
		spec.Values,
		runsCode,
	)
	if !running {
		return nil
	}

	for _, name := range spec.Names {
		findings = append(
			findings,
			strayFindings(
				found,
				varKind,
				name,
			)...,
		)
	}

	return findings
}

func runsCode(
	value ast.Expr,
) (runs bool) {
	return len(source.Nodes[*ast.CallExpr](value))+len(source.Nodes[*ast.FuncLit](value)) > 0
}

func strayFindings(
	found unit,
	kind string,
	name *ast.Ident,
) (findings []report.Finding) {
	return []report.Finding{{
		Rule:     apiRule,
		Location: found.file.Location(name),
		Snippet: fmt.Sprintf(
			strayFormat,
			kind,
			name.Name,
			found.file.Rel,
			internalPath(found.directory),
		),
	}}
}

func commandFindings(
	found unit,
) (findings []report.Finding) {
	return []report.Finding{{
		Rule:     apiRule,
		Location: found.file.Location(found.file.Syntax.Name),
		Snippet: fmt.Sprintf(
			commandFormat,
			filepath.Base(found.file.Rel),
			found.directory,
			internalPath(found.directory),
		),
	}}
}
