package paramperline

import (
	"go/token"
	"testing"
)

func TestReach(
	t *testing.T,
) {
	span := reach{
		start: token.Pos(3),
		end:   token.Pos(9),
	}
	isExact := span.Pos() == token.Pos(3) && span.End() == token.Pos(9)
	if !isExact {
		t.Errorf(
			"reach = %d..%d, want 3..9",
			span.Pos(),
			span.End(),
		)
	}
}
