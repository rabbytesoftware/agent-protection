package notesfile_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/notesfile"
)

func TestNotesfile(
	t *testing.T,
) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{
			name: "notes beside package code are reported, in any case of extension",
			files: map[string]string{
				"internal/square/square.go":  "package square\n",
				"internal/square/NOTES.md":   "TODO: remove factor\n",
				"internal/square/design.TXT": "why it is shaped like this\n",
				"internal/square/plan.rst":   "plan\n",
				"internal/square/README.md":  "square\n",
			},
			want: []string{
				"R07 internal/square/NOTES.md",
				"R07 internal/square/README.md",
				"R07 internal/square/design.TXT",
				"R07 internal/square/plan.rst",
			},
		},
		{
			name: "a directory holding only tests still counts as code",
			files: map[string]string{
				"internal/square/square_test.go": "package square_test\n",
				"internal/square/notes.txt":      "notes\n",
			},
			want: []string{"R07 internal/square/notes.txt"},
		},
		{
			name: "module root, docs, testdata and example trees, directories without Go and other file types pass",
			files: map[string]string{
				"main.go":                         "package main\n",
				"README.md":                       "module\n",
				"CHANGELOG.md":                    "changes\n",
				"docs/guide/guide.go":             "package guide\n",
				"docs/guide/guide.md":             "guide\n",
				"internal/square/testdata/in.txt": "fixture\n",
				"internal/square/testdata/gen.go": "package testdata\n",
				"internal/square/square.go":       "package square\n",
				"internal/square/config.yaml":     "a: 1\n",
				"assets/notes.md":                 "notes\n",
				"examples/basic/main.go":          "package main\n",
				"examples/basic/README.md":        "run it\n",
			},
			want: nil,
		},
		{
			name: "files the package embeds are data, including quoted, all: and glob patterns",
			files: map[string]string{
				"internal/square/square.go":      "package square\n\nimport _ \"embed\"\n\n//go:embed words.txt \"other words.md\"\nvar words string\n\n//go:embed all:*.rst\nvar texts string\n",
				"internal/square/words.txt":      "a b\n",
				"internal/square/other words.md": "c\n",
				"internal/square/table.rst":      "d\n",
				"internal/square/unembedded.md":  "e\n",
			},
			want: []string{"R07 internal/square/unembedded.md"},
		},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			written(
				t,
				tc.files,
			),
			tc.want,
		)
	}
}

func TestNotesfileFailsOnBadEmbedPattern(
	t *testing.T,
) {
	_, err := notesfile.Notesfile(written(
		t,
		map[string]string{
			"pkg/p.go":     "package p\n\n//go:embed [\nvar words string\n",
			"pkg/notes.md": "notes\n",
		},
	))
	if err == nil {
		t.Error("Notesfile() error = nil, want an error for a malformed embed pattern")
	}
}

func written(
	t *testing.T,
	files map[string]string,
) (root string) {
	t.Helper()
	root = t.TempDir()
	for path, content := range files {
		write(
			t,
			filepath.Join(
				root,
				path,
			),
			content,
		)
	}

	return root
}

func TestNotesfileFailsOnMissingRoot(
	t *testing.T,
) {
	_, err := notesfile.Notesfile(filepath.Join(
		t.TempDir(),
		"missing",
	))
	if err == nil {
		t.Error("Notesfile() error = nil, want an error for a missing root")
	}
}

func TestNotesfileSnippet(
	t *testing.T,
) {
	root := t.TempDir()
	write(
		t,
		filepath.Join(
			root,
			"pkg/p.go",
		),
		"package p\n",
	)
	write(
		t,
		filepath.Join(
			root,
			"pkg/NOTES.md",
		),
		"notes\n",
	)
	findings, err := notesfile.Notesfile(root)
	want := "pkg/NOTES.md is a notes file beside the code: what the package does belongs in its names and structure, and why it changed in the commit message; delete it (a README.md at the module root or a docs/ tree is where documentation lives, and a file the package embeds with //go:embed is data)"
	isMismatch := err != nil ||
		len(findings) != 1 ||
		findings[0].Snippet != want
	if isMismatch {
		t.Errorf(
			"Notesfile() = %v, %v, want one finding with snippet %q",
			findings,
			err,
			want,
		)
	}
}

func assertFindings(
	t *testing.T,
	name string,
	root string,
	want []string,
) {
	t.Helper()
	findings, err := notesfile.Notesfile(root)
	if err != nil {
		t.Fatalf(
			"%s: Notesfile() error = %v",
			name,
			err,
		)
	}

	got := []string{}
	for _, finding := range findings {
		got = append(
			got,
			finding.Rule+" "+finding.Location,
		)
	}

	slices.Sort(got)
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

func write(
	t *testing.T,
	path string,
	content string,
) {
	t.Helper()
	err := os.MkdirAll(
		filepath.Dir(path),
		0o755,
	)
	if err != nil {
		t.Fatalf(
			"mkdir %s: %v",
			path,
			err,
		)
	}

	err = os.WriteFile(
		path,
		[]byte(content),
		0o600,
	)
	if err != nil {
		t.Fatalf(
			"write %s: %v",
			path,
			err,
		)
	}
}
