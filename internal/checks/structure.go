package checks

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
)

var bannedPackageNames = map[string]bool{
	"util": true, "utils": true, "common": true, "helpers": true,
	"helper": true, "base": true, "misc": true, "shared": true,
}

// PackageHygiene flags R21 (package naming) and R22 (no util/common/etc).
func PackageHygiene(root string, files []ParsedFile) []report.Finding {
	var findings []report.Finding
	seen := map[string]bool{}
	for _, pf := range files {
		name := pf.AST.Name.Name
		if bannedPackageNames[name] {
			findings = append(findings, report.Finding{
				Rule:     "R22",
				Location: RelPath(root, pf.Path),
				Snippet:  fmt.Sprintf("package %q is a banned generic name", name),
			})
		}
		key := pf.Path + "|" + name
		if seen[key] {
			continue
		}
		seen[key] = true
		if strings.ContainsAny(name, "_") || strings.ToLower(name) != name {
			findings = append(findings, report.Finding{
				Rule:     "R21",
				Location: RelPath(root, pf.Path),
				Snippet:  fmt.Sprintf("package name %q must be short, lowercase, no underscores", name),
			})
		}
	}
	return findings
}

// NoStutter flags R23: an exported identifier that repeats its own package
// name, except the single allowed vault.Vault-style exact-match exception.
func NoStutter(root string, files []ParsedFile) []report.Finding {
	var findings []report.Finding
	for _, pf := range files {
		pkg := strings.ToLower(pf.AST.Name.Name)
		for _, decl := range pf.AST.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gen.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || !ts.Name.IsExported() {
					continue
				}
				lower := strings.ToLower(ts.Name.Name)
				if !strings.HasPrefix(lower, pkg) {
					continue
				}
				if lower == pkg { // exact match (vault.Vault) is the one allowed exception
					continue
				}
				findings = append(findings, report.Finding{
					Rule:     "R23",
					Location: fmt.Sprintf("%s:%d", RelPath(root, pf.Path), pf.FSet.Position(ts.Pos()).Line),
					Snippet:  fmt.Sprintf("%s.%s stutters the package name", pf.AST.Name.Name, ts.Name.Name),
				})
			}
		}
	}
	return findings
}

// OneTypePerFile flags R08: more than one top-level type declaration in a
// single file.
func OneTypePerFile(root string, files []ParsedFile) []report.Finding {
	var findings []report.Finding
	for _, pf := range files {
		count := 0
		for _, decl := range pf.AST.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gen.Specs {
				if _, ok := spec.(*ast.TypeSpec); ok {
					count++
				}
			}
		}
		if count > 1 {
			findings = append(findings, report.Finding{
				Rule:     "R08",
				Location: RelPath(root, pf.Path),
				Snippet:  fmt.Sprintf("%d type declarations in one file (max 1)", count),
			})
		}
	}
	return findings
}

// NoElse flags R03: any if-statement with an else branch (guard clauses /
// early return instead).
func NoElse(root string, files []ParsedFile) []report.Finding {
	var findings []report.Finding
	for _, pf := range files {
		ast.Inspect(pf.AST, func(n ast.Node) bool {
			ifs, ok := n.(*ast.IfStmt)
			if !ok || ifs.Else == nil {
				return true
			}
			pos := pf.FSet.Position(ifs.Else.Pos())
			findings = append(findings, report.Finding{
				Rule:     "R03",
				Location: fmt.Sprintf("%s:%d", RelPath(root, pf.Path), pos.Line),
				Snippet:  "else branch -- use an early return / guard clause instead",
			})
			return true
		})
	}
	return findings
}

// NakedReturn flags R04: a bare `return` in a function with named results.
func NakedReturn(root string, files []ParsedFile) []report.Finding {
	var findings []report.Finding
	for _, pf := range files {
		for _, decl := range pf.AST.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil || fn.Type.Results == nil {
				continue
			}
			named := false
			for _, r := range fn.Type.Results.List {
				if len(r.Names) > 0 {
					named = true
				}
			}
			if !named {
				continue
			}
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				ret, ok := n.(*ast.ReturnStmt)
				if ok && len(ret.Results) == 0 {
					pos := pf.FSet.Position(ret.Pos())
					findings = append(findings, report.Finding{
						Rule:     "R04",
						Location: fmt.Sprintf("%s:%d", RelPath(root, pf.Path), pos.Line),
						Snippet:  fmt.Sprintf("naked return in %s (has named results)", fn.Name.Name),
					})
				}
				return true
			})
		}
	}
	return findings
}

