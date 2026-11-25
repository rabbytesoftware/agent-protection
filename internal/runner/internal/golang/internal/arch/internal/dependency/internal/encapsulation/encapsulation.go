package encapsulation

import (
	"fmt"
	"go/ast"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule            = "R19"
	mainPackage     = "main"
	importQuotes    = "\"`"
	internalSegment = "internal"
	blankImport     = "_"
)

func Encapsulation(
	root string,
	files []source.File,
) (findings []report.Finding, err error) {
	modulePath, err := source.ModulePath(root)
	if err != nil {
		return nil,
			fmt.Errorf(
				"encapsulation: %w",
				err,
			)
	}

	for _, imported := range slices.DeleteFunc(
		importedPackages(
			modulePath,
			files,
		),
		isNotExposed,
	) {
		findings = append(
			findings,
			exposureFindings(
				modulePath,
				imported,
				files,
			)...,
		)
	}

	return findings,
		nil
}

func importedPackages(
	modulePath string,
	files []source.File,
) (packages []string) {
	for _, file := range files {
		packages = append(
			packages,
			moduleImports(
				modulePath,
				file,
			)...,
		)
	}

	slices.Sort(packages)
	return slices.Compact(packages)
}

func moduleImports(
	modulePath string,
	file source.File,
) (packages []string) {
	for _, spec := range slices.DeleteFunc(
		slices.Clone(file.Syntax.Imports),
		isBlankImport,
	) {
		packages = append(
			packages,
			modulePackage(
				modulePath,
				spec,
			)...,
		)
	}

	return packages
}

func modulePackage(
	modulePath string,
	spec *ast.ImportSpec,
) (imported []string) {
	relative, inModule := strings.CutPrefix(
		strings.Trim(
			spec.Path.Value,
			importQuotes,
		),
		modulePath+"/",
	)
	if !inModule {
		return nil
	}

	return []string{relative}
}

func exposureFindings(
	modulePath string,
	imported string,
	files []source.File,
) (findings []report.Finding) {
	importers := importersOf(
		modulePath,
		imported,
		files,
	)
	directories := packageDirectories(importers)
	isSoleConsumer := len(directories) == 1 && importers[0].Syntax.Name.Name != mainPackage
	if !isSoleConsumer {
		return nil
	}

	if isUnderOwnInternal(
		imported,
		directories[0],
	) {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: filepath.FromSlash(imported),
		Snippet: fmt.Sprintf(
			"imported only by %s but visible to the whole module: move it under %s/internal/",
			directories[0],
			directories[0],
		),
	}}
}

func importersOf(
	modulePath string,
	imported string,
	files []source.File,
) (importers []source.File) {
	for _, file := range files {
		importers = append(
			importers,
			importerOrNone(
				modulePath,
				imported,
				file,
			)...,
		)
	}

	return importers
}

func importerOrNone(
	modulePath string,
	imported string,
	file source.File,
) (importer []source.File) {
	if !slices.Contains(
		moduleImports(
			modulePath,
			file,
		),
		imported,
	) {
		return nil
	}

	return []source.File{file}
}

func packageDirectories(
	files []source.File,
) (directories []string) {
	for _, file := range files {
		directories = append(
			directories,
			filepath.ToSlash(filepath.Dir(file.Rel)),
		)
	}

	slices.Sort(directories)
	return slices.Compact(directories)
}

func isUnderOwnInternal(
	imported string,
	importer string,
) (nested bool) {
	return strings.HasPrefix(
		imported,
		path.Join(
			importer,
			internalSegment,
		)+"/",
	)
}

func isNotExposed(
	imported string,
) (excluded bool) {
	return !source.IsInternalTree(imported)
}

func isBlankImport(
	spec *ast.ImportSpec,
) (blank bool) {
	return spec.Name != nil && spec.Name.Name == blankImport
}
