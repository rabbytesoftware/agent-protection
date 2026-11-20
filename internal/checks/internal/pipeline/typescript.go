package pipeline

import (
	"fmt"
	"os"

	"github.com/rabbytesoftware/agent-protection/internal/checks/tsast"
	"github.com/rabbytesoftware/agent-protection/internal/report"
)

// CheckTypeScript runs both the repo's own ESLint config (pass-through) and
// AP's own universal TS manifest.
func CheckTypeScript(root string) ([]report.Finding, error) {
	all := tsast.TSESLint(root)
	manifestFindings, merr := tsast.TSManifest(root)
	if merr != nil {
		fmt.Fprintf(os.Stderr, "ts manifest: %v\n", merr)
	}
	return append(all, manifestFindings...), nil
}
