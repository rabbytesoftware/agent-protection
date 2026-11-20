package arch

import (
	"errors"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/compositionroot"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/encapsulation"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/internalboundary"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/maxfilesperdir"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/siblingimports"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/testpairing"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

// Arch runs every architecture and package layout check over the module at
// root: sibling imports, composition roots, internal/ boundaries,
// single-consumer encapsulation, test pairing and files per directory.
// Every check runs; if any fails, their errors are joined and returned.
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
	err = errors.Join(
		siblingsErr,
		componentsErr,
		boundariesErr,
		exposuresErr,
		pairingsErr,
		crowdedErr,
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
		),
		nil
}
