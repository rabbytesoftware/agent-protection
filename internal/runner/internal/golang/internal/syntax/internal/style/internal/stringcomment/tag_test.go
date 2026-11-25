package stringcomment_test

import (
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/style/internal/stringcomment"
)

func TestStringcommentTags(
	t *testing.T,
) {
	cases := []struct {
		name string
		path string
		code string
		want []string
	}{
		{name: "unknown tag key is reported", path: "p.go", code: "package p\n\ntype T struct {\n\tA int `why:\"keep\"`\n}\n", want: []string{"R07 p.go:4"}},
		{name: "known tag keys pass", path: "p.go", code: "package p\n\ntype T struct {\n\tA int `json:\"a,omitempty\" yaml:\"a\" validate:\"required\"`\n}\n", want: nil},
		{name: "unknown key beside known keys is reported", path: "p.go", code: "package p\n\ntype T struct {\n\tA int `json:\"a\" note:\"x\"`\n}\n", want: []string{"R07 p.go:4"}},
		{name: "every unknown key is reported", path: "p.go", code: "package p\n\ntype T struct {\n\tA int `a:\"1\" b:\"2\"`\n}\n", want: []string{"R07 p.go:4", "R07 p.go:4"}},
		{name: "prose tag is reported", path: "p.go", code: "package p\n\ntype T struct {\n\tA int `keep this field`\n}\n", want: []string{"R07 p.go:4"}},
		{name: "interpreted string tag passes", path: "p.go", code: "package p\n\ntype T struct {\n\tA int \"json:\\\"a\\\"\"\n}\n", want: nil},
		{name: "escaped quote in a value passes", path: "p.go", code: "package p\n\ntype T struct {\n\tA int `json:\"a\\\"b\"`\n}\n", want: nil},
		{name: "empty tag passes", path: "p.go", code: "package p\n\ntype T struct {\n\tA int ``\n}\n", want: nil},
		{name: "tag on an anonymous struct is reported", path: "p.go", code: "package p\n\nfunc F() {\n\tv := struct {\n\t\tA int `why:\"x\"`\n\t}{}\n\tuse(v)\n}\n", want: []string{"R07 p.go:5"}},
		{name: "red team struct tag is reported", path: "p.go", code: "package square\n\ntype square struct {\n\tfactor int `why:\"multiplier kept for the old invoices, do not touch\"`\n}\n", want: []string{"R07 p.go:4"}},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			stringcomment.Stringcomment([]source.File{parsed(
				t,
				tc.path,
				tc.code,
			)}),
			tc.want,
		)
	}
}

func TestStringcommentTagSnippets(
	t *testing.T,
) {
	findings := stringcomment.Stringcomment([]source.File{parsed(
		t,
		"p.go",
		"package p\n\ntype T struct {\n\tA int `why:\"x\"`\n\tB int `keep this`\n}\n",
	)})
	want := []report.Finding{
		{Rule: "R07", Location: "p.go:4", Snippet: "struct tag key \"why\" is not read by any known library: a tag carries machine metadata, not explanations; delete it and make the code say it through names and structure"},
		{Rule: "R07", Location: "p.go:5", Snippet: "struct tag is not a list of key:\"value\" pairs: a tag carries machine metadata, not explanations; delete the prose"},
	}
	if !slices.Equal(
		findings,
		want,
	) {
		t.Errorf(
			"Stringcomment() = %v, want %v",
			findings,
			want,
		)
	}
}
