package checks

import (
	"fmt"
	"os"

	"github.com/rabbytesoftware/agent-protection/internal/checks/tsast"
	"github.com/rabbytesoftware/agent-protection/internal/report"
)

type tsChecker struct{}

func (tsChecker) Check(root string) ([]report.Finding, error) {
	all := tsast.TSESLint(root)
	manifestFindings, merr := tsast.TSManifest(root)
	if merr != nil {
		fmt.Fprintf(os.Stderr, "ts manifest: %v\n", merr)
	}
	return append(all, manifestFindings...), nil
}
