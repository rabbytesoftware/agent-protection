package dependency

import (
	"errors"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/dependency/internal/compositionroot"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/dependency/internal/encapsulation"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/dependency/internal/mockimport"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/dependency/internal/reach"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/dependency/internal/siblingimport"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func Dependency(
	root string,
	files []source.File,
) (findings []report.Finding, err error) {
	siblings, siblingsErr := siblingimport.Siblingimport(
		root,
		files,
	)
	components, componentsErr := compositionroot.Compositionroot(
		root,
		files,
	)
	reaches, reachesErr := reach.Reach(
		root,
		files,
	)
	exposures, exposuresErr := encapsulation.Encapsulation(
		root,
		files,
	)
	err = errors.Join(
		siblingsErr,
		componentsErr,
		reachesErr,
		exposuresErr,
	)
	if err != nil {
		return nil,
			err
	}

	return slices.Concat(
			siblings,
			components,
			reaches,
			exposures,
			mockimport.Mockimport(files),
		),
		nil
}
