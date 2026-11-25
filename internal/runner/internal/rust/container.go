package rust

import (
	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/comments"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/rust/internal/clippy"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/rust/internal/testmodule"
)

func Rust(
	root string,
) (findings []report.Finding, err error) {
	warnings, err := clippy.Clippy(root)
	if err != nil {
		return nil,
			err
	}

	production, err := testmodule.Production(
		root,
		warnings,
	)
	if err != nil {
		return nil,
			err
	}

	commented, err := comments.Comments(
		root,
		comments.RustExtensions(),
	)
	if err != nil {
		return nil,
			err
	}

	return append(
			production,
			commented...,
		),
		nil
}
