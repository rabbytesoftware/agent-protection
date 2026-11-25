package argumentperline_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/argumentperline"
)

func TestLiteralFindings(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "empty and single-element literals stay inline",
			code: "package p\n\nvar a = T{}\nvar b = T{X: g(1)}\nvar c = []int{h()}\n",
			want: nil,
		},
		{
			name: "one element per line with the brace on its own line passes",
			code: "package p\n\nvar a = T{\n\tX: g(),\n\tY: h(),\n}\n",
			want: nil,
		},
		{
			name: "up to three plain identifiers or basic literals stay inline",
			code: "package p\n\nvar a = []string{\"a\", \"b\", \"c\"}\nvar b = Point{X: 1, Y: y}\nvar c = [2]int{one, 2}\n",
			want: nil,
		},
		{
			name: "four plain elements on one line are reported",
			code: "package p\n\nvar a = []string{\"a\", \"b\", \"c\", \"d\"}\n",
			want: []string{"R52 p.go:3"},
		},
		{
			name: "inline element with a call is reported",
			code: "package p\n\nvar a = T{X: g(), Y: 1}\n",
			want: []string{"R52 p.go:3"},
		},
		{
			name: "inline selector element is reported",
			code: "package p\n\nvar a = []E{pkg.A, pkg.B}\n",
			want: []string{"R52 p.go:3"},
		},
		{
			name: "inline nested literal is reported on the outer literal only",
			code: "package p\n\nvar a = []T{{X: 1}, {X: 2}}\n",
			want: []string{"R52 p.go:3"},
		},
		{
			name: "fields packed onto fewer lines are reported",
			code: "package p\n\nfunc f() (queue Queue) {\n\treturn Queue{Handlers: handlers, Jobs: jobs,\n\t\tStore: store}\n}\n",
			want: []string{"R52 p.go:4"},
		},
		{
			name: "closing brace on the last element line is reported",
			code: "package p\n\nvar a = T{\n\tX: 1,\n\tY: 2}\n",
			want: []string{"R52 p.go:3"},
		},
		{
			name: "first element on the opening brace line is reported",
			code: "package p\n\nvar a = T{X: 1,\n\tY: 2,\n}\n",
			want: []string{"R52 p.go:3"},
		},
		{
			name: "multi-line plain literal is held to one element per line",
			code: "package p\n\nvar a = []int{1, 2,\n\t3}\n",
			want: []string{"R52 p.go:3"},
		},
		{
			name: "address of an inline struct is reported",
			code: "package p\n\nvar a = &T{X: g(), Y: h()}\n",
			want: []string{"R52 p.go:3"},
		},
		{
			name: "table rows outside test files are reported",
			code: "package p\n\nvar a = []T{\n\t{X: g(), Y: 1},\n}\n",
			want: []string{"R52 p.go:4"},
		},
	}
	for _, tc := range cases {
		assertLocations(
			t,
			tc.name,
			argumentperline.Argumentperline([]source.File{parsed(
				t,
				tc.code,
			)}),
			tc.want,
		)
	}
}
