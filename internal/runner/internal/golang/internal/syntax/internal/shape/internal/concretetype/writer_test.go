package concretetype_test

import (
	"strings"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/shape/internal/concretetype"
)

func TestConcretetypeWriter(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{name: "value receiver incrementing a map entry is reported", code: "package p\n\ntype Table struct {\n\tseen map[int]int\n}\n\nfunc (t Table) Record(\n\tv int,\n) {\n\tt.seen[v]++\n}\n", want: []string{"R13 p/a.go:3"}},
		{name: "value receiver assigning a slice element is reported", code: "package p\n\ntype Ring struct {\n\titems []int\n}\n\nfunc (r Ring) Put(\n\tv int,\n) {\n\tr.items[0] = v\n}\n", want: []string{"R13 p/a.go:3"}},
		{name: "value receiver adding to a map entry is reported", code: "package p\n\ntype Tally struct {\n\tsums map[string]int\n}\n\nfunc (t Tally) Add(\n\tk string,\n) {\n\tt.sums[k] += 2\n}\n", want: []string{"R13 p/a.go:3"}},
		{name: "value receiver sending on a channel field is reported", code: "package p\n\ntype Bus struct {\n\tevents chan int\n}\n\nfunc (b Bus) Publish(\n\tv int,\n) {\n\tb.events <- v\n}\n", want: []string{"R13 p/a.go:3"}},
		{name: "value receiver storing through a pointer field is reported", code: "package p\n\ntype Counter struct {\n\tn *int\n}\n\nfunc (c Counter) Bump() {\n\t*c.n = 1\n}\n", want: []string{"R13 p/a.go:3"}},
		{name: "value receiver setting a field of a pointer field is reported", code: "package p\n\ntype state struct {\n\tn int\n}\n\ntype Counter struct {\n\t*state\n}\n\nfunc (c Counter) Bump() {\n\tc.state.n = 1\n}\n", want: []string{"R13 p/a.go:7"}},
		{name: "value receiver deleting from a map field is reported", code: "package p\n\ntype Set struct {\n\tm map[int]bool\n}\n\nfunc (s Set) Drop(\n\tv int,\n) {\n\tdelete(s.m, v)\n}\n", want: []string{"R13 p/a.go:3"}},
		{name: "value receiver clearing a map field is reported", code: "package p\n\ntype Set struct {\n\tm map[int]bool\n}\n\nfunc (s Set) Reset() {\n\tclear(s.m)\n}\n", want: []string{"R13 p/a.go:3"}},
		{name: "value receiver reading a map field passes", code: "package p\n\ntype Table struct {\n\tseen map[int]int\n}\n\nfunc (t Table) Seen(\n\tv int,\n) int {\n\treturn t.seen[v]\n}\n", want: nil},
		{name: "value receiver assigning a plain field passes", code: "package p\n\ntype Point struct {\n\tx int\n\ttags []string\n}\n\nfunc (p Point) Moved() Point {\n\tp.x = 1\n\tp.tags = nil\n\treturn p\n}\n", want: nil},
		{name: "value receiver writing a fixed array element passes", code: "package p\n\ntype Grid struct {\n\tcells [4]int\n}\n\nfunc (g Grid) With() Grid {\n\tg.cells[0] = 1\n\treturn g\n}\n", want: nil},
		{name: "value receiver writing a local map passes", code: "package p\n\ntype Table struct {\n\tseen map[int]int\n}\n\nfunc (t Table) Copy() map[int]int {\n\tout := map[int]int{}\n\tout[1]++\n\treturn out\n}\n", want: nil},
		{name: "unexported struct writing through a map field passes", code: "package p\n\ntype table struct {\n\tseen map[int]int\n}\n\nfunc (t table) Record() {\n\tt.seen[1]++\n}\n", want: nil},
		{name: "error type writing through a pointer field passes", code: "package p\n\ntype Failure struct {\n\tn *int\n}\n\nfunc (f Failure) Error() string {\n\t*f.n = 1\n\treturn \"\"\n}\n", want: nil},
		{name: "unnamed value receiver passes", code: "package p\n\ntype Table struct {\n\tseen map[int]int\n}\n\nfunc (Table) Record() {\n\tdelete(seen, 1)\n}\n", want: nil},
	}
	for _, tc := range cases {
		findings := concretetype.Concretetype(parsedFiles(
			t,
			map[string]string{"p/a.go": tc.code},
		))
		assertFindings(
			t,
			tc.name,
			findings,
			tc.want,
		)
		assertWriterSnippet(
			t,
			tc.name,
			findings,
		)
	}
}

func assertWriterSnippet(
	t *testing.T,
	name string,
	findings []report.Finding,
) {
	t.Helper()
	isPrescriptive := len(findings) == 0 || strings.Contains(
		findings[0].Snippet,
		"writes through its field",
	)
	if !isPrescriptive {
		t.Errorf(
			"%s: snippet = %q, want the written field named",
			name,
			findings[0].Snippet,
		)
	}
}
