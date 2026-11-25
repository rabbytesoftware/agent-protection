package rust

import (
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan"
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

	commented, err := commentscan.Comments(
		root,
		commentscan.RustExtensions(),
	)
	if err != nil {
		return nil,
			err
	}

	return slices.Concat(
			production,
			commented,
		),
		nil
}
