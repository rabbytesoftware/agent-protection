package checks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
)

// TestPairing flags R42: every non-test .go file must have a matching
// _test.go file beside it.
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
