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

const generatedHeaderScanBytes = 500

type ParsedFile struct {
	Path string
	FSet *token.FileSet
	AST  *ast.File
}

// walkMatchingFiles walks root, skipping vendor/.git trees, and calls visit
// for every file whose path satisfies matches. Shared traversal/skip logic
// for WalkGoFiles and walkTestFiles.
func walkMatchingFiles(root string, matches func(path string) bool, visit func(path string) error) error {
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
		if !matches(path) {
			return nil
		}
		return visit(path)
	})
}

// parseGoFile parses path (reading it itself if src is nil) and reports
// whether parsing succeeded.
//
// A nil src can't be forwarded directly to parser.ParseFile: passed through
// a []byte-typed parameter, it becomes a non-nil `any` (type []byte, value
// nil), so go/parser's "src != nil" check reads it as empty source instead
// of reading the file -- an empty file then fails to parse (no package
// clause). Caught by a real test that found WalkGoFiles returning nothing.
func parseGoFile(path string, src []byte) (ParsedFile, bool) {
	fset := token.NewFileSet()
	f, perr := parseFileSource(fset, path, src)
	if perr != nil {
		return ParsedFile{}, false
	}
	return ParsedFile{Path: path, FSet: fset, AST: f}, true
}

// parseFileSource reads from src, or from path itself if src is nil. A nil
// src can't be forwarded directly (see parseGoFile's doc comment above for
// why), so the nil case is a distinct early return instead of a shared
// variable assignment.
func parseFileSource(fset *token.FileSet, path string, src []byte) (*ast.File, error) {
	if src == nil {
		return parser.ParseFile(fset, path, nil, parser.ParseComments)
	}
	return parser.ParseFile(fset, path, src, parser.ParseComments)
}

// isCandidateGoSourceFile reports whether path is a non-test, non-generated
// Go source file -- the scope WalkGoFiles parses.
func isCandidateGoSourceFile(path string) bool {
	return strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") && !isGeneratedFile(path)
}

// WalkGoFiles parses every non-test, non-vendor, non-generated .go file
// under root.
func WalkGoFiles(root string) ([]ParsedFile, error) {
	var out []ParsedFile
	err := walkMatchingFiles(root, isCandidateGoSourceFile, func(path string) error {
		if pf, ok := parseGoFile(path, nil); ok {
			out = append(out, pf)
		}
		return nil
	})
	return out, err
}

// isGeneratedFile does a lenient, substring-based check (not the strict
// "// Code generated ... DO NOT EDIT." regex) because real generator tools
// don't always emit the exact canonical marker -- e.g. a real-world
// swaggo-generated docs.go omits the trailing period, which the strict
// marker (and golangci-lint's default "strict" setting) would miss entirely.
func isGeneratedFile(path string) bool {
	b, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	head := string(b)
	if len(head) > generatedHeaderScanBytes {
		head = head[:generatedHeaderScanBytes]
	}
	return strings.Contains(head, "Code generated") || strings.Contains(head, "DO NOT EDIT")
}

func RelPath(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return rel
}

// isCandidateUnitTestFile reports whether path is a _test.go file that
// should be scanned as a unit test: not generated, and not tagged
// `//go:build integration`.
func isCandidateUnitTestFile(path string) bool {
	if !strings.HasSuffix(path, "_test.go") || isGeneratedFile(path) {
		return false
	}
	return !hasIntegrationBuildTag(path)
}

// hasIntegrationBuildTag reports whether path declares a `//go:build
// integration` tag.
func hasIntegrationBuildTag(path string) bool {
	b, rerr := os.ReadFile(path)
	if rerr != nil {
		return false
	}
	return strings.Contains(string(b), "integration") && strings.Contains(string(b), "go:build")
}

// walkTestFiles parses every _test.go file under root that doesn't declare
// an `integration` build tag, and calls fn on each.
func walkTestFiles(root string, fn func(ParsedFile)) error {
	return walkMatchingFiles(root, isCandidateUnitTestFile, func(path string) error {
		b, rerr := os.ReadFile(path)
		if rerr != nil {
			return nil
		}
		if pf, ok := parseGoFile(path, b); ok {
			fn(pf)
		}
		return nil
	})
}

const maxFilesPerDir = 8

// isNonTestGoFile reports whether path is any .go file except a _test.go
// one -- deliberately NOT generated-file-aware like isCandidateGoSourceFile,
// since TestPairing and MaxFilesPerDir intentionally still count generated
// files toward test-pairing and per-directory file limits.
func isNonTestGoFile(path string) bool {
	return strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go")
}

// TestPairing flags R42: every non-test .go file must have a matching
// _test.go file beside it. Declaration-only files (pure interfaces/consts,
// no function bodies -- nothing to assert behaviorally) are exempt, per the
// ruling this project settled on early: there's no behavior to test.
func TestPairing(root string) ([]report.Finding, error) {
	var findings []report.Finding
	err := walkMatchingFiles(root, isNonTestGoFile, func(path string) error {
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
	err := walkMatchingFiles(root, isNonTestGoFile, func(path string) error {
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
