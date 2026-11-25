package siblingimport_test

import (
	"testing"
)

const roleImportingStore = "package role\n\nimport \"example.com/m/store\"\n\nvar Default store.Store\n"

func TestSiblingimportsChains(
	t *testing.T,
) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{
			name: "importer below a sibling names the sibling it sits under",
			files: map[string]string{
				"internal/app/app.go":                 unit("app"),
				"internal/engine/e.go":                unit("engine"),
				"internal/adapter/ad.go":              unit("adapter internal/engine"),
				"internal/app/internal/deep/d.go":     unit("deep internal/adapter"),
				"internal/app/internal/deep/other.go": unit("deep"),
			},
			want: []string{"imports sibling adapter (through app, the sibling this package sits under), which itself imports sibling engine: siblings form one layer; have the parent's container.go pass adapter in, or make adapter a foundation by dropping its own sibling imports"},
		},
		{
			name: "importer inside an internal directory is told it sits inside internal/",
			files: map[string]string{
				"jobq.go":                   unit("jobq"),
				"store/store.go":            unit("store"),
				"contracts/role/role.go":    roleImportingStore,
				"internal/backoff/b.go":     unit("backoff contracts/role"),
				"internal/backoff/other.go": unit("backoff"),
			},
			want: []string{"imports contracts, a sibling of the internal/ directory this package lives in (all of internal/ counts as one package beside contracts), and contracts itself imports sibling store: siblings form one layer; have the parent's container.go pass contracts in, or make contracts a foundation by dropping its own sibling imports"},
		},
		{
			name: "one lifted edge made by many imports is reported once at the first import",
			files: map[string]string{
				"store/store.go":         unit("store"),
				"contracts/role/role.go": roleImportingStore,
				"contracts/work/work.go": "package work\n\ntype Job struct{}\n",
				"internal/a/a.go":        unit("a contracts/role contracts/work"),
				"internal/b/b.go":        unit("b contracts/role"),
				"internal/c/c.go":        unit("c contracts/work"),
				"internal/d/d.go":        unit("d contracts/role"),
			},
			want: []string{"imports contracts, a sibling of the internal/ directory this package lives in (all of internal/ counts as one package beside contracts), and contracts itself imports sibling store: siblings form one layer; have the parent's container.go pass contracts in, or make contracts a foundation by dropping its own sibling imports; 5 imports across 4 packages under internal/ reach contracts the same way, for example internal/a/a.go:5, internal/b/b.go:4, internal/c/c.go:4"},
		},
		{
			name: "several imports of one package name the package once",
			files: map[string]string{
				"internal/a/a.go":   unit("a internal/b/x"),
				"internal/a/a2.go":  unit("a internal/b/y"),
				"internal/b/x/x.go": unit("x internal/c"),
				"internal/b/y/y.go": "package y\n\ntype Y struct{}\n",
				"internal/c/c.go":   unit("c"),
			},
			want: []string{"imports sibling b, which itself imports sibling c: siblings form one layer; have the parent's container.go pass b in, or make b a foundation by dropping its own sibling imports; 2 imports in this package reach b the same way, for example internal/a/a2.go:4"},
		},
		{
			name: "distinct lifted edges are reported separately",
			files: map[string]string{
				"internal/a/a.go": unit("a internal/b internal/x"),
				"internal/b/b.go": unit("b internal/c"),
				"internal/x/x.go": unit("x internal/c"),
				"internal/c/c.go": unit("c"),
			},
			want: []string{
				"imports sibling b, which itself imports sibling c: siblings form one layer; have the parent's container.go pass b in, or make b a foundation by dropping its own sibling imports",
				"imports sibling x, which itself imports sibling c: siblings form one layer; have the parent's container.go pass x in, or make x a foundation by dropping its own sibling imports",
				"imports siblings b, x: wiring siblings together is the parent's job; move it into container.go and inject what this package needs through an interface it declares",
			},
		},
	}
	for _, tc := range cases {
		assertSnippets(
			t,
			checked(
				t,
				tc.files,
			),
			tc.want,
		)
	}
}

func TestSiblingimportsChainLocation(
	t *testing.T,
) {
	findings := checked(
		t,
		map[string]string{
			"internal/a/z.go":  unit("a internal/b"),
			"internal/a/a.go":  "package a\n\nimport (\n\t\"fmt\"\n\t\"os\"\n\t\"example.com/m/internal/b\"\n)\n\nfunc Run() {}\n",
			"internal/a/a2.go": unit("a internal/b"),
			"internal/b/b.go":  unit("b internal/c"),
			"internal/c/c.go":  unit("c"),
		},
	)
	assertFindings(
		t,
		"first offending import sorted by path then line",
		findings,
		[]string{"R16 internal/a/a.go:6"},
	)
}
