package goast

import (
	"fmt"
	"go/ast"
	"os"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
)

const maxFileLines = 600

func bannedPackageNames() map[string]bool {
	return map[string]bool{
		"util": true, "utils": true, "common": true, "helpers": true,
		"helper": true, "base": true, "misc": true, "shared": true,
	}
}

// PackageHygiene flags R21 (package naming) and R22 (no util/common/etc).
func PackageHygiene(root string, files []ParsedFile) []report.Finding {
	var findings []report.Finding
	seen := map[string]bool{}
	for _, pf := range files {
		findings = append(findings, packageHygieneFindings(root, pf, seen)...)
	}
	return findings
}

// packageHygieneFindings reports R22 (banned generic name) and, once per
// (file, package name) pair, R21 (naming format).
func packageHygieneFindings(root string, pf ParsedFile, seen map[string]bool) []report.Finding {
	var findings []report.Finding
	name := pf.AST.Name.Name
	if bannedPackageNames()[name] {
		findings = append(findings, report.Finding{
			Rule:     "R22",
			Location: RelPath(root, pf.Path),
			Snippet:  fmt.Sprintf("package %q is a banned generic name", name),
		})
	}
	key := pf.Path + "|" + name
	if seen[key] {
		return findings
	}
	seen[key] = true
	if strings.ContainsAny(name, "_") || strings.ToLower(name) != name {
		findings = append(findings, report.Finding{
			Rule:     "R21",
			Location: RelPath(root, pf.Path),
			Snippet:  fmt.Sprintf("package name %q must be short, lowercase, no underscores", name),
		})
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
			findings = append(findings, stutterFindingsInDecl(root, pf, pkg, decl)...)
		}
	}
	return findings
}

// stutterFindingsInDecl reports R23 for every exported type in decl whose
// name stutters the package name.
func stutterFindingsInDecl(root string, pf ParsedFile, pkg string, decl ast.Decl) []report.Finding {
	gen, ok := decl.(*ast.GenDecl)
	if !ok {
		return nil
	}
	var findings []report.Finding
	for _, spec := range gen.Specs {
		ts, ok := spec.(*ast.TypeSpec)
		if !ok || !ts.Name.IsExported() || !stuttersPackageName(ts.Name.Name, pkg) {
			continue
		}
		findings = append(findings, report.Finding{
			Rule:     "R23",
			Location: fmt.Sprintf("%s:%d", RelPath(root, pf.Path), pf.FSet.Position(ts.Pos()).Line),
			Snippet:  fmt.Sprintf("%s.%s stutters the package name", pf.AST.Name.Name, ts.Name.Name),
		})
	}
	return findings
}

// stuttersPackageName reports whether typeName repeats pkg, excluding the
// single allowed exact-match exception (vault.Vault-style).
func stuttersPackageName(typeName, pkg string) bool {
	lower := strings.ToLower(typeName)
	return strings.HasPrefix(lower, pkg) && lower != pkg
}

// OneTypePerFile flags R08: more than one top-level type declaration in a
// single file.
func OneTypePerFile(root string, files []ParsedFile) []report.Finding {
	var findings []report.Finding
	for _, pf := range files {
		if count := countTypeDecls(pf.AST.Decls); count > 1 {
			findings = append(findings, report.Finding{
				Rule:     "R08",
				Location: RelPath(root, pf.Path),
				Snippet:  fmt.Sprintf("%d type declarations in one file (max 1)", count),
			})
		}
	}
	return findings
}

// countTypeDecls counts top-level type declarations across decls.
func countTypeDecls(decls []ast.Decl) int {
	count := 0
	for _, decl := range decls {
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
	return count
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
			findings = append(findings, nakedReturnFindings(root, pf, decl)...)
		}
	}
	return findings
}

