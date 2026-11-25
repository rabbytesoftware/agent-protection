package nopanic

import (
	"go/ast"
	"path"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule            = "R27"
	builtin         = "panic"
	mainPackage     = "main"
	panicSnippet    = "panic call: return an error to the caller instead"
	exitSnippet     = " call ends the process from a library: return an error to the caller instead and let main decide whether to exit"
	qualifier       = "."
	logPath         = "log"
	osPath          = "os"
	runtimePath     = "runtime"
	unresolvedName  = ""
	blankImportName = "_"
	dotImportName   = "."
	pathQuotes      = "\"`"
)

func Nopanic(
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
	if file.Syntax.Name.Name == mainPackage {
		return nil
	}

	imports := importedPaths(file.Syntax)
	for _, call := range source.Nodes[*ast.CallExpr](file.Syntax) {
		findings = appendFinding(
			findings,
			file,
			call,
			imports,
		)
	}

	return findings
}

func appendFinding(
	findings []report.Finding,
	file source.File,
	call *ast.CallExpr,
	imports map[string]string,
) (extended []report.Finding) {
	snippet := callSnippet(
		call,
		imports,
	)
	if snippet == unresolvedName {
		return findings
	}

	return append(
		findings,
		report.Finding{
			Rule:     rule,
			Location: file.Location(call),
			Snippet:  snippet,
		},
	)
}

func callSnippet(
	call *ast.CallExpr,
	imports map[string]string,
) (snippet string) {
	callee, isIdent := call.Fun.(*ast.Ident)
	isPanic := isIdent &&
		callee.Name == builtin
	if isPanic {
		return panicSnippet
	}

	name := terminatorName(
		call,
		imports,
	)
	if name == unresolvedName {
		return unresolvedName
	}

	return name + exitSnippet
}

func terminatorName(
	call *ast.CallExpr,
	imports map[string]string,
) (name string) {
	selector, isSelector := call.Fun.(*ast.SelectorExpr)
	if !isSelector {
		return unresolvedName
	}

	receiver, isIdent := selector.X.(*ast.Ident)
	isPackage := isIdent &&
		receiver.Obj == nil
	if !isPackage {
		return unresolvedName
	}

	importPath := imports[receiver.Name]
	isTerminator := slices.Contains(
		terminators()[importPath],
		selector.Sel.Name,
	)
	if !isTerminator {
		return unresolvedName
	}

	return importPath + qualifier + selector.Sel.Name
}

func importedPaths(
	syntax *ast.File,
) (imports map[string]string) {
	imports = map[string]string{}
	for _, spec := range syntax.Imports {
		importPath := strings.Trim(
			spec.Path.Value,
			pathQuotes,
		)
		imports[importName(
			spec,
			importPath,
		)] = importPath
	}

	delete(
		imports,
		blankImportName,
	)
	delete(
		imports,
		dotImportName,
	)
	return imports
}

func importName(
	spec *ast.ImportSpec,
	importPath string,
) (name string) {
	if spec.Name != nil {
		return spec.Name.Name
	}

	return path.Base(importPath)
}

func terminators() (names map[string][]string) {
	return map[string][]string{
		logPath: {
			"Fatal",
			"Fatalf",
			"Fatalln",
			"Panic",
			"Panicf",
			"Panicln",
		},
		osPath:      {"Exit"},
		runtimePath: {"Goexit"},
	}
}
