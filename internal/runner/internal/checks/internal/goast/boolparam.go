package goast

import (
	"fmt"
	"go/ast"

	"github.com/rabbytesoftware/agent-protection/internal/report"
)

// BoolParam flags R32: any function/method signature containing a bool-typed
// parameter.
func BoolParam(root string, files []ParsedFile) []report.Finding {
	var findings []report.Finding
	for _, pf := range files {
		ast.Inspect(pf.AST, func(n ast.Node) bool {
			if finding, ok := boolParamFinding(root, pf, n); ok {
				findings = append(findings, finding)
			}
			return true
		})
	}
	return findings
}

// boolParamFinding reports an R32 violation if n is a function declaration
// with at least one bool-typed parameter.
func boolParamFinding(root string, pf ParsedFile, n ast.Node) (report.Finding, bool) {
	fn, ok := n.(*ast.FuncDecl)
	if !ok || fn.Type.Params == nil || !hasBoolParam(fn) {
		return report.Finding{}, false
	}
	pos := pf.FSet.Position(fn.Pos())
	return report.Finding{
		Rule:     "R32",
		Location: fmt.Sprintf("%s:%d", RelPath(root, pf.Path), pos.Line),
		Snippet:  fmt.Sprintf("func %s(...) has a bool parameter", fn.Name.Name),
	}, true
}

// hasBoolParam reports whether fn declares at least one bool-typed parameter.
func hasBoolParam(fn *ast.FuncDecl) bool {
	for _, field := range fn.Type.Params.List {
		if ident, ok := field.Type.(*ast.Ident); ok && ident.Name == "bool" {
			return true
		}
	}
	return false
}
