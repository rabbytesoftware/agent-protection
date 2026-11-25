package siblingimport

import (
	"cmp"
	"fmt"
	"go/ast"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	chainRule      = "R16"
	exampleLimit   = 3
	directFormat   = "imports sibling %s, which itself imports sibling %s: siblings form one layer; have the parent's container.go pass %s in, or make %s a foundation by dropping its own sibling imports"
	throughFormat  = "imports sibling %s (through %s, the sibling this package sits under), which itself imports sibling %s: siblings form one layer; have the parent's container.go pass %s in, or make %s a foundation by dropping its own sibling imports"
	internalFormat = "imports %s, a sibling of the %s/ directory this package lives in (all of %s/ counts as one package beside %s), and %s itself imports sibling %s: siblings form one layer; have the parent's container.go pass %s in, or make %s a foundation by dropping its own sibling imports"
	spreadFormat   = "; %d imports %s reach %s the same way, for example %s"
	scopeFormat    = "across %d packages under %s"
	samePackage    = "in this package"
	directorySlash = "/"
)

func chainsOf(
	modulePath string,
	files []source.File,
	logic map[string]bool,
) (findings []report.Finding) {
	directories := directoryNodes(files)
	lifted := behaviorGraph(
		moduleEdges(
			modulePath,
			directories,
			files,
		),
		logic,
	)
	for _, group := range groupedByEdge(offensesOf(
		modulePath,
		directories,
		lifted,
		files,
	)) {
		findings = append(
			findings,
			chainFinding(
				directories,
				lifted,
				group,
			),
		)
	}

	return findings
}

func offensesOf(
	modulePath string,
	directories map[string]bool,
	lifted map[string][]string,
	files []source.File,
) (offenses []offense) {
	for _, file := range files {
		offenses = append(
			offenses,
			fileOffenses(
				modulePath,
				directories,
				lifted,
				file,
			)...,
		)
	}

	slices.SortFunc(
		offenses,
		byPosition,
	)
	return offenses
}

func behaviorGraph(
	edges []edge,
	logic map[string]bool,
) (lifted map[string][]string) {
	return liftedGraph(behaviorEdges(
		edges,
		logic,
	))
}

func behaviorEdges(
	edges []edge,
	logic map[string]bool,
) (kept []edge) {
	for _, found := range edges {
		kept = append(
			kept,
			behaviorEdge(
				found,
				logic,
			)...,
		)
	}

	return kept
}

func behaviorEdge(
	found edge,
	logic map[string]bool,
) (kept []edge) {
	if !logic[found.imported] {
		return nil
	}

	return []edge{found}
}

func fileOffenses(
	modulePath string,
	nodes map[string]bool,
	lifted map[string][]string,
	file source.File,
) (offenses []offense) {
	for _, spec := range file.Syntax.Imports {
		offenses = append(
			offenses,
			specOffense(
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

	return offenses
}

func specOffense(
	lifted map[string][]string,
	file source.File,
	spec *ast.ImportSpec,
	edges []edge,
) (offenses []offense) {
	isChain := len(edges) > 0 && len(lifted[edges[0].to]) > 0
	if !isChain {
		return nil
	}

	return []offense{{
		found:    edges[0],
		rel:      filepath.ToSlash(file.Rel),
		line:     file.Line(spec),
		location: file.Location(spec),
	}}
}

func byPosition(
	left offense,
	right offense,
) (order int) {
	return cmp.Or(
		cmp.Compare(
			left.rel,
			right.rel,
		),
		cmp.Compare(
			left.line,
			right.line,
		),
	)
}

func groupedByEdge(
	offenses []offense,
) (groups [][]offense) {
	grouped := map[edge][]offense{}
	order := []edge{}
	for _, found := range offenses {
		key := edge{
			from: found.found.from,
			to:   found.found.to,
		}
		order = appendUnique(
			order,
			key,
		)
		grouped[key] = append(
			grouped[key],
			found,
		)
	}

	for _, key := range order {
		groups = append(
			groups,
			grouped[key],
		)
	}

	return groups
}

func appendUnique(
	order []edge,
	key edge,
) (extended []edge) {
	if slices.Contains(
		order,
		key,
	) {
		return order
	}

	return append(
		order,
		key,
	)
}

func chainFinding(
	nodes map[string]bool,
	lifted map[string][]string,
	group []offense,
) (finding report.Finding) {
	first := group[0]
	return report.Finding{
		Rule:     chainRule,
		Location: first.location,
		Snippet: chainSnippet(
			nodes,
			first.found,
			siblingName(
				nodes,
				lifted[first.found.to][0],
			),
		) + spread(
			nodes,
			group,
		),
	}
}

func chainSnippet(
	nodes map[string]bool,
	found edge,
	next string,
) (snippet string) {
	sibling := siblingName(
		nodes,
		found.to,
	)
	if found.importer == found.from {
		return fmt.Sprintf(
			directFormat,
			sibling,
			next,
			sibling,
			sibling,
		)
	}

	if source.IsInternalDirectory(found.from) {
		return internalSnippet(
			found,
			sibling,
			next,
		)
	}

	return fmt.Sprintf(
		throughFormat,
		sibling,
		siblingName(
			nodes,
			found.from,
		),
		next,
		sibling,
		sibling,
	)
}

func internalSnippet(
	found edge,
	sibling string,
	next string,
) (snippet string) {
	container := filepath.ToSlash(found.from)
	return fmt.Sprintf(
		internalFormat,
		sibling,
		container,
		container,
		sibling,
		sibling,
		next,
		sibling,
		sibling,
	)
}

func spread(
	nodes map[string]bool,
	group []offense,
) (note string) {
	if len(group) == 1 {
		return ""
	}

	packages := []string{}
	examples := []string{}
	for index, found := range group {
		packages = append(
			packages,
			filepath.Dir(found.rel),
		)
		examples = appendExample(
			examples,
			index,
			found,
		)
	}

	slices.Sort(packages)
	return fmt.Sprintf(
		spreadFormat,
		len(group),
		scope(
			nodes,
			group[0].found.from,
			len(slices.Compact(packages)),
		),
		siblingName(
			nodes,
			group[0].found.to,
		),
		strings.Join(
			examples,
			nameSeparator,
		),
	)
}

func scope(
	nodes map[string]bool,
	from string,
	packages int,
) (note string) {
	if packages == 1 {
		return samePackage
	}

	return fmt.Sprintf(
		scopeFormat,
		packages,
		nodeName(
			nodes,
			from,
		),
	)
}

func nodeName(
	nodes map[string]bool,
	node string,
) (name string) {
	if source.IsInternalDirectory(node) {
		return filepath.ToSlash(node) + directorySlash
	}

	return siblingName(
		nodes,
		node,
	)
}

func appendExample(
	examples []string,
	index int,
	found offense,
) (extended []string) {
	isExample := index > 0 && index <= exampleLimit
	if !isExample {
		return examples
	}

	return append(
		examples,
		found.location,
	)
}
