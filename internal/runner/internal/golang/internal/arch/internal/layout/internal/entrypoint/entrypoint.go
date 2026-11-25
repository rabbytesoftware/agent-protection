package entrypoint

import (
	"fmt"
	"go/ast"
	"path/filepath"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	apiRule       = "R14"
	resultRule    = "R64"
	containerFile = "container.go"
	mainFile      = "main.go"
	mainPackage   = "main"
	rootDirectory = "."
	internalChild = "internal/"
)

func Entrypoint(
	root string,
	files []source.File,
) (findings []report.Finding, err error) {
	modulePath, err := source.ModulePath(root)
	if err != nil {
		return nil,
			fmt.Errorf(
				"entry points: %w",
				err,
			)
	}

	paths, err := source.Paths(root)
	if err != nil {
		return nil,
			fmt.Errorf(
				"entry points: %w",
				err,
			)
	}

	return filesFindings(
			modulePath,
			directoriesOf(paths),
			source.WithoutMocks(files),
		),
		nil
}

func filesFindings(
	modulePath string,
	sourceDirectories []string,
	files []source.File,
) (findings []report.Finding) {
	names := packageNames(files)
	interfaces := packageInterfaces(files)
	for _, file := range files {
		directory := filepath.ToSlash(filepath.Dir(file.Rel))
		findings = append(
			findings,
			unitFindings(unit{
				file:       file,
				modulePath: modulePath,
				directory:  directory,
				role: roleOf(
					file,
					sourceDirectories,
				),
				names:      names[directory],
				interfaces: interfaces[directory],
			})...,
		)
	}

	return findings
}

func unitFindings(
	found unit,
) (findings []report.Finding) {
	switch found.role {
	case containerRole:
		return containerFindings(found)
	case mainRole:
		return mainFindings(found)
	case commandRole:
		return commandFindings(found)
	}

	return nil
}

func containerFindings(
	found unit,
) (findings []report.Finding) {
	for _, function := range source.Nodes[*ast.FuncDecl](found.file.Syntax) {
		findings = append(
			findings,
			entryFindings(
				found,
				function,
			)...,
		)
	}

	return findings
}

func entryFindings(
	found unit,
	function *ast.FuncDecl,
) (findings []report.Finding) {
	isEntry := function.Recv == nil && function.Name.IsExported()
	if !isEntry {
		return nil
	}

	return slices.Concat(
		delegationFindings(
			found,
			function,
		),
		resultFindings(
			found,
			function,
		),
	)
}

func roleOf(
	file source.File,
	sourceDirectories []string,
) (found role) {
	if source.IsExample(file.Rel) {
		return noRole
	}

	base := filepath.Base(file.Rel)
	if file.Syntax.Name.Name == mainPackage {
		return commandFileRole(base)
	}

	isContainer := base == containerFile && (source.IsModule(
		filepath.Dir(file.Path),
		sourceDirectories,
	) || !source.IsInternalTree(filepath.Dir(file.Rel)))
	if isContainer {
		return containerRole
	}

	return noRole
}

func commandFileRole(
	base string,
) (found role) {
	if base == mainFile {
		return mainRole
	}

	return commandRole
}

func directoriesOf(
	paths []string,
) (directories []string) {
	for _, path := range paths {
		directories = append(
			directories,
			filepath.Dir(path),
		)
	}

	slices.Sort(directories)
	return slices.Compact(directories)
}

func packageNames(
	files []source.File,
) (names map[string][]string) {
	names = map[string][]string{}
	for _, file := range files {
		directory := filepath.ToSlash(filepath.Dir(file.Rel))
		names[directory] = append(
			names[directory],
			declaredNames(file.Syntax)...,
		)
	}

	return names
}

func declaredNames(
	syntax *ast.File,
) (names []string) {
	for _, declaration := range syntax.Decls {
		names = append(
			names,
			declarationNames(declaration)...,
		)
	}

	return names
}

func declarationNames(
	declaration ast.Decl,
) (names []string) {
	switch typed := declaration.(type) {
	case *ast.FuncDecl:
		return functionName(typed)
	case *ast.GenDecl:
		return identifierNames(topLevelIdentifiers(typed))
	}

	return nil
}

func functionName(
	function *ast.FuncDecl,
) (names []string) {
	if function.Recv != nil {
		return nil
	}

	return []string{function.Name.Name}
}

func topLevelIdentifiers(
	general *ast.GenDecl,
) (identifiers []*ast.Ident) {
	for _, spec := range general.Specs {
		identifiers = append(
			identifiers,
			specIdentifiers(spec)...,
		)
	}

	return identifiers
}

func specIdentifiers(
	spec ast.Spec,
) (identifiers []*ast.Ident) {
	switch typed := spec.(type) {
	case *ast.TypeSpec:
		return []*ast.Ident{typed.Name}
	case *ast.ValueSpec:
		return typed.Names
	}

	return nil
}

func packageInterfaces(
	files []source.File,
) (interfaces map[string][]string) {
	interfaces = map[string][]string{}
	for _, file := range files {
		directory := filepath.ToSlash(filepath.Dir(file.Rel))
		interfaces[directory] = append(
			interfaces[directory],
			interfaceNames(file.Syntax)...,
		)
	}

	return interfaces
}

func interfaceNames(
	syntax *ast.File,
) (names []string) {
	for _, spec := range source.Nodes[*ast.TypeSpec](syntax) {
		names = append(
			names,
			interfaceName(spec)...,
		)
	}

	return names
}

func interfaceName(
	spec *ast.TypeSpec,
) (names []string) {
	_, isInterface := spec.Type.(*ast.InterfaceType)
	if !isInterface {
		return nil
	}

	return []string{spec.Name.Name}
}

func internalPath(
	directory string,
) (path string) {
	if directory == rootDirectory {
		return internalChild
	}

	return directory + "/" + internalChild
}
