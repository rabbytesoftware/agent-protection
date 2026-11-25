package publicapi

import (
	"fmt"
	"go/ast"
	"path/filepath"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	apiRule        = "R14"
	vocabularyRule = "R57"
)

func Publicapi(
	root string,
	files []source.File,
) (findings []report.Finding, err error) {
	supportDirectories, err := testSupport(root)
	if err != nil {
		return nil,
			fmt.Errorf(
				"public api: %w",
				err,
			)
	}

	paths, err := source.Paths(root)
	if err != nil {
		return nil,
			fmt.Errorf(
				"public api: %w",
				err,
			)
	}

	sourceDirectories := directoriesOf(paths)
	structs := packageStructs(files)
	for _, file := range files {
		findings = append(
			findings,
			fileFindings(
				file,
				scopeOf(
					file,
					sourceDirectories,
					supportDirectories,
					structs,
				),
			)...,
		)
	}

	return findings,
		nil
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

func fileFindings(
	file source.File,
	found scope,
) (findings []report.Finding) {
	for _, declaration := range file.Syntax.Decls {
		findings = append(
			findings,
			declarationFindings(
				file,
				found,
				declaration,
			)...,
		)
	}

	return findings
}

func declarationFindings(
	file source.File,
	found scope,
	declaration ast.Decl,
) (findings []report.Finding) {
	findings = apiFindings(
		file,
		found,
		declaration,
	)
	isVocabularyOnly := len(findings) == 0 && found.vocabulary
	if !isVocabularyOnly {
		return findings
	}

	return vocabularyFindings(
		file,
		found,
		declaration,
	)
}
