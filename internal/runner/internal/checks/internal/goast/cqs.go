package goast

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
		addPointerFieldNames(fn.Recv.List, targets)
	}
	if fn.Type.Params != nil {
		addPointerFieldNames(fn.Type.Params.List, targets)
	}
	return targets
}

// addPointerFieldNames adds every name declared by a pointer-typed field
// into targets.
func addPointerFieldNames(fields []*ast.Field, targets map[string]bool) {
	for _, f := range fields {
		if _, ok := f.Type.(*ast.StarExpr); ok {
			for _, n := range f.Names {
				targets[n.Name] = true
			}
		}
	}
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
			if finding, ok := cqsFindingForDecl(root, pf, decl); ok {
				findings = append(findings, finding)
			}
		}
	}
	return findings
}

// cqsFindingForDecl reports an R33 violation for decl, or false if decl
// isn't a qualifying function (returns only error, or has no mutable
// targets, or never actually mutates one).
func cqsFindingForDecl(root string, pf ParsedFile, decl ast.Decl) (report.Finding, bool) {
	fn, ok := decl.(*ast.FuncDecl)
	if !ok || fn.Body == nil || !returnsNonError(fn) {
		return report.Finding{}, false
	}
	targets := mutableTargets(fn)
	if len(targets) == 0 || !funcMutatesAnyTarget(fn, targets) {
		return report.Finding{}, false
	}
	pos := pf.FSet.Position(fn.Pos())
	return report.Finding{
		Rule:     "R33",
		Location: fmt.Sprintf("%s:%d", RelPath(root, pf.Path), pos.Line),
		Snippet:  fmt.Sprintf("func %s(...) returns a value and mutates receiver/pointer param", fn.Name.Name),
	}, true
}

// funcMutatesAnyTarget reports whether fn's body assigns into any of targets.
func funcMutatesAnyTarget(fn *ast.FuncDecl, targets map[string]bool) bool {
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
	return mutated
}
