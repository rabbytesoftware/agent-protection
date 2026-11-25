package compositionroot

import (
	"fmt"
	"go/ast"
	"maps"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule              = "R48"
	importQuotes      = "\"`"
	moduleRoot        = "."
	internalDirectory = "internal"
	mainPackage       = "main"
	sentinelPrefix    = "Err"
)

func Compositionroot(
	root string,
	files []source.File,
) (findings []report.Finding, err error) {
	modulePath, err := source.ModulePath(root)
	if err != nil {
		return nil,
			fmt.Errorf(
				"composition roots: %w",
				err,
			)
	}

	packages := groupByDirectory(files)
	directories := slices.Sorted(maps.Keys(packages))
	vocabularies := source.DeclarationPackages(files)
	for _, directory := range directories {
		findings = append(
			findings,
			moduleFindings(
				module{path: modulePath, vocabularies: vocabularies},
				directory,
				directories,
				packages[directory],
			)...,
		)
	}

	return findings,
		nil
}

func groupByDirectory(
	files []source.File,
) (packages map[string][]source.File) {
	packages = map[string][]source.File{}
	for _, file := range files {
		directory := filepath.Dir(file.Rel)
		packages[directory] = append(
			packages[directory],
			file,
		)
	}

	return packages
}

func moduleFindings(
	owner module,
	directory string,
	directories []string,
	files []source.File,
) (findings []report.Finding) {
	isWiredByContainer := source.IsModule(
		directory,
		directories,
	) && !slices.ContainsFunc(
		files,
		isMain,
	)
	if !isWiredByContainer {
		return nil
	}

	name := moduleName(
		directory,
		files,
	)
	return slices.Concat(
		missingContainerFindings(
			directory,
			name,
			files,
		),
		strayWiringFindings(
			owner,
			directory,
			name,
			files,
		),
	)
}

func strayWiringFindings(
	owner module,
	directory string,
	name string,
	files []source.File,
) (findings []report.Finding) {
	for _, file := range slices.DeleteFunc(
		slices.Clone(files),
		isContainer,
	) {
		findings = append(
			findings,
			wiringFindings(
				owner,
				directory,
				name,
				file,
			)...,
		)
	}

	return findings
}

func missingContainerFindings(
	directory string,
	name string,
	files []source.File,
) (findings []report.Finding) {
	if slices.ContainsFunc(
		files,
		isContainer,
	) {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: filepath.ToSlash(directory),
		Snippet: fmt.Sprintf(
			"module %s has internal packages but no container.go: add container.go and construct them there",
			name,
		),
	}}
}

func wiringFindings(
	owner module,
	directory string,
	name string,
	file source.File,
) (findings []report.Finding) {
	for _, spec := range file.Syntax.Imports {
		findings = append(
			findings,
			wiringFinding(
				owner,
				directory,
				name,
				file,
				spec,
			)...,
		)
	}

	return findings
}

func wiringFinding(
	owner module,
	directory string,
	name string,
	file source.File,
	spec *ast.ImportSpec,
) (findings []report.Finding) {
	imported, inModule := importedDirectory(
		owner.path,
		spec,
	)
	isOwnInternal := inModule && strings.HasPrefix(
		imported,
		filepath.Join(
			directory,
			internalDirectory,
		)+string(filepath.Separator),
	)
	isExempt := !isOwnInternal || isReExport(
		owner,
		imported,
		file,
		spec,
	)
	if isExempt {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: file.Location(spec),
		Snippet: fmt.Sprintf(
			"only container.go wires %s's internal packages",
			name,
		),
	}}
}

func importedDirectory(
	modulePath string,
	spec *ast.ImportSpec,
) (directory string, inModule bool) {
	importPath := strings.Trim(
		spec.Path.Value,
		importQuotes,
	)
	imported, inModule := strings.CutPrefix(
		importPath,
		modulePath+"/",
	)
	return filepath.FromSlash(imported),
		inModule
}

func moduleName(
	directory string,
	files []source.File,
) (name string) {
	if directory == moduleRoot {
		return files[0].Syntax.Name.Name
	}

	return filepath.ToSlash(directory)
}

func isMain(
	file source.File,
) (isMainPackage bool) {
	return file.Syntax.Name.Name == mainPackage
}

func isContainer(
	file source.File,
) (isRoot bool) {
	return source.IsCompositionRoot(
		file.Path,
		file.Syntax.Name.Name,
	)
}
