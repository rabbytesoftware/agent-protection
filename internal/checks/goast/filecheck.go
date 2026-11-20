package goast

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
)

// TestPairing flags R42: every non-test .go file must have a matching
// _test.go file beside it. Declaration-only files (pure interfaces/consts,
// no function bodies -- nothing to assert behaviorally) are exempt, per the
// ruling this project settled on early: there's no behavior to test.
func TestPairing(root string) ([]report.Finding, error) {
	var findings []report.Finding
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == "vendor" || info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		if isDeclarationOnly(path) {
			return nil
		}
		testPath := strings.TrimSuffix(path, ".go") + "_test.go"
		if _, statErr := os.Stat(testPath); statErr != nil {
			findings = append(findings, report.Finding{
				Rule:     "R42",
				Location: RelPath(root, path),
				Snippet:  "no matching _test.go file",
			})
		}
		return nil
	})
	return findings, err
}

// isDeclarationOnly reports whether a Go file has zero function
// declarations -- i.e. it's pure types/interfaces/consts with no behavior.
func isDeclarationOnly(path string) bool {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return false // unparsable -- don't exempt, let other checks surface it
	}
	for _, decl := range f.Decls {
		if _, ok := decl.(*ast.FuncDecl); ok {
			return false
		}
	}
	return true
}

// MaxFilesPerDir flags R46: a directory (excluding root) holds more than 8
// non-test source files.
func MaxFilesPerDir(root string) ([]report.Finding, error) {
	counts := map[string]int{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == "vendor" || info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		dir := filepath.Dir(path)
		if dir == root {
			return nil // exclude project root per rule text
		}
		counts[dir]++
		return nil
	})
	if err != nil {
		return nil, err
	}
	var findings []report.Finding
	for dir, n := range counts {
		if n > 8 {
			findings = append(findings, report.Finding{
				Rule:     "R46",
				Location: RelPath(root, dir),
				Snippet:  fmt.Sprintf("%d non-test source files (max 8)", n),
			})
		}
	}
	return findings, nil
}
