package mockimport

import (
	"fmt"
	"go/ast"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule           = "R63"
	importQuotes   = "\"`"
	testingPackage = "testing"
	snippetFormat  = "production file imports mock package %s: mocks and fakes exist for tests only; wire the real implementation here and import %s only from _test.go files"
)

func Mockimport(
	files []source.File,
) (findings []report.Finding) {
	for _, file := range slices.DeleteFunc(
		slices.Clone(files),
		isExempt,
	) {
		findings = append(
			findings,
			fileFindings(file)...,
		)
	}

	return findings
}

func fileFindings(
	file source.File,
) (findings []report.Finding) {
	for _, spec := range file.Syntax.Imports {
		findings = append(
			findings,
			importFindings(
				file,
				spec,
			)...,
		)
	}

	return findings
}

func importFindings(
	file source.File,
	spec *ast.ImportSpec,
) (findings []report.Finding) {
	imported := importPath(spec)
	if !source.IsMock(imported) {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: file.Location(spec),
		Snippet: fmt.Sprintf(
			snippetFormat,
			imported,
			imported,
		),
	}}
}

func isExempt(
	file source.File,
) (exempt bool) {
	return source.IsTest(file.Rel) || source.IsMock(file.Rel) || slices.Contains(
		importPaths(file.Syntax),
		testingPackage,
	)
}

func importPaths(
	syntax *ast.File,
) (paths []string) {
	for _, spec := range syntax.Imports {
		paths = append(
			paths,
			importPath(spec),
		)
	}

	return paths
}

func importPath(
	spec *ast.ImportSpec,
) (imported string) {
	return strings.Trim(
		spec.Path.Value,
		importQuotes,
	)
}
