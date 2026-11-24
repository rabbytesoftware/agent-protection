package source

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const (
	internalDirectory = "internal"
	moduleDirective   = "module "
)

// CompositionRootThreshold is the number of distinct sibling packages a
// file must import before it acts as a composition root.
const CompositionRootThreshold = 2

// ModulePath reads the module path declared by root's go.mod.
func ModulePath(
	root string,
) (modulePath string, err error) {
	content, err := os.ReadFile(filepath.Join(
		root,
		"go.mod",
	))
	if err != nil {
		return "",
			fmt.Errorf(
				"read go.mod: %w",
				err,
			)
	}

	lines := strings.Split(
		string(content),
		"\n",
	)
	index := slices.IndexFunc(
		lines,
		isModuleDirective,
	)
	if index < 0 {
		return "",
			fmt.Errorf(
				"no module directive in %s",
				filepath.Join(
					root,
					"go.mod",
				),
			)
	}

	return strings.TrimSpace(strings.TrimPrefix(
			strings.TrimSpace(lines[index]),
			moduleDirective,
		)),
		nil
}

// IsDeclarationOnly reports whether the Go file at path declares no
// functions or methods, meaning it holds only types, constants and
// variables and has no behavior of its own.
func IsDeclarationOnly(
	path string,
) (declarationOnly bool) {
	syntax, err := parser.ParseFile(
		token.NewFileSet(),
		path,
		nil,
		parser.SkipObjectResolution,
	)
	if err != nil {
		return false
	}

	return !slices.ContainsFunc(
		syntax.Decls,
		isFuncDecl,
	)
}

// IsInternalTree reports whether dir is an internal/ directory or sits
// anywhere beneath one.
func IsInternalTree(
	dir string,
) (inside bool) {
	return slices.Contains(
		strings.Split(
			filepath.ToSlash(dir),
			"/",
		),
		internalDirectory,
	)
}

// IsInternalDirectory reports whether dir is itself named internal.
func IsInternalDirectory(
	dir string,
) (isInternal bool) {
	return filepath.Base(dir) == internalDirectory
}

// IsSharedSupport reports whether dir is a shared vocabulary package
// (errors, models, dto, types, contracts, rules, report) that any package
// may import regardless of layering.
func IsSharedSupport(
	dir string,
) (shared bool) {
	return slices.Contains(
		sharedSupportNames(),
		filepath.Base(dir),
	)
}

// IsMock reports whether rel, a path relative to the checked root, is a
// test double: its file name or one of its directories mentions mock or
// fake. Mocks exist only to stand in for real code in tests, so they have
// no tests of their own and keep the shape the interface they fake needs.
func IsMock(
	rel string,
) (isMock bool) {
	return slices.ContainsFunc(
		strings.Split(
			filepath.ToSlash(rel),
			"/",
		),
		mentionsMock,
	)
}

// IsCompositionRoot reports whether the file at path is conventionally
// named as a composition root: container.go, wire.go, main.go, new.go, or
// the file named after its own package.
func IsCompositionRoot(
	path string,
	packageName string,
) (isRoot bool) {
	base := filepath.Base(path)
	return slices.Contains(
		compositionRootNames(),
		base,
	) || base == packageName+goExtension
}

// HasSource reports whether dir directly contains at least one non-test
// .go file.
func HasSource(
	dir string,
) (hasSource bool) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	return slices.ContainsFunc(
		entries,
		isSourceEntry,
	)
}

// Entries returns the paths of the non-test .go files directly
// inside dir.
func Entries(
	dir string,
) (paths []string, err error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil,
			fmt.Errorf(
				"read %s: %w",
				dir,
				err,
			)
	}

	for _, entry := range slices.DeleteFunc(
		entries,
		isNotSourceEntry,
	) {
		paths = append(
			paths,
			filepath.Join(
				dir,
				entry.Name(),
			),
		)
	}

	return paths,
		nil
}

func isModuleDirective(
	line string,
) (matches bool) {
	return strings.HasPrefix(
		strings.TrimSpace(line),
		moduleDirective,
	)
}

func isFuncDecl(
	decl ast.Decl,
) (isFunc bool) {
	_, isFunc = decl.(*ast.FuncDecl)
	return isFunc
}

func isSourceEntry(
	entry os.DirEntry,
) (isSource bool) {
	return !entry.IsDir() && filepath.Ext(entry.Name()) == goExtension && !IsTest(entry.Name())
}

func isNotSourceEntry(
	entry os.DirEntry,
) (excluded bool) {
	return !isSourceEntry(entry)
}

func mentionsMock(
	segment string,
) (mentions bool) {
	lower := strings.ToLower(segment)
	return strings.Contains(
		lower,
		"mock",
	) || strings.Contains(
		lower,
		"fake",
	)
}

func sharedSupportNames() (names []string) {
	return []string{"errors", "models", "dto", "types", "contracts", "rules", "report"}
}

func compositionRootNames() (names []string) {
	return []string{"container.go", "wire.go", "main.go", "new.go"}
}
