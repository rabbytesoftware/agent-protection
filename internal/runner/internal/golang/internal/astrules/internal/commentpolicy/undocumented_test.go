package commentpolicy_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/commentpolicy"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestCommentpolicyRequiresExportedDocs(
	t *testing.T,
) {
	cases := []struct {
		name string
		path string
		code string
		want []string
	}{
		{
			name: "exported func without doc is reported",
			path: "p.go",
			code: `package p

func F() {
}
`,
			want: []string{"R07 p.go:3"},
		},
		{
			name: "exported func documented only by a directive is reported",
			path: "p.go",
			code: `package p

//go:noinline
func F() {
}
`,
			want: []string{"R07 p.go:4"},
		},
		{
			name: "exported var without doc is reported",
			path: "p.go",
			code: `package p

var V = 1
`,
			want: []string{"R07 p.go:3"},
		},
		{
			name: "test functions need no doc",
			path: "p_test.go",
			code: `package p_test

import "testing"

func TestF(
	t *testing.T,
) {
}

func BenchmarkF(
	b *testing.B,
) {
}

func FuzzF(
	f *testing.F,
) {
}

func TestMain(
	m *testing.M,
) {
}

func ExampleF() {
}
`,
			want: nil,
		},
		{
			name: "exported helper in a test file needs a doc",
			path: "p_test.go",
			code: `package p_test

import "testing"

func Helper(
	t *testing.T,
) {
}
`,
			want: []string{"R07 p_test.go:5"},
		},
		{
			name: "test-shaped function outside a test file needs a doc",
			path: "p.go",
			code: `package p

import "testing"

func TestF(
	t *testing.T,
) {
}
`,
			want: []string{"R07 p.go:5"},
		},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			commentpolicy.Commentpolicy([]source.File{parsed(
				t,
				tc.path,
				tc.code,
			)}),
			tc.want,
		)
	}
}
