package golang

import (
	"fmt"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/lint"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

// Golang checks the Go module rooted at root against the full Go manifest:
// package architecture and layout rules, per-declaration syntax rules, and
// the golangci-lint proxies. Source and test files are parsed once and
// shared by every check.
func Golang(
	root string,
) (findings []report.Finding, err error) {
	files, tests, err := load(root)
	if err != nil {
		return nil,
			err
	}

	architecture, err := arch.Arch(
		root,
		files,
	)
	if err != nil {
		return nil,
			fmt.Errorf(
				"architecture checks: %w",
				err,
			)
	}

	linted, err := lint.Lint(root)
	if err != nil {
		return nil,
			fmt.Errorf(
				"lint checks: %w",
				err,
			)
	}

	return slices.Concat(
			architecture,
			astrules.Astrules(
				files,
				tests,
			),
			linted,
		),
		nil
}

func load(
	root string,
) (files []source.File, tests []source.File, err error) {
	files, err = source.Source(root)
	if err != nil {
		return nil,
			nil,
			fmt.Errorf(
				"load go sources: %w",
				err,
			)
	}

	tests, err = source.Tests(root)
	if err != nil {
		return nil,
			nil,
			fmt.Errorf(
				"load go tests: %w",
				err,
			)
	}

	return files,
		tests,
		nil
}
