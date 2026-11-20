package goast

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type ParsedFile struct {
	Path string
	FSet *token.FileSet
	AST  *ast.File
}

// WalkGoFiles parses every non-test, non-vendor .go file under root.
func WalkGoFiles(root string) ([]ParsedFile, error) {
	var out []ParsedFile
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
		if isGeneratedFile(path) {
			return nil
		}
		fset := token.NewFileSet()
		f, perr := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if perr != nil {
			return nil // skip unparsable files, don't fail the whole run
		}
		out = append(out, ParsedFile{Path: path, FSet: fset, AST: f})
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
	if len(head) > 500 {
		head = head[:500]
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

// walkTestFiles parses every _test.go file under root that doesn't declare
// an `integration` build tag, and calls fn on each.
func walkTestFiles(root string, fn func(ParsedFile)) error {
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
		if !strings.HasSuffix(path, "_test.go") || isGeneratedFile(path) {
			return nil
		}
		b, rerr := os.ReadFile(path)
		if rerr != nil {
			return nil
		}
		if strings.Contains(string(b), "integration") && strings.Contains(string(b), "go:build") {
			return nil // skip integration-tagged test files
		}
		fset := token.NewFileSet()
		f, perr := parser.ParseFile(fset, path, b, parser.ParseComments)
		if perr != nil {
			return nil
		}
		fn(ParsedFile{Path: path, FSet: fset, AST: f})
		return nil
	})
}
