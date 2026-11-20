package rust

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/rust/internal/clippy"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/rust/internal/testmodule"
)

// Rust checks the crate at root with cargo clippy and reports every
// panic-risk warning outside test code: files whose path mentions "test"
// and inline #[cfg(test)] modules are skipped. It fails when clippy fails
// or a warned file cannot be read.
func Rust(
	root string,
) (findings []report.Finding, err error) {
	warnings, err := clippy.Clippy(root)
	if err != nil {
		return nil,
			err
	}
	testLines, err := inlineTestLines(
		root,
		warnings,
	)
	if err != nil {
		return nil,
			err
	}
	for _, warning := range warnings {
		findings = append(
			findings,
			warningFindings(
				warning,
				testLines[warning.Path],
			)...,
		)
	}
	return findings,
		nil
}

func inlineTestLines(
	root string,
	warnings []clippy.Warning,
) (lines map[string][]int, err error) {
	lines = map[string][]int{}
	failures := []error{}
	for _, path := range warnedPaths(warnings) {
		fileLines, readErr := testmodule.Testmodule(filepath.Join(
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
