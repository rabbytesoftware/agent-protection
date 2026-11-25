package gofmtcheck_test

import (
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/gofmtcheck"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestGofmtcheck(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "unindented body is reported at line one",
			code: "package p\n\nfunc f() {\nfor {\n}\n}\n",
			want: []string{"R62 p.go:1"},
		},
		{
			name: "spacing gofmt would change is reported",
			code: "// header\npackage p\n\nvar x   = 1\n",
			want: []string{"R62 p.go:1"},
		},
		{
			name: "formatted file passes",
			code: "package p\n\nfunc f() {\n\tfor {\n\t}\n}\n",
			want: nil,
		},
		{
			name: "unparseable content is left to the compiler",
			code: "package p\n\nfunc f( {\n",
			want: nil,
		},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			gofmtcheck.Gofmtcheck([]source.File{{Path: "p.go", Rel: "p.go", Content: []byte(tc.code)}}),
			tc.want,
		)
	}
}

func assertFindings(
	t *testing.T,
	name string,
	findings []report.Finding,
	want []string,
) {
	t.Helper()
	got := []string{}
	for _, finding := range findings {
		got = append(
			got,
			finding.Rule+" "+finding.Location,
		)
	}

	isMismatch := !slices.Equal(
		got,
		want,
	) && len(got)+len(want) > 0
	if isMismatch {
		t.Errorf(
			"%s: findings = %v, want %v",
			name,
			got,
			want,
		)
	}
}
