package siblingimport

import (
	"fmt"
	"go/ast"
	"maps"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	wiringRule      = "R47"
	containerFile   = "container.go"
	nameSeparator   = ", "
	wiringThreshold = 2
)

func wiringFindings(
	nodes map[string]bool,
	edges []edge,
	logic map[string]bool,
) (findings []report.Finding) {
	wired := wiringGraph(
		edges,
		logic,
	)
	for _, directory := range slices.Sorted(maps.Keys(wired)) {
		findings = append(
			findings,
			wiringFinding(
				nodes,
				directory,
				wired[directory],
			)...,
		)
	}

	return findings
}

func wiringGraph(
	edges []edge,
	logic map[string]bool,
) (wired map[string]map[string][]string) {
	wired = map[string]map[string][]string{}
	for _, found := range edges {
		addWiring(
			wired,
			found,
			logic,
		)
	}

	return wired
}

func addWiring(
	wired map[string]map[string][]string,
	found edge,
	logic map[string]bool,
) {
	if !logic[found.imported] {
		return
	}

	targets := wired[found.from]
	if targets == nil {
		targets = map[string][]string{}
		wired[found.from] = targets
	}

	targets[found.to] = append(
		targets[found.to],
		found.importer,
	)
}

func wiringFinding(
	nodes map[string]bool,
	directory string,
	siblings map[string][]string,
) (findings []report.Finding) {
	if len(siblings) < wiringThreshold {
		return nil
	}

	names := []string{}
	for _, sibling := range slices.Sorted(maps.Keys(siblings)) {
		names = append(
			names,
			siblingName(
				nodes,
				sibling,
			)+importedFrom(
				directory,
				siblings[sibling],
			),
		)
	}

	return []report.Finding{{
		Rule:     wiringRule,
		Location: filepath.ToSlash(directory),
		Snippet: fmt.Sprintf(
			"imports siblings %s: wiring siblings together is the parent's job; move it into %s and inject what this package needs through an interface it declares",
			strings.Join(
				names,
				nameSeparator,
			),
			path.Join(
				owner(
					nodes,
					directory,
				),
				containerFile,
			),
		),
	}}
}

func importedFrom(
	directory string,
	importers []string,
) (note string) {
	if slices.Contains(
		importers,
		directory,
	) {
		return ""
	}

	return fmt.Sprintf(
		" (from %s)",
		filepath.ToSlash(slices.Min(importers)),
	)
}

func logicDirectories(
	files []source.File,
) (logic map[string]bool) {
	logic = map[string]bool{}
	for _, file := range files {
		directory := filepath.Dir(file.Rel)
		logic[directory] = logic[directory] || slices.ContainsFunc(
			source.Nodes[*ast.FuncDecl](file.Syntax),
			isLogic,
		)
	}

	return logic
}

func isLogic(
	function *ast.FuncDecl,
) (hasLogic bool) {
	isRendering := function.Recv != nil && slices.Contains(
		source.ErrorMethods(),
		function.Name.Name,
	) && source.IsPlainRendering(function.Body)
	return function.Body != nil && !isRendering
}
