package arch

import (
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/dependency"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/layout"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func Arch(
	root string,
	files []source.File,
) (findings []report.Finding, err error) {
	shaped, err := layout.Layout(
		root,
		files,
	)
	if err != nil {
		return nil,
			err
	}

	wired, err := dependency.Dependency(
		root,
		files,
	)
	if err != nil {
		return nil,
			err
	}

	return slices.Concat(
			shaped,
			wired,
		),
		nil
}
