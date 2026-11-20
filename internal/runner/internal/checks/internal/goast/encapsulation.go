package goast

import (
	"fmt"
	"go/ast"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
)

// SingleConsumerExposure flags R19: a package reachable from the entire
// module (its path contains exactly one internal/ segment) but imported by
// exactly one other package elsewhere in the tree. It should be nested
// under a second internal/ boundary scoped to that one real consumer, so
// Go's own visibility rule -- not just convention -- restricts it.
func SingleConsumerExposure(root string, files []ParsedFile) ([]report.Finding, error) {
	modPath, err := readModulePath(root)
	if err != nil {
		return nil, err
	}
	importedBy := importerDirsByPackageDir(root, modPath, files)

	var candidates []string
	for dir := range importedBy {
		// A shared leaf (errors/models/rules/report/...) is deliberately
		// meant to be broadly reachable data/contract -- that is the exact
		// opposite of an accidentally-exposed implementation detail.
		if isExposedToWholeModule(root, dir) && !isSharedSupportPackage(dir) {
			candidates = append(candidates, dir)
		}
	}
	sort.Strings(candidates)

	mainImporters := mainPackageDirs(files)

	var findings []report.Finding
	for _, dir := range candidates {
		if len(importedBy[dir]) != 1 {
			continue
		}
		sole := oneOf(importedBy[dir])
		// Go already forbids importing package main from anywhere, so a
		// package whose only importer is a cmd/ entrypoint carries no risk
		// of a stray import from elsewhere in the module -- nesting it
		// under cmd/'s own internal/ buys nothing real.
		if mainImporters[sole] {
			continue
		}
		findings = append(findings, report.Finding{
			Rule:     "R19",
			Location: RelPath(root, dir),
			Snippet:  fmt.Sprintf("imported only by %s but reachable from the whole module -- nest under %s's own internal/ subtree", RelPath(root, sole), RelPath(root, sole)),
		})
	}
	return findings, nil
}

// mainPackageDirs returns the set of directories whose Go files declare
// package main.
func mainPackageDirs(files []ParsedFile) map[string]bool {
	dirs := map[string]bool{}
	for _, pf := range files {
		if pf.AST.Name.Name == "main" {
			dirs[filepath.Dir(pf.Path)] = true
		}
	}
	return dirs
}

// importerDirsByPackageDir maps each module-internal package directory to
// the set of distinct directories that import it.
func importerDirsByPackageDir(root, modPath string, files []ParsedFile) map[string]map[string]bool {
	importedBy := map[string]map[string]bool{}
	for _, pf := range files {
		importerDir := filepath.Dir(pf.Path)
		for _, imp := range pf.AST.Imports {
			importedDir, ok := resolveInternalImport(root, modPath, imp)
			if !ok || importedDir == importerDir {
				continue
			}
			if importedBy[importedDir] == nil {
				importedBy[importedDir] = map[string]bool{}
			}
			importedBy[importedDir][importerDir] = true
		}
	}
	return importedBy
}

// resolveInternalImport returns the directory a module-internal import
// path resolves to, or false if the import isn't internal to this module.
func resolveInternalImport(root, modPath string, imp *ast.ImportSpec) (string, bool) {
	path, uerr := strconv.Unquote(imp.Path.Value)
	if uerr != nil || !strings.HasPrefix(path, modPath+"/") {
		return "", false
	}
	relDir := strings.TrimPrefix(path, modPath+"/")
	return filepath.Join(root, relDir), true
}

// isExposedToWholeModule reports whether dir's path contains exactly one
// internal/ segment -- i.e. Go's own visibility rule restricts it only to
// "anywhere in this module", not to any one real consumer.
func isExposedToWholeModule(root, dir string) bool {
	rel := RelPath(root, dir)
	count := 0
	for _, seg := range strings.Split(filepath.ToSlash(rel), "/") {
		if seg == "internal" {
			count++
		}
	}
	return count == 1
}

// oneOf returns an arbitrary element of a non-empty set, for a single-entry
// map where "arbitrary" means "the only one."
func oneOf(set map[string]bool) string {
	for k := range set {
		return k
	}
	return ""
}
