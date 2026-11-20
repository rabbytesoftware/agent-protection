package checks

import "github.com/rabbytesoftware/agent-protection/internal/report"

// Checker runs every registered check for one language against root and
// returns the findings. This package is the composition root for the
// goast/rustast/tsast sibling packages: it's the one place that imports
// and wires all three together, exposing them to callers only through this
// interface -- never through an exported struct.
type Checker interface {
	Check(root string) ([]report.Finding, error)
}

// GoChecker returns the Checker for Go repositories.
func GoChecker() Checker { return goChecker{} }

// RustChecker returns the Checker for Rust repositories.
func RustChecker() Checker { return rustChecker{} }

// TypeScriptChecker returns the Checker for TypeScript repositories.
func TypeScriptChecker() Checker { return tsChecker{} }
