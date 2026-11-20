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

const maxFilesPerDir = 8

// walkNonTestGoFiles walks root, skipping vendor/.git trees, and calls visit
// for every non-test .go file found. Shared by TestPairing and
// MaxFilesPerDir -- both need the exact same traversal and skip rules.
func walkNonTestGoFiles(root string, visit func(path string) error) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
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
		return visit(path)
	})
}

// TestPairing flags R42: every non-test .go file must have a matching
// _test.go file beside it. Declaration-only files (pure interfaces/consts,
// no function bodies -- nothing to assert behaviorally) are exempt, per the
// ruling this project settled on early: there's no behavior to test.
func TestPairing(root string) ([]report.Finding, error) {
	var findings []report.Finding
	err := walkNonTestGoFiles(root, func(path string) error {
		if finding, ok := testPairingFinding(root, path); ok {
			findings = append(findings, finding)
		}
		return nil
	})
	return findings, err
}

// testPairingFinding reports an R42 violation for path, or false if path is
// declaration-only or already has a matching _test.go file.
func testPairingFinding(root, path string) (report.Finding, bool) {
	if isDeclarationOnly(path) {
		return report.Finding{}, false
	}
	testPath := strings.TrimSuffix(path, ".go") + "_test.go"
	if _, statErr := os.Stat(testPath); statErr == nil {
		return report.Finding{}, false
	}
	return report.Finding{
		Rule:     "R42",
		Location: RelPath(root, path),
		Snippet:  "no matching _test.go file",
	}, true
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
	err := walkNonTestGoFiles(root, func(path string) error {
		dir := filepath.Dir(path)
		if dir != root { // exclude project root per rule text
			counts[dir]++
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return filesPerDirFindings(root, counts), nil
}

// filesPerDirFindings turns a per-directory file-count tally into R46
// findings for every directory over budget.
func filesPerDirFindings(root string, counts map[string]int) []report.Finding {
	var findings []report.Finding
	for dir, n := range counts {
		if n > maxFilesPerDir {
			findings = append(findings, report.Finding{
				Rule:     "R46",
				Location: RelPath(root, dir),
				Snippet:  fmt.Sprintf("%d non-test source files (max %d)", n, maxFilesPerDir),
			})
		}
	}
	return findings
}
