package siblingimport

import (
	"fmt"
	"go/ast"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	blankImport  = "_"
	moduleRoot   = "."
	importQuotes = "\"`"
)

func Siblingimport(
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

	files = source.WithoutMocks(files)
	packages := packageNodes(files)
	logic := logicDirectories(files)
	return slices.Concat(
			chainsOf(
				modulePath,
				files,
				logic,
			),
			wiringFindings(
				packages,
				moduleEdges(
					modulePath,
					packages,
					files,
				),
				logic,
			),
		),
		nil
}

func moduleEdges(
	modulePath string,
	nodes map[string]bool,
	files []source.File,
) (edges []edge) {
	for _, file := range files {
		edges = append(
			edges,
			fileEdges(
				modulePath,
				nodes,
				file,
			)...,
		)
	}

	return edges
}

func fileEdges(
	modulePath string,
	nodes map[string]bool,
	file source.File,
) (edges []edge) {
	for _, spec := range file.Syntax.Imports {
		edges = append(
			edges,
			siblingEdge(
				modulePath,
				nodes,
				filepath.Dir(file.Rel),
				spec,
			)...,
		)
	}

	return edges
}

func siblingEdge(
	modulePath string,
	nodes map[string]bool,
	importer string,
	spec *ast.ImportSpec,
) (edges []edge) {
	imported, inModule := importedDirectory(
		modulePath,
		spec,
	)
	isEdgeless := !inModule ||
		spec.Name != nil && spec.Name.Name == blankImport
	if isEdgeless {
		return nil
	}

	return liftedEdge(
		nodes,
		importer,
		imported,
	)
}

func liftedGraph(
	edges []edge,
) (lifted map[string][]string) {
	lifted = map[string][]string{}
	for _, found := range edges {
		lifted[found.from] = append(
			lifted[found.from],
			found.to,
		)
	}

	for from, targets := range lifted {
		lifted[from] = slices.Compact(slices.Sorted(slices.Values(targets)))
	}

	return lifted
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
