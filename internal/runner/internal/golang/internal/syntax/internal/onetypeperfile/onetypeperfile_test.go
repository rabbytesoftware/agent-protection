package onetypeperfile_test

import (
	"go/parser"
	"go/token"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/onetypeperfile"
)

func TestOnetypeperfile(
	t *testing.T,
) {
	cases := []struct {
		name string
		path string
		code string
		want []string
	}{
		{
			name: "two types in one file are reported at the second type",
			path: "a.go",
			code: "package p\n\ntype A struct{}\n\ntype B struct{}\n",
			want: []string{"a.go:5"},
		},
		{
			name: "grouped type declaration counts every type",
			path: "a.go",
			code: "package p\n\ntype (\n\tA int\n\tB int\n)\n",
			want: []string{"a.go:5"},
		},
		{
			name: "single type in a file named after another word is reported",
			path: "types.go",
			code: "package p\n\ntype Walker struct{}\n",
			want: []string{"types.go:3"},
		},
		{
			name: "single type in its snake case file passes",
			path: "parsed_file.go",
			code: "package p\n\ntype ParsedFile struct{}\n",
			want: nil,
		},
		{
			name: "leading acronym splits before the next word",
			path: "http_server.go",
			code: "package p\n\ntype HTTPServer struct{}\n",
			want: nil,
		},
		{
			name: "single type in the file named after the package is reported",
			path: "p.go",
			code: "package p\n\ntype Output struct{}\n",
			want: []string{"p.go:3"},
		},
		{
			name: "single type in container.go passes",
			path: "container.go",
			code: "package p\n\ntype aggregatePool struct{}\n",
			want: nil,
		},
		{
			name: "single type in wire.go is reported",
			path: "wire.go",
			code: "package p\n\ntype aggregatePool struct{}\n",
			want: []string{"wire.go:3"},
		},
		{
			name: "type local to a function body is reported",
			path: "walker.go",
			code: "package p\n\ntype walker struct{}\n\nfunc f() {\n\ttype local int\n}\n",
			want: []string{"walker.go:6"},
		},
		{
			name: "local struct type is reported once at its declaration",
			path: "helpers.go",
			code: "package p\n\nfunc f() {\n\ttype entry struct {\n\t\tkey int\n\t}\n\t_ = entry{}\n}\n",
			want: []string{"helpers.go:4"},
		},
		{
			name: "anonymous struct in a composite literal is reported",
			path: "helpers.go",
			code: "package p\n\nfunc f() {\n\trows := []struct {\n\t\tlimit int\n\t}{{limit: 1}}\n\t_ = rows\n}\n",
			want: []string{"helpers.go:4"},
		},
		{
			name: "anonymous struct in a package var is reported",
			path: "helpers.go",
			code: "package p\n\nvar settings struct {\n\tlimit int\n}\n",
			want: []string{"helpers.go:3"},
		},
		{
			name: "anonymous struct in a local var is reported",
			path: "helpers.go",
			code: "package p\n\nfunc f() {\n\tvar payload struct {\n\t\tName string\n\t}\n\t_ = payload\n}\n",
			want: []string{"helpers.go:4"},
		},
		{
			name: "empty struct as a set value or signal passes",
			path: "helpers.go",
			code: "package p\n\nfunc f() {\n\tseen := map[string]struct{}{}\n\tdone := make(chan struct{})\n\tseen[\"a\"] = struct{}{}\n\t_ = done\n}\n",
			want: nil,
		},
		{
			name: "nested struct field inside a named type passes",
			path: "config.go",
			code: "package p\n\ntype config struct {\n\tdb struct {\n\t\thost string\n\t}\n}\n",
			want: nil,
		},
		{
			name: "file without types passes",
			path: "helpers.go",
			code: "package p\n\nfunc f() {}\n",
			want: nil,
		},
	}
	for _, tc := range cases {
		assertLocations(
			t,
			tc.name,
			onetypeperfile.Onetypeperfile([]source.File{parsed(
				t,
				tc.path,
				tc.code,
			)}),
			tc.want,
		)
	}
}

func parsed(
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
