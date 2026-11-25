package nopanic_test

import (
	"go/parser"
	"go/token"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/nopanic"
)

func TestNopanic(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "panic in a library package is reported at the call",
			code: "package p\n\nfunc f() {\n\tpanic(\"boom\")\n}\n",
			want: []string{"p.go:4"},
		},
		{
			name: "panic inside a function literal is reported",
			code: "package p\n\nvar f = func() {\n\tpanic(\"boom\")\n}\n",
			want: []string{"p.go:4"},
		},
		{
			name: "panic in package main passes",
			code: "package main\n\nfunc main() {\n\tpanic(\"boom\")\n}\n",
			want: nil,
		},
		{
			name: "log.Fatalf in a library package is reported",
			code: "package p\n\nimport \"log\"\n\nfunc f() {\n\tlog.Fatalf(\"boom\")\n}\n",
			want: []string{"p.go:6"},
		},
		{
			name: "log.Panicln, os.Exit and runtime.Goexit are reported",
			code: "package p\n\nimport (\n\t\"log\"\n\t\"os\"\n\t\"runtime\"\n)\n\nfunc f() {\n\tlog.Panicln(\"boom\")\n\tos.Exit(1)\n\truntime.Goexit()\n}\n",
			want: []string{"p.go:10", "p.go:11", "p.go:12"},
		},
		{
			name: "aliased os import is reported",
			code: "package p\n\nimport system \"os\"\n\nfunc f() {\n\tsystem.Exit(1)\n}\n",
			want: []string{"p.go:6"},
		},
		{
			name: "os.Exit in package main passes",
			code: "package main\n\nimport \"os\"\n\nfunc main() {\n\tos.Exit(1)\n}\n",
			want: nil,
		},
		{
			name: "non-terminating log and os calls pass",
			code: "package p\n\nimport (\n\t\"log\"\n\t\"os\"\n)\n\nfunc f() {\n\tlog.Printf(\"ok\")\n\t_ = os.Getenv(\"X\")\n}\n",
			want: nil,
		},
		{
			name: "Fatal on a package not imported as log passes",
			code: "package p\n\nimport log \"github.com/sirupsen/logrus\"\n\nfunc f() {\n\tlog.Fatal(\"boom\")\n}\n",
			want: nil,
		},
		{
			name: "Fatal on a local variable shadowing log passes",
			code: "package p\n\nimport \"log\"\n\nfunc f(\n\tlog recorder,\n) {\n\tlog.Fatal(\"boom\")\n}\n",
			want: nil,
		},
		{
			name: "blank and dot imports are ignored",
			code: "package p\n\nimport (\n\t_ \"os\"\n\t. \"log\"\n)\n\nfunc f() {\n\tFatal(\"boom\")\n\tx.Exit()\n}\n",
			want: nil,
		},
		{
			name: "method named panic passes",
			code: "package p\n\nfunc f(\n\tx alarm,\n) {\n\tx.panic()\n}\n",
			want: nil,
		},
	}
	for _, tc := range cases {
		assertLocations(
			t,
			tc.name,
			nopanic.Nopanic([]source.File{parsed(
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
		"p.go",
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
		Path:    "p.go",
		Rel:     "p.go",
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
