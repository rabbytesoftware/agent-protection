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

// isWiringFile reports whether path is a conventionally-named composition
// root: container.go/wire.go/main.go/new.go, OR the file named after its
// own package (e.g. config.go inside package config) -- the other half of
// rule 10's "the package's main free function lives in the file named after
// the package" convention.
func isWiringFile(path, packageName string) bool {
	base := filepath.Base(path)
	for _, hint := range wiringFileHints() {
		if base == hint || strings.HasSuffix(base, "_"+hint) || strings.Contains(base, "container") {
			return true
		}
	}
	return base == packageName+".go"
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

// wiringThreshold is the number of distinct sibling packages a file must
// import before it's treated as a de facto composition root (R47) rather
// than an ordinary sibling-import violation (R16) per import.
const wiringThreshold = 2

// siblingImportsInFile checks a single parsed file for R16/R47 violations.
// Conventionally-named composition roots are skipped entirely -- same
// allowlist concept as R37's (removed) DI-wiring check. A file that is NOT
// conventionally named but imports enough siblings to be acting as one
// anyway is flagged once as R47 instead of once per import as R16: the
// problem isn't N separate layering violations, it's one unlabeled
// composition root.
func siblingImportsInFile(root, modPath string, pf ParsedFile) []report.Finding {
	if isWiringFile(pf.Path, pf.AST.Name.Name) {
		return nil
	}
	importerDir := filepath.Dir(pf.Path)
	if !isInsideInternalTree(importerDir) {
		return nil
	}
	var sibFindings []report.Finding
	for _, imp := range pf.AST.Imports {
		if finding, ok := siblingImportFinding(root, modPath, importerDir, pf, imp); ok {
			sibFindings = append(sibFindings, finding)
		}
	}
	if len(sibFindings) >= wiringThreshold {
		return []report.Finding{unconventionalWiringFinding(root, pf, sibFindings)}
	}
	return sibFindings
}

// unconventionalWiringFinding reports R47 for a file that imports enough
// siblings to be acting as a composition root, but isn't named like one.
func unconventionalWiringFinding(root string, pf ParsedFile, sibFindings []report.Finding) report.Finding {
	return report.Finding{
		Rule:     "R47",
		Location: RelPath(root, pf.Path),
		Snippet:  fmt.Sprintf("imports %d sibling packages (acts as a composition root) but isn't named container.go/wire.go/<package>.go", len(sibFindings)),
	}
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
