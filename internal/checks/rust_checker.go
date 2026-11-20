package checks

import (
	"github.com/rabbytesoftware/agent-protection/internal/checks/rustast"
	"github.com/rabbytesoftware/agent-protection/internal/report"
)

type rustChecker struct{}

func (rustChecker) Check(root string) ([]report.Finding, error) {
	return rustast.RustClippy(root), nil
}
