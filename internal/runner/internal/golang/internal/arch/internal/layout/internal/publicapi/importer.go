package publicapi

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	goExtension       = ".go"
	testdataDirectory = "testdata"
	importQuotes      = "\"`"
	testingPackage    = "testing"
)

type importer struct {
	directory string
	paths     []string
	test      bool
}

func testSupport(
	root string,
	modulePath string,
	sourceDirectories []string,
) (directories []string, err error) {
	found, err := importers(root)
	if err != nil {
		return nil,
			err
	}

	productionImporters := slices.DeleteFunc(
		slices.Clone(found),
		isTestImporter,
	)
	candidates := without(
		slices.Concat(
			importedBy(
				modulePath,
				slices.DeleteFunc(
					found,
					isProductionImporter,
				),
			),
			testHelpers(productionImporters),
		),
		importedBy(
			modulePath,
			productionImporters,
		),
	)
	return slices.DeleteFunc(
			candidates,
			moduleDirectory(
				root,
				sourceDirectories,
			),
		),
		nil
}

func moduleDirectory(
	root string,
	sourceDirectories []string,
) (isModule func(string) bool) {
	return func(
		directory string,
	) (excluded bool) {
		return directory == rootDirectory || source.IsModule(
			filepath.Join(
				root,
				directory,
			),
			sourceDirectories,
		)
	}
}

func testHelpers(
	productionImporters []importer,
) (directories []string) {
	for _, each := range slices.DeleteFunc(
		slices.Clone(productionImporters),
		isNotTestingImporter,
	) {
		directories = append(
			directories,
			each.directory,
		)
	}

	return directories
}

func importers(
	root string,
) (found []importer, err error) {
	directories, err := source.Directories(root)
	if err != nil {
		return nil,
			err
	}

	failures := []error{}
	for _, directory := range directories {
		directoryFound, directoryErr := directoryImporters(
			root,
			directory,
		)
		found = append(
			found,
			directoryFound...,
		)
		failures = append(
			failures,
			directoryErr,
		)
	}

	return found,
		errors.Join(failures...)
}

func directoryImporters(
	root string,
	directory string,
) (found []importer, err error) {
	rel, err := source.RelPath(
		root,
		directory,
	)
	if err != nil {
		return nil,
			err
	}

	if isTestdata(rel) {
		return nil,
			nil
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil,
			fmt.Errorf(
				"read %s: %w",
				directory,
				err,
			)
	}

	return entryImporters(
		directory,
		filepath.ToSlash(rel),
		slices.DeleteFunc(
			entries,
			isNotGoFile,
		),
	)
}

func entryImporters(
	directory string,
	rel string,
	entries []os.DirEntry,
) (found []importer, err error) {
	failures := []error{}
	for _, entry := range entries {
		entryFound, entryErr := fileImporter(
			rel,
			filepath.Join(
				directory,
				entry.Name(),
			),
		)
		found = append(
			found,
			entryFound...,
		)
		failures = append(
			failures,
			entryErr,
		)
	}

	return found,
		errors.Join(failures...)
}

func fileImporter(
	directory string,
	path string,
) (found []importer, err error) {
	syntax, err := parser.ParseFile(
		token.NewFileSet(),
		path,
		nil,
		parser.ImportsOnly,
	)
	if err != nil {
		return nil,
			fmt.Errorf(
				"parse imports of %s: %w",
				path,
				err,
			)
	}

	return []importer{{directory: directory, paths: importPaths(syntax), test: source.IsTest(path)}},
		nil
}

func importPaths(
	syntax *ast.File,
) (paths []string) {
	for _, spec := range syntax.Imports {
		paths = append(
			paths,
			strings.Trim(
				spec.Path.Value,
				importQuotes,
			),
		)
	}

	return paths
}

func importedBy(
	modulePath string,
	found []importer,
) (packages []string) {
	for _, each := range found {
		packages = append(
			packages,
			foreignPackages(
				modulePath,
				each,
			)...,
		)
	}

	slices.Sort(packages)
	return slices.Compact(packages)
}

func foreignPackages(
	modulePath string,
	each importer,
) (packages []string) {
	for _, path := range each.paths {
		packages = append(
			packages,
			foreignPackage(
				modulePath,
				each.directory,
				path,
			)...,
		)
	}

	return packages
}

func foreignPackage(
	modulePath string,
	directory string,
	path string,
) (imported []string) {
	if path == modulePath {
		path = modulePath + "/" + rootDirectory
	}

	rel, inModule := strings.CutPrefix(
		path,
		modulePath+"/",
	)
	isForeign := inModule && rel != directory
	if !isForeign {
		return nil
	}

	return []string{rel}
}

func without(
	packages []string,
	excluded []string,
) (kept []string) {
	for _, imported := range packages {
		kept = append(
			kept,
			absent(
				imported,
				excluded,
			)...,
		)
	}

	return kept
}

func absent(
	imported string,
	excluded []string,
) (kept []string) {
	if slices.Contains(
		excluded,
		imported,
	) {
		return nil
	}

	return []string{imported}
}

func isTestdata(
	rel string,
) (inside bool) {
	return slices.Contains(
		strings.Split(
			filepath.ToSlash(rel),
			"/",
		),
		testdataDirectory,
	)
}

func isNotGoFile(
	entry os.DirEntry,
) (excluded bool) {
	return entry.IsDir() || filepath.Ext(entry.Name()) != goExtension
}

func isTestImporter(
	each importer,
) (isTest bool) {
	return each.test
}

func isProductionImporter(
	each importer,
) (isProduction bool) {
	return !each.test
}

func isNotTestingImporter(
	each importer,
) (excluded bool) {
	return !slices.Contains(
		each.paths,
		testingPackage,
	)
}
