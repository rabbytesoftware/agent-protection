package siblingimports

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
	siblingRule  = "R16"
	wiringRule   = "R47"
	importQuotes = "\"`"
)

func Siblingimports(
	root string,
	files []source.File,
) (findings []report.Finding, err error) {
	modulePath, err := source.ModulePath(root)
	if err != nil {
		return nil,
			fmt.Errorf(
				"sibling imports: %w",
				err,
			)
	}

	for _, file := range slices.DeleteFunc(
		slices.Clone(files),
		isExempt,
	) {
		findings = append(
			findings,
			fileFindings(
				modulePath,
				file,
			)...,
		)
	}

	return findings,
		nil
}

func fileFindings(
	modulePath string,
	file source.File,
) (findings []report.Finding) {
	for _, spec := range file.Syntax.Imports {
		findings = append(
			findings,
			importFindings(
				modulePath,
				file,
				spec,
			)...,
		)
	}

	if len(findings) >= source.CompositionRootThreshold {
		return []report.Finding{wiringFinding(
			file,
			len(findings),
		)}
	}

	return findings
}

func importFindings(
	modulePath string,
	file source.File,
	spec *ast.ImportSpec,
) (findings []report.Finding) {
	importPath := strings.Trim(
		spec.Path.Value,
		importQuotes,
	)
	imported, inModule := strings.CutPrefix(
		importPath,
		modulePath+"/",
	)
	isViolation := inModule && isSibling(
		packageDirectory(file),
		imported,
	)
	if !isViolation {
		return nil
	}

	return []report.Finding{{
		Rule:     siblingRule,
		Location: file.Location(spec),
		Snippet: fmt.Sprintf(
			"imports sibling package %s: reach it through the shared parent's composition root",
			importPath,
		),
	}}
}

func wiringFinding(
	file source.File,
	siblings int,
) (finding report.Finding) {
	return report.Finding{
		Rule:     wiringRule,
		Location: file.Location(file.Syntax.Name),
		Snippet: fmt.Sprintf(
			"imports %d sibling packages, acting as a composition root: name it container.go, wire.go or %s.go",
			siblings,
			file.Syntax.Name.Name,
		),
	}
}

func isSibling(
	importer string,
	imported string,
) (sibling bool) {
	sharesParent := path.Dir(importer) == path.Dir(imported)
	return sharesParent && source.IsInternalTree(imported)
}

func isExempt(
	file source.File,
) (exempt bool) {
	return source.IsCompositionRoot(
		file.Path,
		file.Syntax.Name.Name,
	) || !source.IsInternalTree(packageDirectory(file))
}

func packageDirectory(
	file source.File,
) (directory string) {
	return filepath.ToSlash(filepath.Dir(file.Rel))
}
