package pipeline

import (
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/checks/internal/rustast"
	"github.com/rabbytesoftware/agent-protection/internal/report"
)

// CheckRust runs cargo clippy against root and maps its output to findings.
func CheckRust(root string) ([]report.Finding, error) {
	return rustast.RustClippy(root), nil
}
