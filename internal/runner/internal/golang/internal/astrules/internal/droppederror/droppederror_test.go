package droppederror_test

import (
	"go/parser"
	"go/token"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/droppederror"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestDroppederror(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{name: "best-effort save returning on a failed write is reported", code: "package p\n\nfunc save() error\n\nfunc saveBestEffort() {\n\terr := save()\n\tif err != nil {\n\t\treturn\n\t}\n}\n", want: []string{"R26 p.go:7"}},
		{name: "unsubscribe error checked then silently returned is reported", code: "package p\n\nfunc unsubscribe() {\n\tif err := bus.Unsubscribe(id); err != nil {\n\t\treturn\n\t}\n}\n", want: []string{"R26 p.go:4"}},
		{name: "id generator returning an empty string on failure is reported", code: "package p\n\nfunc New() string {\n\tvar b [16]byte\n\tif _, err := rand.Read(b[:]); err != nil {\n\t\treturn \"\"\n\t}\n\n\treturn string(b[:])\n}\n", want: []string{"R26 p.go:5"}},
		{name: "panic swallowed through a blank assignment is reported", code: "package p\n\nfunc swallowPanic() {\n\trecovered := recover()\n\t_ = recovered\n}\n", want: []string{"R26 p.go:5"}},
		{name: "panic swallowed through a nil check is reported", code: "package p\n\nfunc swallowPanic() {\n\tif recover() != nil {\n\t\treturn\n\t}\n}\n", want: []string{"R26 p.go:4"}},
		{name: "panic swallowed through a named recovered value is reported", code: "package p\n\nfunc discardPanic() {\n\tif r := recover(); r != nil {\n\t\treturn\n\t}\n}\n", want: []string{"R26 p.go:4"}},
		{name: "error wrapped with %w passes", code: "package p\n\nfunc f() (err error) {\n\terr = g()\n\tif err != nil {\n\t\treturn fmt.Errorf(\"g: %w\", err)\n\t}\n\n\treturn nil\n}\n", want: nil},
		{name: "recovered panic handed to a handler passes", code: "package p\n\nfunc f() {\n\tif r := recover(); r != nil {\n\t\treport(r)\n\t\treturn\n\t}\n}\n", want: nil},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			droppederror.Droppederror([]source.File{parsed(
				t,
				tc.code,
			)}),
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
