package checks

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
)

var pureSoundingPrefixes = []string{"Get", "Find", "Compute", "List", "Calc", "Is", "Has"}

var writeVerbSuffixes = []string{
	"Write", "WriteFile", "Save", "Delete", "Remove", "RemoveAll", "Create",
	"Update", "Insert", "Exec", "ExecContext", "Post", "Put", "Patch", "Send",
	"Set", "MkdirAll", "Store",
}

func hasPureSoundingName(name string) bool {
	for _, p := range pureSoundingPrefixes {
		if strings.HasPrefix(name, p) {
			return true
		}
	}
	return false
}

func looksLikeWrite(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	for _, suf := range writeVerbSuffixes {
		if sel.Sel.Name == suf {
			return true
		}
	}
	return false
}

// HiddenSideEffect flags R34: a function named like a pure query (Get/Find/
// Compute/List/Calc/Is/Has) whose body performs an I/O write or similar
// mutating call.
func HiddenSideEffect(root string, files []ParsedFile) []report.Finding {
	var findings []report.Finding
	for _, pf := range files {
		for _, decl := range pf.AST.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil || !hasPureSoundingName(fn.Name.Name) {
				continue
			}
			var hit *ast.CallExpr
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if ok && looksLikeWrite(call) {
					hit = call
				}
				return true
			})
			if hit != nil {
				pos := pf.FSet.Position(fn.Pos())
				sel := hit.Fun.(*ast.SelectorExpr)
				findings = append(findings, report.Finding{
					Rule:     "R34",
					Location: fmt.Sprintf("%s:%d", RelPath(root, pf.Path), pos.Line),
					Snippet:  fmt.Sprintf("func %s(...) calls .%s() despite its pure-sounding name", fn.Name.Name, sel.Sel.Name),
				})
			}
		}
	}
	return findings
}
