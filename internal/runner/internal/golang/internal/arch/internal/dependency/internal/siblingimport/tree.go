package siblingimport

import (
	"path/filepath"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func packageNodes(
	files []source.File,
) (nodes map[string]bool) {
	nodes = map[string]bool{moduleRoot: true}
	for _, file := range files {
		addNode(
			nodes,
			filepath.Dir(file.Rel),
			source.IsInternalDirectory,
		)
	}

	return nodes
}

func directoryNodes(
	files []source.File,
) (nodes map[string]bool) {
	nodes = map[string]bool{moduleRoot: true}
	for _, file := range files {
		addNode(
			nodes,
			filepath.Dir(file.Rel),
			isDirectory,
		)
	}

	return nodes
}

func addNode(
	nodes map[string]bool,
	directory string,
	isNode func(string) bool,
) {
	nodes[directory] = true
	for current := directory; current != moduleRoot; current = filepath.Dir(current) {
		nodes[current] = nodes[current] || isNode(current)
	}
}

func isDirectory(
	directory string,
) (isNode bool) {
	return directory != ""
}

func ancestry(
	nodes map[string]bool,
	directory string,
) (chain []string) {
	chain = []string{directory}
	for current := directory; current != moduleRoot; current = filepath.Dir(current) {
		chain = append(
			chain,
			nodeOrNone(
				nodes,
				filepath.Dir(current),
			)...,
		)
	}

	return chain
}

func nodeOrNone(
	nodes map[string]bool,
	directory string,
) (node []string) {
	if !nodes[directory] {
		return nil
	}

	return []string{directory}
}

func liftedEdge(
	nodes map[string]bool,
	importer string,
	imported string,
) (edges []edge) {
	importers := ancestry(
		nodes,
		importer,
	)
	importeds := ancestry(
		nodes,
		imported,
	)
	below, importedBelow := commonAncestor(
		importers,
		importeds,
	)
	isRelated := below == 0 || importedBelow == 0
	if isRelated {
		return nil
	}

	return []edge{{
		from:     importers[below-1],
		to:       importeds[importedBelow-1],
		importer: importer,
		imported: imported,
	}}
}

func commonAncestor(
	importers []string,
	importeds []string,
) (importerIndex int, importedIndex int) {
	importerIndex = slices.IndexFunc(
		importers,
		isAmong(importeds),
	)
	return importerIndex,
		slices.Index(
			importeds,
			importers[importerIndex],
		)
}

func isAmong(
	directories []string,
) (isMember func(string) bool) {
	return func(
		directory string,
	) (found bool) {
		return slices.Contains(
			directories,
			directory,
		)
	}
}

func siblingName(
	nodes map[string]bool,
	node string,
) (name string) {
	parent := ancestry(
		nodes,
		node,
	)[1]
	return filepath.ToSlash(strings.TrimPrefix(
		node,
		parent+string(filepath.Separator),
	))
}

func owner(
	nodes map[string]bool,
	directory string,
) (parent string) {
	parent = ancestry(
		nodes,
		directory,
	)[1]
	if source.IsInternalDirectory(parent) {
		return filepath.ToSlash(filepath.Dir(parent))
	}

	return filepath.ToSlash(parent)
}
