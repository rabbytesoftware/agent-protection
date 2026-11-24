package droppederror_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/droppederror"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestDroppederrorCalleesSpanTheDirectory(
	t *testing.T,
) {
	cases := []struct {
		name       string
		callerPath string
		want       []string
	}{
		{name: "erroring function declared in a sibling file is known", callerPath: "pkg/b.go", want: []string{"R26 pkg/b.go:5"}},
		{name: "erroring function declared in another directory is unknown", callerPath: "other/b.go", want: nil},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			droppederror.Droppederror([]source.File{
				parsedAt(
					t,
					"pkg/a.go",
					"package p\n\nfunc save() (failure error)\n",
				),
				parsedAt(
					t,
					tc.callerPath,
					"package p\n\nfunc f() {\n\tproblem := save()\n\tif problem != nil {\n\t\treturn\n\t}\n}\n",
				),
			}),
			tc.want,
		)
	}
}
