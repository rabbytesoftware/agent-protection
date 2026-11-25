package command

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/guide"
)

func TestManual(
	t *testing.T,
) {
	assertCode(
		t,
		"Manual().Run()",
		Manual(guide.Guide).Run(""),
		cleanExitCode,
	)
}
