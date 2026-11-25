package mockimport

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
	rule           = "R63"
	importQuotes   = "\"`"
	testingPackage = "testing"
	snippetFormat  = "production file imports mock package %s: mocks and fakes exist for tests only; wire the real implementation here and import %s only from _test.go files"
	usedFormat     = "%s is named like a mock but production file %s uses its declarations: mocks and fakes exist for tests only, so it is checked as production code; name the file after what it implements and give it its own test, or reference it only from _test.go files"
)

func Mockimport(
	files []source.File,
) (findings []report.Finding) {
	for _, file := range slices.DeleteFunc(
		source.WithoutMocks(files),
		isExempt,
	) {
		findings = append(
			findings,
			fileFindings(file)...,
		)
	}

	users := source.MockUsers(files)
	for _, mock := range slices.Sorted(maps.Keys(users)) {
		findings = append(
			findings,
			usedMockFindings(
				mock,
				users[mock],
			)...,
		)
	}

	return findings
}

func usedMockFindings(
	mock string,
	user string,
) (findings []report.Finding) {
	directory := filepath.Dir(mock)
	isReportedByImport := filepath.Dir(user) != directory && source.IsMock(directory)
	if isReportedByImport {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: mock,
		Snippet: fmt.Sprintf(
			usedFormat,
			filepath.Base(mock),
			user,
		),
	}}
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
	return source.IsTest(file.Rel) || slices.Contains(
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
