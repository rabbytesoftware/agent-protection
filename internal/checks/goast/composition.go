package goast

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
)

// CompositionRoot flags R48 and R13: a directory with 2+ sibling internal/
// child packages must itself be a package, in a file named after itself,
// declaring at least one exported interface -- never an exported struct.
func CompositionRoot(root string) ([]report.Finding, error) {
	candidates, err := directoriesNeedingContainers(root)
	if err != nil {
		return nil, err
	}
	var findings []report.Finding
	for _, dir := range candidates {
		findings = append(findings, containerFindings(root, dir)...)
	}
	return findings, nil
}

// directoriesNeedingContainers returns every directory inside an internal/
// tree (excluding "internal" itself, which isn't a composable package) that
// has 2+ immediate child directories directly containing Go source --
// i.e. is the parent of 2+ sibling internal packages.
func directoriesNeedingContainers(root string) ([]string, error) {
	allDirs, err := listDirs(root)
	if err != nil {
		return nil, err
	}
	childCount := packageChildCounts(allDirs)

	var candidates []string
	for _, dir := range allDirs {
		if qualifiesAsContainerCandidate(dir, root, childCount[dir]) {
			candidates = append(candidates, dir)
		}
	}
	sort.Strings(candidates)
	return candidates, nil
}

// listDirs returns every directory under root, skipping vendor/.git trees.
func listDirs(root string) ([]string, error) {
	var allDirs []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, werr error) error {
		if werr != nil {
			return werr
		}
		if !d.IsDir() {
			return nil
		}
		if d.Name() == "vendor" || d.Name() == ".git" {
			return filepath.SkipDir
		}
		allDirs = append(allDirs, path)
		return nil
	})
	return allDirs, err
}

// packageChildCounts tallies, per directory, how many of its immediate
// children directly contain Go source -- i.e. are themselves packages.
func packageChildCounts(dirs []string) map[string]int {
	counts := map[string]int{}
	for _, dir := range dirs {
		if dirHasGoSource(dir) {
			counts[filepath.Dir(dir)]++
		}
	}
	return counts
}

// qualifiesAsContainerCandidate reports whether dir is the parent of 2+
// sibling internal packages, excluding the module root and the literal
// "internal" directory (which isn't itself a composable package).
func qualifiesAsContainerCandidate(dir, root string, childPkgCount int) bool {
	if dir == root || filepath.Base(dir) == "internal" {
		return false
	}
	if !isInsideInternalTree(dir) {
		return false
	}
	return childPkgCount >= wiringThreshold
}

// dirHasGoSource reports whether dir directly contains at least one
// non-test .go file (i.e. is itself a Go package, not just a namespacing
// folder of subpackages).
func dirHasGoSource(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".go") && !strings.HasSuffix(e.Name(), "_test.go") {
			return true
		}
	}
	return false
}

// containerFindings checks dir's composition-root file: it must exist (any
// conventionally-named one -- container.go/wire.go/main.go/new.go, or
// named after dir's own package, same convention R47 already recognizes),
// declare at least one exported interface, must not export a concrete
// struct type, and must be the ONLY loose non-test .go file directly in
// dir -- any other implementation belongs in a nested internal/ subtree.
func containerFindings(root, dir string) []report.Finding {
	containerPath := findContainerFile(dir)
	if containerPath == "" {
		return []report.Finding{{
			Rule:     "R48",
			Location: RelPath(root, dir),
			Snippet:  fmt.Sprintf("directory has 2+ sibling internal packages but no composition root (container.go, wire.go, or %s)", filepath.Base(dir)+".go"),
		}}
	}
	findings := interfaceExposureFindings(root, containerPath)
	findings = append(findings, strayFileFindings(root, dir, containerPath)...)
	return findings
}

// strayFileFindings reports R49 for every non-test .go file directly in
// dir besides containerPath itself -- a composition-root directory may
// contain only its own container file (plus test), never loose peers.
func strayFileFindings(root, dir, containerPath string) []report.Finding {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var findings []report.Finding
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		if path == containerPath {
			continue
		}
		findings = append(findings, report.Finding{
			Rule:     "R49",
			Location: RelPath(root, path),
			Snippet:  fmt.Sprintf("loose file beside composition root %s -- move into a nested internal/ subdirectory", filepath.Base(containerPath)),
		})
	}
	return findings
}

// findContainerFile returns the path of dir's composition-root file, if
// any exists, using the exact same naming convention as isWiringFile.
func findContainerFile(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		if isWiringFile(path, filepath.Base(dir)) {
			return path
		}
	}
	return ""
}

// interfaceExposureFindings reports R48 if containerPath declares zero
// exported interfaces, and R13 for every exported struct type it declares
// instead.
func interfaceExposureFindings(root, containerPath string) []report.Finding {
	fset := token.NewFileSet()
	f, perr := parser.ParseFile(fset, containerPath, nil, 0)
	if perr != nil {
		return nil // unparsable -- other checks will surface this
	}
	var findings []report.Finding
	interfaceCount := 0
	for _, ts := range exportedTypeSpecs(f) {
		if _, isInterface := ts.Type.(*ast.InterfaceType); isInterface {
			interfaceCount++
			continue
		}
		if finding, ok := exportedStructFinding(root, containerPath, fset, ts); ok {
			findings = append(findings, finding)
		}
	}
	if interfaceCount == 0 {
		findings = append(findings, report.Finding{
			Rule:     "R48",
			Location: RelPath(root, containerPath),
			Snippet:  "composition root exists but declares no exported interface",
		})
	}
	return findings
}

// exportedTypeSpecs returns every exported top-level type declared in f.
func exportedTypeSpecs(f *ast.File) []*ast.TypeSpec {
	var specs []*ast.TypeSpec
	for _, decl := range f.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range gen.Specs {
			if ts, ok := spec.(*ast.TypeSpec); ok && ts.Name.IsExported() {
				specs = append(specs, ts)
			}
		}
	}
	return specs
}

// exportedStructFinding reports R13 if ts is an exported struct type.
func exportedStructFinding(root, containerPath string, fset *token.FileSet, ts *ast.TypeSpec) (report.Finding, bool) {
	if _, isStruct := ts.Type.(*ast.StructType); !isStruct {
		return report.Finding{}, false
	}
	return report.Finding{
		Rule:     "R13",
		Location: fmt.Sprintf("%s:%d", RelPath(root, containerPath), fset.Position(ts.Pos()).Line),
		Snippet:  fmt.Sprintf("composition root exports struct %s -- expose an interface instead", ts.Name.Name),
	}, true
}
