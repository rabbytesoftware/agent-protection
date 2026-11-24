package typescript

import (
	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/typescript/internal/manifest"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/typescript/internal/passthrough"
)

// Typescript checks the TypeScript component at root twice over: once
// with the repo's own eslint config and once with Agent Protection's own
// rule set, and returns both sets of findings. It fails as soon as either
// eslint run fails.
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
	return append(
			own,
			manifested...,
		),
		nil
}
