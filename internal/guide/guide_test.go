package guide_test

import (
	"strings"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/guide"
)

func TestGuide(
	t *testing.T,
) {
	for _, heading := range []string{"NO COMMENTS", "THE MODULE MODEL", "container.go", "HOW TO WORK"} {
		if !strings.Contains(
			guide.Guide(),
			heading,
		) {
			t.Errorf(
				"Guide() lacks %q",
				heading,
			)
		}
	}
}
