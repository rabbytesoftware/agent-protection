package nopanic

import (
	"go/ast"
	"go/types"
	"path"
	"path/filepath"
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
	valueSnippet    = " used as a value ends the process from a library when called: return an error to the caller instead and let main decide whether to exit"
	qualifier       = "."
	logPath         = "log"
	osPath          = "os"
	runtimePath     = "runtime"
	syscallPath     = "syscall"
	unixPath        = "golang.org/x/sys/unix"
	windowsPath     = "golang.org/x/sys/windows"
	unresolvedName  = ""
	blankImportName = "_"
	dotImportName   = "."
	pathQuotes      = "\"`"
)

func Nopanic(
	files []source.File,
) (findings []report.Finding) {
	fields := loggerFields(files)
	for _, file := range files {
		findings = append(
			findings,
			fileFindings(
				file,
				fields[filepath.Dir(file.Path)],
			)...,
		)
	}

	return findings
}

func fileFindings(
	file source.File,
	fields []string,
) (findings []report.Finding) {
	if file.Syntax.Name.Name == mainPackage {
		return nil
	}

	found := scopeOf(
		file.Syntax,
		fields,
	)
	for _, expression := range source.Nodes[ast.Expr](file.Syntax) {
		findings = appendFinding(
			findings,
			file,
			expression,
			found,
		)
	}

	return findings
}

func appendFinding(
	findings []report.Finding,
	file source.File,
	expression ast.Expr,
	found scope,
) (extended []report.Finding) {
	snippet := expressionSnippet(
		expression,
		found,
	)
	if snippet == unresolvedName {
		return findings
	}

	return append(
		findings,
		report.Finding{
			Rule:     rule,
			Location: file.Location(expression),
			Snippet:  snippet,
		},
	)
}

func expressionSnippet(
	expression ast.Expr,
	found scope,
) (snippet string) {
	switch typed := expression.(type) {
	case *ast.CallExpr:
		return panicCallSnippet(typed)
	case *ast.SelectorExpr:
		return selectorSnippet(
			typed,
			found,
		)
	}

	return unresolvedName
}

func panicCallSnippet(
	call *ast.CallExpr,
) (snippet string) {
	callee, isIdent := call.Fun.(*ast.Ident)
	isPanic := isIdent &&
		callee.Name == builtin
	if isPanic {
		return panicSnippet
	}

	return unresolvedName
}

func selectorSnippet(
	selector *ast.SelectorExpr,
	found scope,
) (snippet string) {
	name := terminatorName(
		selector,
		found.imports,
	)
	isLoggerExit := name == unresolvedName && slices.Contains(
		terminators()[logPath],
		selector.Sel.Name,
	) && found.isLogger(selector.X)
	if isLoggerExit {
		name = types.ExprString(selector)
	}

	if name == unresolvedName {
		return unresolvedName
	}

	if found.isCalled(selector) {
		return name + exitSnippet
	}

	return name + valueSnippet
}

func terminatorName(
	selector *ast.SelectorExpr,
	imports map[string]string,
) (name string) {
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
		syscallPath: {"Exit"},
		unixPath:    {"Exit"},
		windowsPath: {"Exit"},
	}
}
