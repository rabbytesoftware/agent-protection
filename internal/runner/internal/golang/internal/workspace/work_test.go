package workspace_test

import (
	"os"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/workspace"
)

func TestWorkspaceRejectsUnreadableGoWork(
	t *testing.T,
) {
	root := moduleWith(
		t,
		map[string]string{"go.work/x": "x\n"},
	)
	_, err := workspace.Workspace(
		root,
		os.Getenv,
		located(root),
	)
	assertFailed(
		t,
		"an unreadable go.work is an error",
		err,
	)
}
