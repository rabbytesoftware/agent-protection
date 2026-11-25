package typescript

import (
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/comments"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/typescript/internal/manifest"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/typescript/internal/passthrough"
)

func Typescript(
	root string,
) (findings []report.Finding, err error) {
	own, err := passthrough.Passthrough(root)
	if err != nil {
		return nil,
			err
	}

	manifested, err := manifest.Manifest(root)
	if err != nil {
		return nil,
			err
	}

	commented, err := comments.Comments(
		root,
		comments.ScriptExtensions(),
	)
	if err != nil {
		return nil,
			err
	}

	return slices.Concat(
			own,
			manifested,
			commented,
		),
		nil
}
