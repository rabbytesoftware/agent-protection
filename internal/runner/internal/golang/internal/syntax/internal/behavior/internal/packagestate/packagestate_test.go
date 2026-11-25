package packagestate_test

import (
	"go/parser"
	"go/token"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/behavior/internal/packagestate"
)

func TestPackagestate(
	t *testing.T,
) {
	cases := []struct {
		name string
		rel  string
		code string
		want []string
	}{
		{
			name: "package-level map is reported",
			rel:  "p.go",
			code: "package p\n\nvar cache = map[int]int{}\n",
			want: []string{"R24 p.go:3"},
		},
		{
			name: "uninitialized string var in package main is a linker target",
			rel:  "main.go",
			code: "package main\n\nvar version string\n",
			want: nil,
		},
		{
			name: "initialized string var in package main is state",
			rel:  "main.go",
			code: "package main\n\nvar version = \"dev\"\n",
			want: []string{"R24 main.go:3"},
		},
		{
			name: "uninitialized string var outside main is state",
			rel:  "p.go",
			code: "package p\n\nvar version string\n",
			want: []string{"R24 p.go:3"},
		},
		{
			name: "init is reported",
			rel:  "p.go",
			code: "package p\n\nfunc init() {\n}\n",
			want: []string{"R24 p.go:3"},
		},
		{
			name: "init method is not the init function",
			rel:  "p.go",
			code: "package p\n\ntype t struct{}\n\nfunc (t) init() {\n}\n",
			want: nil,
		},
		{
			name: "sentinel errors built by a call pass",
			rel:  "p.go",
			code: "package p\n\nimport \"errors\"\n\nvar (\n\tErrMissing = errors.New(\"missing\")\n\terrClosed  = newClosed()\n)\n",
			want: nil,
		},
		{
			name: "sentinel names without a call are reported",
			rel:  "p.go",
			code: "package p\n\nvar ErrMissing = missing{}\n\nvar Errand = run()\n",
			want: []string{"R24 p.go:3", "R24 p.go:5"},
		},
		{
			name: "re-exported sentinel errors pass, other selectors do not",
			rel:  "p.go",
			code: "package p\n\nimport \"example.com/m/models\"\n\nvar (\n\tErrMissing = models.ErrMissing\n\tErrLimit   = models.Limit\n)\n",
			want: []string{"R24 p.go:7"},
		},
		{
			name: "names sharing one multi-value call pass",
			rel:  "p.go",
			code: "package p\n\nvar ErrA, ErrB = pair()\n",
			want: nil,
		},
		{
			name: "names without values are reported",
			rel:  "p.go",
			code: "package p\n\nvar a, b int\n",
			want: []string{"R24 p.go:3", "R24 p.go:3"},
		},
		{
			name: "blank interface assertions pass",
			rel:  "p.go",
			code: "package p\n\nvar _ I = (*T)(nil)\n",
			want: nil,
		},
		{
			name: "blank interface assertions with composite literals pass",
			rel:  "p.go",
			code: "package p\n\nvar (\n\t_ I = T{}\n\t_ I = &T{}\n\t_ I[int] = (*G[int])(nil)\n)\n",
			want: nil,
		},
		{
			name: "blank var initialized by a call is reported",
			rel:  "p.go",
			code: "package p\n\nimport \"flag\"\n\nvar _ = flag.String(\n\t\"mode\",\n\t\"fast\",\n\t\"mode\",\n)\n",
			want: []string{"R24 p.go:5"},
		},
		{
			name: "blank var running a func literal is reported, in main too",
			rel:  "main.go",
			code: "package main\n\nvar _ = func() (ok bool) {\n\tregister()\n\treturn true\n}()\n",
			want: []string{"R24 main.go:3"},
		},
		{
			name: "blank interface assertion hiding a call is reported",
			rel:  "p.go",
			code: "package p\n\nvar _ I = T{\n\tready: start(),\n}\n\nvar _ I = (*T)(boot())\n",
			want: []string{"R24 p.go:3", "R24 p.go:7"},
		},
		{
			name: "blank var without a call passes",
			rel:  "p.go",
			code: "package p\n\nvar _ = limit\n\nvar _ = func() {\n}\n",
			want: nil,
		},
		{
			name: "sentinel built by a func literal call is reported",
			rel:  "p.go",
			code: "package p\n\nvar ErrBoot = func() (err error) {\n\tregister()\n\treturn nil\n}()\n",
			want: []string{"R24 p.go:3"},
		},
		{
			name: "embedded data passes, alone or grouped",
			rel:  "p.go",
			code: "package p\n\nimport _ \"embed\"\n\n//go:embed a.txt\nvar a string\n\nvar (\n\t//go:embed b.txt\n\tb string\n\tc string\n)\n",
			want: []string{"R24 p.go:11"},
		},
		{
			name: "constants and types pass",
			rel:  "p.go",
			code: "package p\n\nconst a = 1\n\ntype t int\n",
			want: nil,
		},
		{
			name: "test files are skipped",
			rel:  "p_test.go",
			code: "package p\n\nvar fixture = 1\n\nfunc init() {\n}\n",
			want: nil,
		},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			packagestate.Packagestate([]source.File{parsed(
				t,
				tc.rel,
				tc.code,
			)}),
			tc.want,
		)
	}
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
			"parse fixture: %v",
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
