package testmodule

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/rust/internal/clippy"
)

func Production(
	root string,
	warnings []clippy.Warning,
) (findings []report.Finding, err error) {
	testLines, err := inlineTestLines(
		root,
		warnings,
	)
	if err != nil {
		return nil,
			err
	}

	return productionFindings(
			warnings,
			testLines,
		),
		nil
}

func productionFindings(
	warnings []clippy.Warning,
	testLines map[string][]int,
) (findings []report.Finding) {
	for _, warning := range warnings {
		findings = append(
			findings,
			warningFindings(
				warning,
				testLines[warning.Path],
			)...,
		)
	}

	return findings
}

func inlineTestLines(
	root string,
	warnings []clippy.Warning,
) (lines map[string][]int, err error) {
	lines = map[string][]int{}
	failures := []error{}
	for _, path := range warnedPaths(warnings) {
		fileLines, readErr := Testmodule(filepath.Join(
			root,
			path,
		))
		lines[path] = fileLines
		failures = append(
			failures,
			readErr,
		)
	}

	err = errors.Join(failures...)
	if err != nil {
		return nil,
			err
	}

	return lines,
		nil
}

func warnedPaths(
	warnings []clippy.Warning,
) (paths []string) {
	for _, warning := range warnings {
		paths = append(
			paths,
			warning.Path,
		)
	}

	slices.Sort(paths)
	return slices.Compact(paths)
}

func warningFindings(
	warning clippy.Warning,
	testLines []int,
) (findings []report.Finding) {
	if slices.Contains(
		testLines,
		warning.Line,
	) {
		return nil
	}

	return []report.Finding{{
		Rule: warning.Rule,
		Location: fmt.Sprintf(
			"%s:%d",
			warning.Path,
			warning.Line,
		),
		Snippet: warning.Text,
	}}
}
