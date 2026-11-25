package realinfra_test

import (
	"go/parser"
	"go/token"
	"maps"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/behavior/internal/realinfra"
)

type fixture struct {
	name  string
	files map[string]string
	want  []string
}

func TestRealinfra(
	t *testing.T,
) {
	assertCases(
		t,
		[]fixture{
			{
				name: "a package-level http call to a remote host is reported",
				files: map[string]string{
					"p/p_test.go": "package p\n\nimport (\n\t\"net/http\"\n\t\"testing\"\n)\n\nfunc TestF(\n\tt *testing.T,\n) {\n\thttp.Get(\"https://example.com\")\n}\n",
				},
				want: []string{
					"p/p_test.go:11",
				},
			},
			{
				name: "a loopback string literal is still live infrastructure",
				files: map[string]string{
					"p/p_test.go": "package p\n\nimport (\n\t\"net/http\"\n\t\"testing\"\n)\n\nfunc TestF(\n\tt *testing.T,\n) {\n\thttp.Get(\"http://localhost:8080/health\")\n\thttp.Post(\"http://\" + \"127.0.0.1\" + \":8080\", \"text/plain\", nil)\n}\n",
				},
				want: []string{
					"p/p_test.go:11",
					"p/p_test.go:12",
				},
			},
			{
				name: "an http call aimed at a test server created in the test passes",
				files: map[string]string{
					"p/p_test.go": "package p\n\nimport (\n\t\"net/http\"\n\t\"net/http/httptest\"\n\t\"testing\"\n)\n\nfunc TestF(\n\tt *testing.T,\n) {\n\tsrv := httptest.NewServer(nil)\n\thttp.Get(srv.URL + \"/health\")\n\tsrv.Client().Get(srv.URL)\n}\n",
				},
			},
			{
				name: "a database open and a dial are reported",
				files: map[string]string{
					"p/p_test.go": "package p\n\nimport (\n\t\"crypto/tls\"\n\t\"database/sql\"\n\t\"net\"\n\t\"testing\"\n)\n\nfunc TestF(\n\tt *testing.T,\n) {\n\tsql.Open(\"postgres\", \"host=db.internal\")\n\tsql.OpenDB(nil)\n\tnet.DialTimeout(\"tcp\", \"db.example.com:5432\", 0)\n\ttls.Dial(\"tcp\", \"localhost:443\", nil)\n}\n",
				},
				want: []string{
					"p/p_test.go:13",
					"p/p_test.go:14",
					"p/p_test.go:15",
					"p/p_test.go:16",
				},
			},
			{
				name: "a selector on an undeclared server is not trusted",
				files: map[string]string{
					"p/p_test.go": "package p\n\nimport (\n\t\"net/http\"\n\t\"testing\"\n)\n\nfunc TestF(\n\tt *testing.T,\n) {\n\thttp.Get(server.URL)\n}\n",
				},
				want: []string{
					"p/p_test.go:11",
				},
			},
			{
				name: "pure net and http helpers, plain calls and bodiless functions pass",
				files: map[string]string{
					"p/p_test.go": "package p\n\nimport (\n\t\"net\"\n\t\"net/http\"\n\t\"testing\"\n)\n\nfunc external()\n\nfunc TestF(\n\tt *testing.T,\n) {\n\tnet.ParseIP(\"10.0.0.1\")\n\tnet.JoinHostPort(\"a\", \"1\")\n\thttp.NewRequest(\"GET\", \"/\", nil)\n\thelper(1)\n\tfunc() {}()\n}\n",
				},
			},
		},
	)
}

func TestRealinfraSnippet(
	t *testing.T,
) {
	findings := realinfra.Realinfra([]source.File{parsed(
		t,
		"p/p_test.go",
		"package p\n\nimport \"net/http\"\n\nfunc f() {\n\thttp.Get(\"https://example.com\")\n}\n",
	)})
	want := "unit test makes a live call to http.Get against a target it does not own: inject an in-memory fake through the interface the code depends on, or aim the call at an httptest.Server (srv.URL, srv.Client()) or a listener the test opens on 127.0.0.1:0 and reads back with Addr(); a localhost or host string literal is still live infrastructure"
	isMismatch := len(findings) != 1 || findings[0].Snippet != want || findings[0].Rule != "R41"
	if isMismatch {
		t.Errorf(
			"findings = %v, want one R41 finding with snippet %q",
			findings,
			want,
		)
	}
}

func assertCases(
	t *testing.T,
	cases []fixture,
) {
	t.Helper()
	for _, tc := range cases {
		assertLocations(
			t,
			tc,
			checked(
				t,
				tc.files,
			),
		)
	}
}

func checked(
	t *testing.T,
	contents map[string]string,
) (findings []report.Finding) {
	t.Helper()
	tests := []source.File{}
	for _, rel := range slices.Sorted(maps.Keys(contents)) {
		tests = append(
			tests,
			parsed(
				t,
				rel,
				contents[rel],
			),
		)
	}

	return realinfra.Realinfra(tests)
}

func parsed(
	t *testing.T,
	rel string,
	code string,
) (file source.File) {
	t.Helper()
	fileSet := token.NewFileSet()
	syntax, err := parser.ParseFile(
		fileSet,
		rel,
		code,
		parser.ParseComments,
	)
	if err != nil {
		t.Fatalf(
			"parse %s: %v",
			rel,
			err,
		)
	}

	return source.File{
		Path:    rel,
		Rel:     rel,
		FileSet: fileSet,
		Syntax:  syntax,
	}
}

func assertLocations(
	t *testing.T,
	tc fixture,
	findings []report.Finding,
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
		tc.want,
	) && len(got)+len(tc.want) > 0
	if isMismatch {
		t.Errorf(
			"%s: locations = %v, want %v",
			tc.name,
			got,
			tc.want,
		)
	}
}
