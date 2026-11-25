package realinfra_test

import (
	"go/parser"
	"go/token"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/behavior/internal/realinfra"
)

func TestNoinfraintests(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "outbound http call is reported at the call",
			code: "package p\n\nimport (\n\t\"net/http\"\n\t\"testing\"\n)\n\nfunc TestF(\n\tt *testing.T,\n) {\n\thttp.Get(\"https://example.com\")\n}\n",
			want: []string{"p_test.go:11"},
		},
		{
			name: "real database open is reported",
			code: "package p\n\nimport (\n\t\"database/sql\"\n\t\"testing\"\n)\n\nfunc TestF(\n\tt *testing.T,\n) {\n\tsql.Open(\"postgres\", \"host=db.internal\")\n}\n",
			want: []string{"p_test.go:11"},
		},
		{
			name: "dial to an address taken from a loopback listener passes",
			code: "package p\n\nimport (\n\t\"net\"\n\t\"testing\"\n)\n\nfunc TestF(\n\tt *testing.T,\n) {\n\tlistener, err := net.Listen(\"tcp\", \"127.0.0.1:0\")\n\tif err != nil {\n\t\tt.Fatal(err)\n\t}\n\taddress := listener.Addr().String()\n\tnet.Dial(\"tcp\", address)\n}\n",
			want: nil,
		},
		{
			name: "loopback url built by concatenation passes",
			code: "package p\n\nimport (\n\t\"net/http\"\n\t\"testing\"\n)\n\nfunc TestF(\n\tt *testing.T,\n) {\n\thttp.Get(\"http://\" + \"127.0.0.1\" + \":8080\")\n}\n",
			want: nil,
		},
		{
			name: "test server url selector passes",
			code: "package p\n\nimport (\n\t\"net/http\"\n\t\"testing\"\n)\n\nfunc TestF(\n\tt *testing.T,\n) {\n\thttp.Get(server.URL)\n}\n",
			want: nil,
		},
		{
			name: "plain calls, field assignments and bodiless functions pass",
			code: "package p\n\nimport \"testing\"\n\nfunc external()\n\nfunc TestF(\n\tt *testing.T,\n) {\n\tcount := 1\n\tconfig.address = \"127.0.0.1:1\"\n\thelper(count)\n}\n",
			want: nil,
		},
		{
			name: "building an in-memory request passes",
			code: "package p\n\nimport (\n\t\"net/http\"\n\t\"testing\"\n)\n\nfunc TestF(\n\tt *testing.T,\n) {\n\thttp.NewRequest(\"GET\", \"/\", nil)\n}\n",
			want: nil,
		},
	}
	for _, tc := range cases {
		assertLocations(
			t,
			tc.name,
			realinfra.Realinfra([]source.File{parsed(
				t,
				tc.code,
			)}),
			tc.want,
		)
	}
}

func parsed(
	t *testing.T,
	code string,
) (file source.File) {
	t.Helper()
	fileSet := token.NewFileSet()
	syntax, err := parser.ParseFile(
		fileSet,
		"p_test.go",
		code,
		parser.ParseComments,
	)
	if err != nil {
		t.Fatalf(
			"parse fixture: %v",
			err,
		)
	}

	return source.File{
		Path:    "p_test.go",
		Rel:     "p_test.go",
		FileSet: fileSet,
		Syntax:  syntax,
	}
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
