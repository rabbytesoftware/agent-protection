package compositionroot

import (
	"fmt"
	"go/ast"
	"go/token"
	"maps"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	missingRule = "R48"
	structRule  = "R13"
	looseRule   = "R49"
)

// Compositionroot checks every component directory: a directory inside an
// internal/ tree with at least two child packages, counting those under its
// own internal/. Its composition-root file must expose an exported interface
// or a function named after the package, export no struct, and stand alone.
func Compositionroot(
	root string,
	files []source.File,
) (findings []report.Finding, err error) {
	directories, err := source.Directories(root)
	if err != nil {
		return nil,
			fmt.Errorf(
				"composition roots: %w",
				err,
			)
	}
	paths, err := source.Paths(root)
	if err != nil {
		return nil,
			fmt.Errorf(
				"composition roots: %w",
				err,
			)
	}
	groups := groupByDirectory(paths)
	parsed := indexByPath(files)
	for _, directory := range components(
		root,
		directories,
		slices.Collect(maps.Keys(groups)),
	) {
		findings = append(
			findings,
			componentFindings(
				root,
				directory,
				groups[directory],
				parsed,
			)...,
		)
	}
	return findings,
		nil
}

func groupByDirectory(
	paths []string,
) (groups map[string][]string) {
	groups = map[string][]string{}
	for _, path := range paths {
		directory := filepath.Dir(path)
		groups[directory] = append(
			groups[directory],
			path,
		)
	}
	return groups
}

func indexByPath(
	files []source.File,
) (index map[string]source.File) {
	index = map[string]source.File{}
	for _, file := range files {
		index[file.Path] = file
	}
	return index
}

func components(
	root string,
	directories []string,
	sourceDirectories []string,
) (found []string) {
	for _, directory := range directories {
		found = append(
			found,
			asComponent(
				root,
				directory,
				sourceDirectories,
			)...,
		)
	}
	return found
}

func asComponent(
	root string,
	directory string,
	sourceDirectories []string,
) (component []string) {
	isInsideBoundary := source.IsInternalTree(source.RelPath(
		root,
		directory,
	)) && !source.IsInternalDirectory(directory)
	isComponent := isInsideBoundary && len(childPackages(
		directory,
		sourceDirectories,
	)) >= source.CompositionRootThreshold
	if !isComponent {
		return nil
	}
	return []string{directory}
}

func childPackages(
	directory string,
	sourceDirectories []string,
) (children []string) {
	for _, sourceDirectory := range sourceDirectories {
		children = append(
			children,
			childPackage(
				directory,
				sourceDirectory,
			)...,
		)
	}
	slices.Sort(children)
	return slices.Compact(children)
}

func childPackage(
	directory string,
	sourceDirectory string,
) (child []string) {
	rel, isBeneath := strings.CutPrefix(
		sourceDirectory,
		directory+string(filepath.Separator),
	)
	if !isBeneath {
		return nil
	}
	segments := strings.Split(
		filepath.ToSlash(rel),
		"/",
	)
	if !source.IsInternalDirectory(segments[0]) {
		return segments[:1]
	}
	if len(segments) == 1 {
		return nil
	}
	return []string{strings.Join(
		segments[:2],
		"/",
	)}
}

func componentFindings(
	root string,
	directory string,
	entries []string,
	parsed map[string]source.File,
) (findings []report.Finding) {
	index := slices.IndexFunc(
		entries,
		isCompositionRoot,
	)
	if index < 0 {
		return []report.Finding{missingRootFinding(
			root,
			directory,
		)}
	}
	strays := slices.Delete(
		slices.Clone(entries),
		index,
		index+1,
	)
	return slices.Concat(
		contractFindings(parsed[entries[index]]),
		strayFindings(
			root,
			strays,
		),
	)
}

func missingRootFinding(
	root string,
	directory string,
) (finding report.Finding) {
	return report.Finding{
		Rule: missingRule,
		Location: source.RelPath(
			root,
			directory,
		),
		Snippet: fmt.Sprintf(
			"component directory has no composition root: add %s.go",
			filepath.Base(directory),
		),
	}
}

func contractFindings(
	file source.File,
) (findings []report.Finding) {
	if file.Syntax == nil {
		return nil
	}
	types := exportedTypes(file.Syntax)
	findings = structFindings(
		file,
		types,
	)
	isExposed := slices.ContainsFunc(
		types,
		isInterface,
	) || hasEntryFunction(file.Syntax)
	if isExposed {
		return findings
	}
	return append(
		findings,
		report.Finding{
			Rule:     missingRule,
			Location: file.Location(file.Syntax.Name),
			Snippet: fmt.Sprintf(
				"composition root exposes no exported interface and no func %s",
				entryName(file.Syntax.Name.Name),
			),
		},
	)
}

func structFindings(
	file source.File,
	types []*ast.TypeSpec,
) (findings []report.Finding) {
	for _, spec := range slices.DeleteFunc(
		slices.Clone(types),
		isNotStruct,
	) {
		findings = append(
			findings,
			report.Finding{
				Rule:     structRule,
				Location: file.Location(spec),
				Snippet: fmt.Sprintf(
					"composition root exports struct %s: expose an interface or an entry function instead",
					spec.Name.Name,
				),
			},
		)
	}
	return findings
}

func strayFindings(
	root string,
	strays []string,
) (findings []report.Finding) {
	for _, stray := range strays {
		findings = append(
			findings,
			report.Finding{
				Rule: looseRule,
				Location: source.RelPath(
					root,
					stray,
				),
				Snippet: "file beside the composition root: move it into its own package under internal/",
			},
		)
	}
	return findings
}

func exportedTypes(
	syntax *ast.File,
) (types []*ast.TypeSpec) {
	for _, declaration := range syntax.Decls {
		types = append(
			types,
			typeSpecs(declaration)...,
		)
	}
	return slices.DeleteFunc(
		types,
		isUnexported,
	)
}

func typeSpecs(
	declaration ast.Decl,
) (specs []*ast.TypeSpec) {
	general, isGeneral := declaration.(*ast.GenDecl)
	isTypeDeclaration := isGeneral && general.Tok == token.TYPE
	if !isTypeDeclaration {
		return nil
	}
	return source.Nodes[*ast.TypeSpec](general)
}

func hasEntryFunction(
	syntax *ast.File,
) (hasEntry bool) {
	names := []string{}
	for _, function := range slices.DeleteFunc(
		source.Nodes[*ast.FuncDecl](syntax),
		isNotExportedFunction,
	) {
		names = append(
			names,
			strings.ToLower(function.Name.Name),
		)
	}
	return slices.Contains(
		names,
		strings.ToLower(syntax.Name.Name),
	)
}

func entryName(
	packageName string,
) (name string) {
	return strings.ToUpper(packageName[:1]) + packageName[1:]
}

func isCompositionRoot(
	path string,
) (isRoot bool) {
	return source.IsCompositionRoot(
		path,
		filepath.Base(filepath.Dir(path)),
	)
}

func isUnexported(
	spec *ast.TypeSpec,
) (unexported bool) {
	return !spec.Name.IsExported()
}

func isInterface(
	spec *ast.TypeSpec,
) (matches bool) {
	_, matches = spec.Type.(*ast.InterfaceType)
	return matches
}

func isNotStruct(
	spec *ast.TypeSpec,
) (notStruct bool) {
	_, isStruct := spec.Type.(*ast.StructType)
	return !isStruct
}

func isNotExportedFunction(
	function *ast.FuncDecl,
) (excluded bool) {
	return function.Recv != nil || !function.Name.IsExported()
}
