package checks

import (
	"fmt"
	"go/ast"

	"github.com/rabbytesoftware/agent-protection/internal/report"
)

func returnsNonError(fn *ast.FuncDecl) bool {
	if fn.Type.Results == nil {
		return false
	}
	for _, r := range fn.Type.Results.List {
		if ident, ok := r.Type.(*ast.Ident); ok && ident.Name == "error" {
			continue
		}
		return true
	}
	return false
}

// mutableTargets returns the set of identifier names this function could
// mutate through: the receiver, and any pointer-typed parameter.
func mutableTargets(fn *ast.FuncDecl) map[string]bool {
	targets := map[string]bool{}
	if fn.Recv != nil {
		for _, f := range fn.Recv.List {
			if _, ok := f.Type.(*ast.StarExpr); ok {
				for _, n := range f.Names {
					targets[n.Name] = true
				}
			}
		}
	}
	if fn.Type.Params != nil {
		for _, f := range fn.Type.Params.List {
			if _, ok := f.Type.(*ast.StarExpr); ok {
				for _, n := range f.Names {
					targets[n.Name] = true
				}
			}
		}
	}
	return targets
}

func mutatesTarget(lhs ast.Expr, targets map[string]bool) bool {
	switch e := lhs.(type) {
	case *ast.SelectorExpr:
		if ident, ok := e.X.(*ast.Ident); ok {
			return targets[ident.Name]
		}
	case *ast.StarExpr:
		if ident, ok := e.X.(*ast.Ident); ok {
			return targets[ident.Name]
		}
	case *ast.IndexExpr:
		if ident, ok := e.X.(*ast.Ident); ok {
			return targets[ident.Name]
		}
	}
	return false
}

// CQSMismatch flags R33: a function that returns a non-error value AND
// mutates its receiver or a pointer parameter in the same body.
func CQSMismatch(root string, files []ParsedFile) []report.Finding {
	var findings []report.Finding
	for _, pf := range files {
		for _, decl := range pf.AST.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil || !returnsNonError(fn) {
				continue
			}
			targets := mutableTargets(fn)
			if len(targets) == 0 {
				continue
			}
			mutated := false
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				assign, ok := n.(*ast.AssignStmt)
				if !ok {
					return true
				}
				for _, lhs := range assign.Lhs {
					if mutatesTarget(lhs, targets) {
						mutated = true
					}
				}
				return true
			})
			if mutated {
				pos := pf.FSet.Position(fn.Pos())
				findings = append(findings, report.Finding{
					Rule:     "R33",
					Location: fmt.Sprintf("%s:%d", RelPath(root, pf.Path), pos.Line),
					Snippet:  fmt.Sprintf("func %s(...) returns a value and mutates receiver/pointer param", fn.Name.Name),
				})
			}
		}
	}
	return findings
}
