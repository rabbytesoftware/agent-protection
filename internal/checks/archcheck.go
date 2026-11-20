package checks

import (
	"fmt"
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

var supportPackageNames = map[string]bool{
	"errors": true, "models": true, "dto": true, "types": true, "contracts": true,
}

func isSharedSupportPackage(dir string) bool {
	return supportPackageNames[filepath.Base(dir)]
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
		if isWiringFile(pf.Path) {
			continue // composition roots (container.go/wire.go/main.go) legitimately
			// import many siblings to wire them together -- not a layering
			// violation. Same allowlist concept as R37's DI-wiring check.
		}
		importerDir := filepath.Dir(pf.Path)
		if !strings.Contains(importerDir, string(filepath.Separator)+"internal"+string(filepath.Separator)) &&
			!strings.HasSuffix(importerDir, string(filepath.Separator)+"internal") {
			continue // only apply the sibling rule inside internal/ trees
		}
		for _, imp := range pf.AST.Imports {
			path, uerr := strconv.Unquote(imp.Path.Value)
			if uerr != nil || !strings.HasPrefix(path, modPath+"/internal/") {
				continue
			}
			importedRelDir := strings.TrimPrefix(path, modPath+"/")
			importedDir := filepath.Join(root, importedRelDir)
			importerRelDir := RelPath(root, importerDir)
			importedRelDirClean := filepath.Clean(importedRelDir)
			if importerRelDir == importedRelDirClean {
				continue // importing itself, not a sibling
			}
			if filepath.Dir(importerDir) == filepath.Dir(importedDir) && importerDir != importedDir {
				if isSharedSupportPackage(importedDir) {
					continue // a conventionally-shared leaf package (errors/models/dto/
					// types/contracts) is the intentional "extract to a shared lower
					// package" target, not a peer business package -- depending on it
					// from a sibling is the correct pattern, not a violation.
				}
				findings = append(findings, report.Finding{
					Rule:     "R16",
					Location: fmt.Sprintf("%s:%d", RelPath(root, pf.Path), pf.FSet.Position(imp.Pos()).Line),
					Snippet:  fmt.Sprintf("imports sibling package %s", path),
				})
			}
		}
	}
	return findings, nil
}
