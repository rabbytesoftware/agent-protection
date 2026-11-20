package goast

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
)

func pureSoundingPrefixes() []string {
	return []string{"Get", "Find", "Compute", "List", "Calc", "Is", "Has"}
}

func writeVerbSuffixes() []string {
	return []string{
		"Write", "WriteFile", "Save", "Delete", "Remove", "RemoveAll", "Create",
		"Update", "Insert", "Exec", "ExecContext", "Post", "Put", "Patch", "Send",
		"Set", "MkdirAll", "Store",
	}
}

func hasPureSoundingName(name string) bool {
	for _, p := range pureSoundingPrefixes() {
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
	for _, suf := range writeVerbSuffixes() {
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
			if finding, ok := hiddenSideEffectFinding(root, pf, decl); ok {
				findings = append(findings, finding)
			}
		}
	}
	return findings
}

// hiddenSideEffectFinding reports an R34 violation for decl, or false if
// decl isn't a pure-sounding function, or its body never calls a
// write-looking method.
func hiddenSideEffectFinding(root string, pf ParsedFile, decl ast.Decl) (report.Finding, bool) {
	fn, ok := decl.(*ast.FuncDecl)
	if !ok || fn.Body == nil || !hasPureSoundingName(fn.Name.Name) {
		return report.Finding{}, false
	}
	hit := findWriteCall(fn.Body)
	if hit == nil {
		return report.Finding{}, false
	}
	pos := pf.FSet.Position(fn.Pos())
	sel := hit.Fun.(*ast.SelectorExpr)
	return report.Finding{
		Rule:     "R34",
		Location: fmt.Sprintf("%s:%d", RelPath(root, pf.Path), pos.Line),
		Snippet:  fmt.Sprintf("func %s(...) calls .%s() despite its pure-sounding name", fn.Name.Name, sel.Sel.Name),
	}, true
}

// findWriteCall returns the first write-looking call expression inside body,
// or nil if none exists.
func findWriteCall(body *ast.BlockStmt) *ast.CallExpr {
	var hit *ast.CallExpr
	ast.Inspect(body, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok && looksLikeWrite(call) {
			hit = call
		}
		return true
	})
	return hit
}
