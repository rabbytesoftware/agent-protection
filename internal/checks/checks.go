package checks

import (
	"github.com/rabbytesoftware/agent-protection/internal/checks/internal/adapter"
	"github.com/rabbytesoftware/agent-protection/internal/checks/internal/pipeline"
	"github.com/rabbytesoftware/agent-protection/internal/report"
)

// Checker runs every registered check for one language against root and
// returns the findings. This package is the composition root for the
// goast/rustast/tsast sibling packages: it's the one place that imports
// and wires all three together, exposing them to callers only through this
// interface -- never through an exported struct. The actual per-language
// orchestration is unexported implementation in internal/pipeline, per
// rule 15 -- everything beyond the public contract lives in the package's
// own internal/ subtree, not loose beside this file.
type Checker interface {
	Check(root string) ([]report.Finding, error)
}

// GoChecker returns the Checker for Go repositories.
func GoChecker() Checker { return adapter.Func(pipeline.CheckGo) }

// RustChecker returns the Checker for Rust repositories.
func RustChecker() Checker { return adapter.Func(pipeline.CheckRust) }

// TypeScriptChecker returns the Checker for TypeScript repositories.
func TypeScriptChecker() Checker { return adapter.Func(pipeline.CheckTypeScript) }
