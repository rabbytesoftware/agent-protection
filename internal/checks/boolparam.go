package checks

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
			fn, ok := n.(*ast.FuncDecl)
			if !ok || fn.Type.Params == nil {
				return true
			}
			for _, field := range fn.Type.Params.List {
				ident, ok := field.Type.(*ast.Ident)
				if !ok || ident.Name != "bool" {
					continue
				}
				pos := pf.FSet.Position(fn.Pos())
				findings = append(findings, report.Finding{
					Rule:     "R32",
					Location: fmt.Sprintf("%s:%d", RelPath(root, pf.Path), pos.Line),
					Snippet:  fmt.Sprintf("func %s(...) has a bool parameter", fn.Name.Name),
				})
				break
			}
			return true
		})
	}
	return findings
}
