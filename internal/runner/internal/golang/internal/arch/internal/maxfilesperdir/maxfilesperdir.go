package maxfilesperdir

import (
	"errors"
	"fmt"
	"maps"
	"path/filepath"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule     = "R46"
	maxFiles = 8
)

// Maxfilesperdir reports every directory other than root that holds more
// than eight non-test .go files, in directory order.
func Maxfilesperdir(
	root string,
) (findings []report.Finding, err error) {
	paths, err := source.Paths(root)
	if err != nil {
		return nil,
			fmt.Errorf(
				"max files per directory: %w",
				err,
			)
	}

	counts := map[string]int{}
	for _, path := range paths {
		counts[filepath.Dir(path)]++
	}

	failures := []error{}
	for _, directory := range slices.Sorted(maps.Keys(counts)) {
		directoryFound, directoryErr := directoryFindings(
			root,
			directory,
			counts[directory],
		)
		findings = append(
			findings,
			directoryFound...,
		)
		failures = append(
			failures,
			directoryErr,
		)
	}

	return findings,
		errors.Join(failures...)
}

func directoryFindings(
	root string,
	directory string,
	count int,
) (findings []report.Finding, err error) {
	isWithinLimit := directory == filepath.Clean(root) || count <= maxFiles
	if isWithinLimit {
		return nil,
			nil
	}

	location, err := source.RelPath(
		root,
		directory,
	)
	if err != nil {
		return nil,
			fmt.Errorf(
				"max files per directory: %w",
				err,
			)
	}

	return []report.Finding{{
			Rule:     rule,
			Location: location,
			Snippet: fmt.Sprintf(
				"%d non-test source files, max %d: split the package",
				count,
				maxFiles,
			),
		}},
		nil
}