// nakedReturnFindings reports R04 for every bare `return` inside decl, if
// decl is a function with named results.
func nakedReturnFindings(root string, pf ParsedFile, decl ast.Decl) []report.Finding {
	fn, ok := decl.(*ast.FuncDecl)
	if !ok || fn.Body == nil || fn.Type.Results == nil || !hasNamedResults(fn) {
		return nil
	}
	var findings []report.Finding
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		ret, ok := n.(*ast.ReturnStmt)
		if !ok || len(ret.Results) != 0 {
			return true
		}
		pos := pf.FSet.Position(ret.Pos())
		findings = append(findings, report.Finding{
			Rule:     "R04",
			Location: fmt.Sprintf("%s:%d", RelPath(root, pf.Path), pos.Line),
			Snippet:  fmt.Sprintf("naked return in %s (has named results)", fn.Name.Name),
		})
		return true
	})
	return findings
}

// hasNamedResults reports whether fn declares at least one named result.
func hasNamedResults(fn *ast.FuncDecl) bool {
	for _, r := range fn.Type.Results.List {
		if len(r.Names) > 0 {
			return true
		}
	}
	return false
}

// isCtxType reports whether e is exactly context.Context.
func isCtxType(e ast.Expr) bool {
	sel, ok := e.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "context" && sel.Sel.Name == "Context"
}

// CtxFirstParam flags R25: a context.Context parameter that isn't first, or
// a struct field typed context.Context (never store ctx in a struct).
func CtxFirstParam(root string, files []ParsedFile) []report.Finding {
	var findings []report.Finding
	for _, pf := range files {
		for _, decl := range pf.AST.Decls {
			findings = append(findings, ctxParamOrderFindings(root, pf, decl)...)
		}
		findings = append(findings, ctxStructFieldFindings(root, pf)...)
	}
	return findings
}

// ctxParamOrderFindings reports R25 for every context.Context parameter in
// decl that isn't the function's first parameter.
func ctxParamOrderFindings(root string, pf ParsedFile, decl ast.Decl) []report.Finding {
	fn, ok := decl.(*ast.FuncDecl)
	if !ok || fn.Type.Params == nil {
		return nil
	}
	var findings []report.Finding
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
	return findings
}

// ctxStructFieldFindings reports R25 for every struct field in pf typed
// context.Context.
func ctxStructFieldFindings(root string, pf ParsedFile) []report.Finding {
	var findings []report.Finding
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
			if !isCtxType(f.Type) {
				continue
			}
			pos := pf.FSet.Position(f.Pos())
			findings = append(findings, report.Finding{
				Rule:     "R25",
				Location: fmt.Sprintf("%s:%d", RelPath(root, pf.Path), pos.Line),
				Snippet:  fmt.Sprintf("struct %s stores a context.Context field", ts.Name.Name),
			})
		}
		return true
	})
	return findings
}

// FileLineLimit flags R44/R45: a source file over 600 lines.
func FileLineLimit(root string, files []ParsedFile) []report.Finding {
	var findings []report.Finding
	for _, pf := range files {
		// Count raw source lines, not pf.AST.End()'s position -- AST end is
		// the last DECLARATION's position, so a file ending in trailing
		// comments with no following declaration under-counts its real
		// length. Caught by a real test with a comment-only trailing body.
		b, err := os.ReadFile(pf.Path)
		if err != nil {
			continue
		}
		lineCount := strings.Count(string(b), "\n") + 1
		if lineCount > maxFileLines {
			findings = append(findings, report.Finding{
				Rule:     "R44",
				Location: RelPath(root, pf.Path),
				Snippet:  fmt.Sprintf("%d lines (max %d)", lineCount, maxFileLines),
			})
		}
	}
	return findings
}

// liveInfraCalls lists the method names that, when called against the
// http/net/sql packages, perform a real outbound call.
func liveInfraCalls() map[string]bool {
	return map[string]bool{
		"Get": true, "Post": true, "Put": true, "Patch": true, "Do": true,
		"Dial": true, "DialContext": true, "Open": true, "OpenDB": true,
	}
}

// NoInfraInUnitTests flags the "no real infra in unit tests" rule: a
// _test.go file (without an //go:build integration tag) making an actual
// outbound network/db call -- importing net/http for its *types* (building
// an in-memory httptest.Request) is normal and is explicitly not the target.
func NoInfraInUnitTests(root string, files []ParsedFile) []report.Finding {
	var findings []report.Finding
	_ = walkTestFiles(root, func(pf ParsedFile) {
		for _, decl := range pf.AST.Decls {
			findings = append(findings, liveInfraFindingsInDecl(root, pf, decl)...)
		}
	})
	return findings
}

