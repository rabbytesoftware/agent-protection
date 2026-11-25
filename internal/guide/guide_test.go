package guide_test

import (
	"strings"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/guide"
)

func TestGuide(
	t *testing.T,
) {
	missing := []string{}
	for _, heading := range []string{"NO COMMENTS", "THE MODULE MODEL", "container.go", "HOW TO WORK"} {
		missing = appendMissing(
			missing,
			heading,
		)
	}

	if len(missing) > 0 {
		t.Errorf(
			"Guide() lacks %v",
			missing,
		)
	}
}

func appendMissing(
	missing []string,
	heading string,
) (extended []string) {
	if strings.Contains(
		guide.Guide(),
		heading,
	) {
		return missing
	}

	return append(
		missing,
		heading,
	)
}
