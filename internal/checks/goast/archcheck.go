package goast

import (
	"fmt"
	"go/ast"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
)

func readModulePath(root string) (string, error) {
	b, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return "", fmt.Errorf("read go.mod: %w", err)
	}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module")), nil
		}
	}
	return "", fmt.Errorf("no module directive found in go.mod")
}

func supportPackageNames() map[string]bool {
	return map[string]bool{
		"errors": true, "models": true, "dto": true, "types": true, "contracts": true,
	}
}

func isSharedSupportPackage(dir string) bool {
	return supportPackageNames()[filepath.Base(dir)]
}

func wiringFileHints() []string {
	return []string{"container.go", "wire.go", "main.go", "new.go"}
}

func isWiringFile(path string) bool {
	base := filepath.Base(path)
	for _, hint := range wiringFileHints() {
		if base == hint || strings.HasSuffix(base, "_"+hint) || strings.Contains(base, "container") {
			return true
		}
	}
	return false
}

// SiblingImports flags R16: a package importing a sibling package under the
// same internal/ parent directory.
func SiblingImports(root string, files []ParsedFile) ([]report.Finding, error) {
	modPath, err := readModulePath(root)
	if err != nil {
		return nil, err
	}
	var findings []report.Finding
	for _, pf := range files {
		findings = append(findings, siblingImportsInFile(root, modPath, pf)...)
	}
	return findings, nil
}

// siblingImportsInFile checks a single parsed file for R16 violations.
// Composition roots and files outside internal/ are skipped entirely --
// same allowlist concept as R37's (removed) DI-wiring check.
func siblingImportsInFile(root, modPath string, pf ParsedFile) []report.Finding {
	if isWiringFile(pf.Path) {
		return nil
	}
	importerDir := filepath.Dir(pf.Path)
	if !isInsideInternalTree(importerDir) {
		return nil
	}
	var findings []report.Finding
	for _, imp := range pf.AST.Imports {
		if finding, ok := siblingImportFinding(root, modPath, importerDir, pf, imp); ok {
			findings = append(findings, finding)
		}
	}
	return findings
}

// isInsideInternalTree reports whether dir sits under an internal/ parent,
// the only scope R16 applies to.
func isInsideInternalTree(dir string) bool {
	sep := string(filepath.Separator)
	return strings.Contains(dir, sep+"internal"+sep) || strings.HasSuffix(dir, sep+"internal")
}

// siblingImportFinding reports an R16 violation for a single import, or
// false if the import is self-referential, not a sibling, or targets a
// conventionally-shared leaf package (errors/models/dto/types/contracts --
// the intentional "extract to a shared lower package" target, not a peer
// business package).
func siblingImportFinding(root, modPath, importerDir string, pf ParsedFile, imp *ast.ImportSpec) (report.Finding, bool) {
	path, uerr := strconv.Unquote(imp.Path.Value)
	if uerr != nil || !strings.HasPrefix(path, modPath+"/internal/") {
		return report.Finding{}, false
	}
	importedRelDir := strings.TrimPrefix(path, modPath+"/")
	importedDir := filepath.Join(root, importedRelDir)
	if RelPath(root, importerDir) == filepath.Clean(importedRelDir) {
		return report.Finding{}, false
	}
	if !isSiblingPair(importerDir, importedDir) {
		return report.Finding{}, false
	}
	if isSharedSupportPackage(importedDir) {
		return report.Finding{}, false
	}
	return report.Finding{
		Rule:     "R16",
		Location: fmt.Sprintf("%s:%d", RelPath(root, pf.Path), pf.FSet.Position(imp.Pos()).Line),
		Snippet:  fmt.Sprintf("imports sibling package %s", path),
	}, true
}

// isSiblingPair reports whether a and b are distinct directories sharing
// the same parent.
func isSiblingPair(a, b string) bool {
	return filepath.Dir(a) == filepath.Dir(b) && a != b
}
