package arch

import (
	"errors"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/compositionroot"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/encapsulation"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/internalboundary"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/maxfilesperdir"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/packageset"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/publicapi"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/reach"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/siblingimports"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/testpairing"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func Arch(
	root string,
	files []source.File,
) (findings []report.Finding, err error) {
	siblings, siblingsErr := siblingimports.Siblingimports(
		root,
		files,
	)
	components, componentsErr := compositionroot.Compositionroot(
		root,
		files,
	)
	boundaries, boundariesErr := internalboundary.Internalboundary(root)
	exposures, exposuresErr := encapsulation.Encapsulation(
		root,
		files,
	)
	pairings, pairingsErr := testpairing.Testpairing(root)
	crowded, crowdedErr := maxfilesperdir.Maxfilesperdir(root)
	surfaces, surfacesErr := publicapi.Publicapi(
		root,
		files,
	)
	reaches, reachesErr := reach.Reach(
		root,
		files,
	)
	names, namesErr := packageset.Packageset(
		root,
		files,
	)
	err = errors.Join(
		siblingsErr,
		componentsErr,
		boundariesErr,
		exposuresErr,
		pairingsErr,
		crowdedErr,
		surfacesErr,
		reachesErr,
		namesErr,
	)
	if err != nil {
		return nil,
			err
	}

	return slices.Concat(
			siblings,
			components,
			boundaries,
			exposures,
			pairings,
			crowded,
			surfaces,
			reaches,
			names,
		),
		nil
}
