package checks

import (
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/checks/internal"
	"github.com/rabbytesoftware/agent-protection/internal/report"
)

// Checker runs every registered check for one language against root and
// returns the findings. This package is the composition root for the
// per-language implementation: it's the one place that imports and wires
// the whole pipeline together, exposing it to callers only through this
// interface -- never through an exported struct. Everything beyond this
// contract lives in this package's own internal/ subtree (rule 15), with
// internal/ nested as deep as each piece actually needs (rule 19) --
// goast/rustast/tsast are visible only within internal/checks/, never to
// the rest of the module. The directory is named internal/ for Go's own
// visibility enforcement; the package itself is declared "pipeline".
type Checker interface {
	Check(root string) ([]report.Finding, error)
}

// GoChecker returns the Checker for Go repositories.
func GoChecker() Checker { return pipeline.Func(pipeline.CheckGo) }

// RustChecker returns the Checker for Rust repositories.
func RustChecker() Checker { return pipeline.Func(pipeline.CheckRust) }

// TypeScriptChecker returns the Checker for TypeScript repositories.
func TypeScriptChecker() Checker { return pipeline.Func(pipeline.CheckTypeScript) }
