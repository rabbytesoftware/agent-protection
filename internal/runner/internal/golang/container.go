package golang

import (
	"fmt"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/lint"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/rewrite"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax"
)

func Golang(
	root string,
) (findings []report.Finding, err error) {
	files, tests, err := source.Load(root)
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
			syntax.Syntax(
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
	files, tests, err := source.Load(root)
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
		syntax.Edits,
	)
}
