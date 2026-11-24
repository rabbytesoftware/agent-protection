package nopanic

import (
	"go/ast"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule        = "R27"
	builtin     = "panic"
	mainPackage = "main"
)

// Nopanic reports every call to the builtin panic outside package main:
// library code hands failures back to its caller as errors and leaves the
// decision to crash to the program's entry point.
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
	calls := slices.DeleteFunc(
		source.Nodes[*ast.CallExpr](file.Syntax),
		isNotPanic,
	)
	for _, call := range calls {
		findings = append(
			findings,
			report.Finding{
				Rule:     rule,
				Location: file.Location(call),
				Snippet:  "panic call: return an error to the caller instead",
			},
		)
	}
	return findings
}

func isNotPanic(
	call *ast.CallExpr,
) (notPanic bool) {
	callee, isIdent := call.Fun.(*ast.Ident)
	return !isIdent || callee.Name != builtin
}
