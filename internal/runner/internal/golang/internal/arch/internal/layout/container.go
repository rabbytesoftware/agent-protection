package layout

import (
	"errors"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/layout/internal/internalboundary"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/layout/internal/maxfilesperdir"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/layout/internal/packageset"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/layout/internal/publicapi"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/layout/internal/testpairing"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func Layout(
	root string,
	files []source.File,
) (findings []report.Finding, err error) {
	surfaces, surfacesErr := publicapi.Publicapi(
		root,
		files,
	)
	names, namesErr := packageset.Packageset(
		root,
		files,
	)
	boundaries, boundariesErr := internalboundary.Internalboundary(root)
	pairings, pairingsErr := testpairing.Testpairing(root)
	crowded, crowdedErr := maxfilesperdir.Maxfilesperdir(root)
	err = errors.Join(
		surfacesErr,
		namesErr,
		boundariesErr,
		pairingsErr,
		crowdedErr,
	)
	if err != nil {
		return nil,
			err
	}

	return slices.Concat(
			surfaces,
			names,
			boundaries,
			pairings,
			crowded,
		),
		nil
}