// CtxFirstParam flags R25: a context.Context parameter that isn't first, or
// a struct field typed context.Context (never store ctx in a struct).
func CtxFirstParam(root string, files []ParsedFile) []report.Finding {
	var findings []report.Finding
	isCtxType := func(e ast.Expr) bool {
		sel, ok := e.(*ast.SelectorExpr)
		if !ok {
			return false
		}
		pkg, ok := sel.X.(*ast.Ident)
		return ok && pkg.Name == "context" && sel.Sel.Name == "Context"
	}
	for _, pf := range files {
		for _, decl := range pf.AST.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Type.Params == nil {
				continue
			}
			for i, field := range fn.Type.Params.List {
				if isCtxType(field.Type) && i != 0 {
					pos := pf.FSet.Position(fn.Pos())
					findings = append(findings, report.Finding{
						Rule:     "R25",
						Location: fmt.Sprintf("%s:%d", RelPath(root, pf.Path), pos.Line),
						Snippet:  fmt.Sprintf("func %s: context.Context must be the first parameter", fn.Name.Name),
					})
				}
			}
		}
		ast.Inspect(pf.AST, func(n ast.Node) bool {
			ts, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				return true
			}
			for _, f := range st.Fields.List {
				if isCtxType(f.Type) {
					pos := pf.FSet.Position(f.Pos())
					findings = append(findings, report.Finding{
						Rule:     "R25",
						Location: fmt.Sprintf("%s:%d", RelPath(root, pf.Path), pos.Line),
						Snippet:  fmt.Sprintf("struct %s stores a context.Context field", ts.Name.Name),
					})
				}
			}
			return true
		})
	}
	return findings
}

// FileLineLimit flags R44/R45: a source file over 600 lines.
func FileLineLimit(root string, files []ParsedFile) []report.Finding {
	var findings []report.Finding
	for _, pf := range files {
		lastLine := pf.FSet.Position(pf.AST.End()).Line
		if lastLine > 600 {
			findings = append(findings, report.Finding{
				Rule:     "R44",
				Location: RelPath(root, pf.Path),
				Snippet:  fmt.Sprintf("%d lines (max 600)", lastLine),
			})
		}
	}
	return findings
}

// NoInfraInUnitTests flags the "no real infra in unit tests" rule: a
// _test.go file (without an //go:build integration tag) making an actual
// outbound network/db call -- importing net/http for its *types* (building
// an in-memory httptest.Request) is normal and is explicitly not the target.
var liveInfraCalls = map[string]bool{
	"Get": true, "Post": true, "Put": true, "Patch": true, "Do": true,
	"Dial": true, "DialContext": true, "Open": true, "OpenDB": true,
}

func NoInfraInUnitTests(root string, files []ParsedFile) []report.Finding {
	var findings []report.Finding
	err := walkTestFiles(root, func(pf ParsedFile) {
		for _, decl := range pf.AST.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			tainted := loopbackTaintedIdents(fn.Body)
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok || !liveInfraCalls[sel.Sel.Name] {
					return true
				}
				pkgIdent, ok := sel.X.(*ast.Ident)
				if !ok {
					return true
				}
				if pkgIdent.Name != "http" && pkgIdent.Name != "net" && pkgIdent.Name != "sql" {
					return true
				}
				if callArgsAreLoopback(call, tainted) {
					return true
				}
				pos := pf.FSet.Position(call.Pos())
				findings = append(findings, report.Finding{
					Rule:     "R41",
					Location: fmt.Sprintf("%s:%d", RelPath(root, pf.Path), pos.Line),
					Snippet:  fmt.Sprintf("unit test makes a live call: %s.%s(...)", pkgIdent.Name, sel.Sel.Name),
				})
				return true
			})
		}
	})
	_ = err
	return findings
}

// loopbackTaintedIdents does a simple same-function, single-assignment taint
// pass: any local variable whose initializer contains a loopback hint (an
// Addr()/URL field off a listener/server, or a literal loopback address) is
// tracked by name, so a later `net.Dial("tcp", addr)` using that variable is
// still recognized as loopback traffic, not external infra.
func loopbackTaintedIdents(body *ast.BlockStmt) map[string]bool {
	tainted := map[string]bool{}
	ast.Inspect(body, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for i, rhs := range assign.Rhs {
			if !containsLoopbackHint(rhs) || i >= len(assign.Lhs) {
				continue
			}
			if ident, ok := assign.Lhs[i].(*ast.Ident); ok {
				tainted[ident.Name] = true
			}
		}
		return true
	})
	return tainted
}

func callArgsAreLoopback(call *ast.CallExpr, tainted map[string]bool) bool {
	for _, arg := range call.Args {
		if containsLoopbackHint(arg) {
			return true
		}
		if ident, ok := arg.(*ast.Ident); ok && tainted[ident.Name] {
			return true
		}
	}
	return false
}

var loopbackFieldOrMethodNames = map[string]bool{
	"Addr": true, "URL": true, "String": true,
}

func containsLoopbackHint(e ast.Expr) bool {
	switch v := e.(type) {
	case *ast.BasicLit:
		return strings.Contains(v.Value, "127.0.0.1") ||
			strings.Contains(v.Value, "localhost") ||
			strings.Contains(v.Value, "unix")
	case *ast.BinaryExpr:
		return containsLoopbackHint(v.X) || containsLoopbackHint(v.Y)
	case *ast.CallExpr:
		for _, a := range v.Args {
			if containsLoopbackHint(a) {
				return true
			}
		}
		if sel, ok := v.Fun.(*ast.SelectorExpr); ok {
			return loopbackFieldOrMethodNames[sel.Sel.Name]
		}
	case *ast.SelectorExpr:
		return loopbackFieldOrMethodNames[v.Sel.Name]
	}
	return false
}
