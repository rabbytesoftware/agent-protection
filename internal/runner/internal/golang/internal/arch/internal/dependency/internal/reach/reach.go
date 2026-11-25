package reach

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
	reachRule    = "R15"
	exampleRule  = "R61"
	importQuotes = "\"`"
	moduleRoot   = "."
)

func Reach(
	root string,
	files []source.File,
) (findings []report.Finding, err error) {
	modulePath, err := source.ModulePath(root)
	if err != nil {
		return nil,
			fmt.Errorf(
				"reach: %w",
				err,
			)
	}

	modules := moduleDirectories(files)
	for _, file := range files {
		findings = append(
			findings,
			fileFindings(
				modulePath,
				modules,
				file,
			)...,
		)
	}

	return findings,
		nil
}

func fileFindings(
	modulePath string,
	modules []string,
	file source.File,
) (findings []report.Finding) {
	for _, spec := range file.Syntax.Imports {
		findings = append(
			findings,
			importFindings(
				modulePath,
				modules,
				file,
				spec,
			)...,
		)
	}

	return findings
}

func moduleDirectories(
	files []source.File,
) (modules []string) {
	directories := map[string]bool{}
	for _, file := range files {
		directories[filepath.Dir(file.Rel)] = true
	}

	sorted := slices.Sorted(maps.Keys(directories))
	for _, directory := range sorted {
		modules = append(
			modules,
			asModule(
				directory,
				sorted,
			)...,
		)
	}

	return modules
}

func asModule(
	directory string,
	directories []string,
) (module []string) {
	isModule := directory != moduleRoot && source.IsModule(
		directory,
		directories,
	)
	if !isModule {
		return nil
	}

	return []string{directory}
}

func importFindings(
	modulePath string,
	modules []string,
	file source.File,
	spec *ast.ImportSpec,
) (findings []report.Finding) {
	imported, inModule := importedDirectory(
		modulePath,
		spec,
	)
	if !inModule {
		return nil
	}

	importer := filepath.Dir(file.Rel)
	return slices.Concat(
		reachFindings(
			modules,
			importer,
			imported,
			file,
			spec,
		),
		exampleFindings(
			importer,
			imported,
			file,
			spec,
		),
	)
}

func reachFindings(
	modules []string,
	importer string,
	imported string,
	file source.File,
	spec *ast.ImportSpec,
) (findings []report.Finding) {
	module, reaches := reachedModule(
		modules,
		importer,
		imported,
	)
	if !reaches {
		return nil
	}

	return []report.Finding{{
		Rule:     reachRule,
		Location: file.Location(spec),
		Snippet: fmt.Sprintf(
			"reaches into module %s (imports %s): import %s itself and use what its API exposes",
			filepath.ToSlash(module),
			filepath.ToSlash(imported),
			filepath.ToSlash(module),
		),
	}}
}

func reachedModule(
	modules []string,
	importer string,
	imported string,
) (module string, reaches bool) {
	reached := []string{}
	for _, candidate := range modules {
		reached = append(
			reached,
			reachedCandidate(
				candidate,
				importer,
				imported,
			)...,
		)
	}

	if len(reached) == 0 {
		return "",
			false
	}

	return reached[0],
		true
}

func reachedCandidate(
	candidate string,
	importer string,
	imported string,
) (reached []string) {
	isReach := isBelow(
		imported,
		candidate,
	) && !isWithin(
		importer,
		candidate,
	)
	if !isReach {
		return nil
	}

	return []string{candidate}
}

func exampleFindings(
	importer string,
	imported string,
	file source.File,
	spec *ast.ImportSpec,
) (findings []report.Finding) {
	isProductionImport := source.IsExample(imported) && !source.IsExample(importer)
	if !isProductionImport {
		return nil
	}

	return []report.Finding{{
		Rule:     exampleRule,
		Location: file.Location(spec),
		Snippet: fmt.Sprintf(
			"imports example package %s: production code never depends on examples; move what it needs out of the example tree into a production package",
			filepath.ToSlash(imported),
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
	if importPath == modulePath {
		return moduleRoot,
			true
	}

	imported, inModule := strings.CutPrefix(
		importPath,
		modulePath+"/",
	)
	return filepath.FromSlash(imported),
		inModule
}

func isWithin(
	directory string,
	module string,
) (within bool) {
	return directory == module || isBelow(
		directory,
		module,
	)
}

func isBelow(
	directory string,
	module string,
) (below bool) {
	return strings.HasPrefix(
		directory,
		module+string(filepath.Separator),
	)
}
