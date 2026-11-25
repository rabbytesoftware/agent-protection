package rust

import (
	"github.com/rabbytesoftware/agent-protection/internal/report"
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

	return testmodule.Production(
		root,
		warnings,
	)
}
