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
	chainRule    = "R16"
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

	files = slices.DeleteFunc(
		slices.Clone(files),
		isMock,
	)
	packages := packageNodes(files)
	return slices.Concat(
			chainsOf(
				modulePath,
				files,
			),
			wiringFindings(
				packages,
				moduleEdges(
					modulePath,
					packages,
					files,
				),
				logicDirectories(files),
			),
		),
		nil
}

func chainsOf(
	modulePath string,
	files []source.File,
) (findings []report.Finding) {
	directories := directoryNodes(files)
	lifted := liftedGraph(moduleEdges(
		modulePath,
		directories,
		files,
	))
	for _, file := range files {
		findings = append(
			findings,
			chainFindings(
				modulePath,
				directories,
				lifted,
				file,
			)...,
		)
	}

	return findings
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

func chainFindings(
	modulePath string,
	nodes map[string]bool,
	lifted map[string][]string,
	file source.File,
) (findings []report.Finding) {
	for _, spec := range file.Syntax.Imports {
		findings = append(
			findings,
			chainFinding(
				nodes,
				lifted,
				file,
				spec,
				siblingEdge(
					modulePath,
					nodes,
					filepath.Dir(file.Rel),
					spec,
				),
			)...,
		)
	}

	return findings
}

func chainFinding(
	nodes map[string]bool,
	lifted map[string][]string,
	file source.File,
	spec *ast.ImportSpec,
	edges []edge,
) (findings []report.Finding) {
	isChain := len(edges) > 0 && len(lifted[edges[0].to]) > 0
	if !isChain {
		return nil
	}

	found := edges[0]
	sibling := siblingName(
		nodes,
		found.to,
	)
	return []report.Finding{{
		Rule:     chainRule,
		Location: file.Location(spec),
		Snippet: fmt.Sprintf(
			"imports sibling %s%s, which itself imports sibling %s: siblings form one layer; have the parent's container.go pass %s in",
			sibling,
			through(
				nodes,
				found,
			),
			siblingName(
				nodes,
				lifted[found.to][0],
			),
			sibling,
		),
	}}
}

func through(
	nodes map[string]bool,
	found edge,
) (note string) {
	if found.importer == found.from {
		return ""
	}

	return fmt.Sprintf(
		" (through %s)",
		siblingName(
			nodes,
			found.from,
		),
	)
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

func isMock(
	file source.File,
) (mock bool) {
	return source.IsMock(file.Rel)
}