// liveInfraFindingsInDecl reports R41 for every live-infra call inside decl
// that isn't itself targeting loopback traffic.
func liveInfraFindingsInDecl(root string, pf ParsedFile, decl ast.Decl) []report.Finding {
	fn, ok := decl.(*ast.FuncDecl)
	if !ok || fn.Body == nil {
		return nil
	}
	tainted := loopbackTaintedIdents(fn.Body)
	var findings []report.Finding
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if finding, hit := liveInfraCallFinding(root, pf, n, tainted); hit {
			findings = append(findings, finding)
		}
		return true
	})
	return findings
}

// liveInfraCallFinding reports R41 for n if it's a live-infra call targeting
// a non-loopback destination.
func liveInfraCallFinding(root string, pf ParsedFile, n ast.Node, tainted map[string]bool) (report.Finding, bool) {
	call, ok := n.(*ast.CallExpr)
	if !ok {
		return report.Finding{}, false
	}
	pkgIdent, methodName, ok := matchingLiveInfraCall(call)
	if !ok || callArgsAreLoopback(call, tainted) {
		return report.Finding{}, false
	}
	pos := pf.FSet.Position(call.Pos())
	return report.Finding{
		Rule:     "R41",
		Location: fmt.Sprintf("%s:%d", RelPath(root, pf.Path), pos.Line),
		Snippet:  fmt.Sprintf("unit test makes a live call: %s.%s(...)", pkgIdent.Name, methodName),
	}, true
}

// matchingLiveInfraCall reports the package identifier and method name if
// call selects a live-infra method off the http/net/sql packages.
func matchingLiveInfraCall(call *ast.CallExpr) (*ast.Ident, string, bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || !liveInfraCalls()[sel.Sel.Name] {
		return nil, "", false
	}
	pkgIdent, ok := sel.X.(*ast.Ident)
	if !ok || !isLiveInfraPackage(pkgIdent.Name) {
		return nil, "", false
	}
	return pkgIdent, sel.Sel.Name, true
}

// isLiveInfraPackage reports whether name is one of the packages this check
// treats as real outbound infrastructure.
func isLiveInfraPackage(name string) bool {
	return name == "http" || name == "net" || name == "sql"
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

func loopbackFieldOrMethodNames() map[string]bool {
	return map[string]bool{
		"Addr": true, "URL": true, "String": true,
	}
}

// containsLoopbackHint reports whether e is, or calls/selects something
// that looks like, a loopback address (127.0.0.1/localhost/unix socket, or
// an Addr()/URL/String() accessor off a listener/server).
func containsLoopbackHint(e ast.Expr) bool {
	switch v := e.(type) {
	case *ast.BasicLit:
		return literalLooksLoopback(v)
	case *ast.BinaryExpr:
		return containsLoopbackHint(v.X) || containsLoopbackHint(v.Y)
	case *ast.CallExpr:
		return callLooksLoopback(v)
	case *ast.SelectorExpr:
		return loopbackFieldOrMethodNames()[v.Sel.Name]
	}
	return false
}

// literalLooksLoopback reports whether a basic literal's text contains a
// recognizable loopback address fragment.
func literalLooksLoopback(v *ast.BasicLit) bool {
	return strings.Contains(v.Value, "127.0.0.1") ||
		strings.Contains(v.Value, "localhost") ||
		strings.Contains(v.Value, "unix")
}

// callLooksLoopback reports whether a call expression either passes a
// loopback-looking argument, or selects a loopback-looking method/field.
func callLooksLoopback(v *ast.CallExpr) bool {
	for _, a := range v.Args {
		if containsLoopbackHint(a) {
			return true
		}
	}
	sel, ok := v.Fun.(*ast.SelectorExpr)
	return ok && loopbackFieldOrMethodNames()[sel.Sel.Name]
}
