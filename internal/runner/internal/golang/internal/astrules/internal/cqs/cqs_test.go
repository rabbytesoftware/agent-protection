package cqs_test

import (
	"go/parser"
	"go/token"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/cqs"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestCqs(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "returning a value while mutating the receiver is reported",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc (t *T) f() (v int) {\n\tt.v = 5\n\treturn t.v\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "returning a value while writing through a pointer parameter is reported",
			code: "package p\n\nfunc f(\n\ttarget *int,\n) (v int) {\n\t*target = 5\n\treturn 5\n}\n",
			want: []string{"p.go:3"},
		},
		{
			name: "pure query passes",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc (t *T) f() (v int) {\n\treturn t.v\n}\n",
			want: nil,
		},
		{
			name: "command returning only an error passes",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc (t *T) f() (err error) {\n\tt.v = 5\n\treturn nil\n}\n",
			want: nil,
		},
		{
			name: "assigning to a value receiver copy passes",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc (t T) f() (v int) {\n\tt.v = 5\n\treturn t.v\n}\n",
			want: nil,
		},
	}
	for _, tc := range cases {
		assertCase(
			t,
			tc.name,
			tc.code,
			tc.want,
		)
	}
}

func TestCqsSkipsMocks(
	t *testing.T,
) {
	code := "package mocks\n\ntype Store struct {\n\tGetCalls int\n}\n\nfunc (s *Store) Get() (v int) {\n\ts.GetCalls++\n\treturn 1\n}\n"
	cases := []struct {
		name string
		path string
		want []string
	}{
		{
			name: "call-counting mock passes",
			path: "p/mocks/store.go",
			want: nil,
		},
		{
			name: "the same code outside a mock is reported",
			path: "p/store.go",
			want: []string{"p/store.go:7"},
		},
	}
	for _, tc := range cases {
		assertLocations(
			t,
			tc.name,
			cqs.Cqs([]source.File{parsedAt(
				t,
				tc.path,
				code,
			)}),
			tc.want,
		)
	}
}

func TestCqsSnippet(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want string
	}{
		{
			name: "direct write names only the function",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc (t *T) Get() (v int) {\n\tt.v++\n\treturn t.v\n}\n",
			want: "func Get(...) returns a value and mutates its receiver or a pointer, map, slice or channel parameter: split the query from the command",
		},
		{
			name: "write through a helper names the first mutating call",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc (t *T) bump() {\n\tt.v++\n}\n\nfunc (t *T) Get() (v int) {\n\t_ = t.value()\n\tt.bump()\n\treturn t.v\n}\n\nfunc (t *T) value() (v int) {\n\treturn t.v\n}\n",
			want: "func Get(...) returns a value and mutates its receiver or a pointer, map, slice or channel parameter through t.bump(...): split the query from the command",
		},
	}
	for _, tc := range cases {
		assertSnippet(
			t,
			tc.name,
			cqs.Cqs([]source.File{parsed(
				t,
				tc.code,
			)}),
			tc.want,
		)
	}
}

func assertSnippet(
	t *testing.T,
	name string,
	findings []report.Finding,
	want string,
) {
	t.Helper()
	isMatch := len(findings) == 1 &&
		findings[0].Snippet == want
	if !isMatch {
		t.Errorf(
			"%s: findings = %v, want one with snippet %q",
			name,
			findings,
			want,
		)
	}
}

func parsed(
	t *testing.T,
	code string,
) (file source.File) {
	t.Helper()
	return parsedAt(
		t,
		"p.go",
		code,
	)
}

func parsedAt(
	t *testing.T,
	path string,
	code string,
) (file source.File) {
	t.Helper()
	fileSet := token.NewFileSet()
	syntax, err := parser.ParseFile(
		fileSet,
		path,
		code,
		parser.ParseComments,
	)
	if err != nil {
		t.Fatalf(
			"parse fixture: %v",
			err,
		)
	}

	return source.File{Path: path, Rel: path, FileSet: fileSet, Syntax: syntax}
}

func assertCase(
	t *testing.T,
	name string,
	code string,
	want []string,
) {
	t.Helper()
	assertLocations(
		t,
		name,
		cqs.Cqs([]source.File{parsed(
			t,
			code,
		)}),
		want,
	)
}

func assertLocations(
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
			finding.Location,
		)
	}

	isMismatch := !slices.Equal(
		got,
		want,
	) && len(got)+len(want) > 0
	if isMismatch {
		t.Errorf(
			"%s: locations = %v, want %v",
			name,
			got,
			want,
		)
	}
}
