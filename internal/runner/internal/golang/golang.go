package golang

import (
	"fmt"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/lint"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/rewrite"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

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

func Fix(
	root string,
) (changed []string, err error) {
	files, tests, err := load(root)
	if err != nil {
		return nil,
			err
	}

	return rewrite.Rewrite(
		root,
		slices.Concat(
			files,
			tests,
		),
		astrules.Edits,
	)
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
