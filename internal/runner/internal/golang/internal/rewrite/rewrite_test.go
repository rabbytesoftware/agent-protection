package rewrite_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/rewrite"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax"
)

func TestRewrite(
	t *testing.T,
) {
	root := t.TempDir()
	untouched := "package p\n\nvar x = 1\n"
	files := []source.File{
		fixture(
			t,
			root,
			"aligned.go",
			"package p\n\nvar  y = 1\n",
		),
		fixture(
			t,
			root,
			"fixed.go",
			"package p\n\nfunc f(a int, b int) {\n\tif a > b {\n\t}\n\tg(a, b)\n}\n",
		),
		fixture(
			t,
			root,
			"untouched.go",
			untouched,
		),
	}
	changed, err := rewrite.Rewrite(
		root,
		files,
		syntax.Edits,
	)
	if err != nil {
		t.Fatalf(
			"Rewrite() error = %v",
			err,
		)
	}

	assertChanged(
		t,
		changed,
		[]string{"aligned.go", "fixed.go"},
	)
	assertContent(
		t,
		filepath.Join(
			root,
			"aligned.go",
		),
		"package p\n\nvar y = 1\n",
	)
	assertContent(
		t,
		filepath.Join(
			root,
			"fixed.go",
		),
		"package p\n\nfunc f(\n\ta int,\n\tb int,\n) {\n\tif a > b {\n\t}\n\n\tg(\n\t\ta,\n\t\tb,\n\t)\n}\n",
	)
	assertContent(
		t,
		filepath.Join(
			root,
			"untouched.go",
		),
		untouched,
	)
}

func TestRewriteFailures(
	t *testing.T,
) {
	cases := []struct {
		name  string
		edits func(source.File) []source.Edit
	}{
		{name: "edit breaking the syntax", edits: breakSyntax},
		{name: "edits that never settle", edits: neverSettle},
	}
	for _, tc := range cases {
		assertRewriteFails(
			t,
			tc.name,
			tc.edits,
		)
	}
}

func assertRewriteFails(
	t *testing.T,
	name string,
	edits func(source.File) []source.Edit,
) {
	t.Helper()
	root := t.TempDir()
	changed, err := rewrite.Rewrite(
		root,
		[]source.File{fixture(
			t,
			root,
			"p.go",
			"package p\n",
		)},
		edits,
	)
	isMismatch := err == nil || len(changed) > 0
	if isMismatch {
		t.Errorf(
			"%s: Rewrite() = %v, %v, want no change and an error",
			name,
			changed,
			err,
		)
	}
}

func TestRewriteReportsWriteFailures(
	t *testing.T,
) {
	root := t.TempDir()
	file := fixture(
		t,
		root,
		"p.go",
		"package p\n\nfunc f(a int) {\n}\n",
	)
	err := os.Chmod(
		file.Path,
		0o400,
	)
	if err != nil {
		t.Fatalf(
			"chmod: %v",
			err,
		)
	}

	_, err = rewrite.Rewrite(
		root,
		[]source.File{file},
		syntax.Edits,
	)
	if err == nil {
		t.Error("Rewrite() error = nil, want a write error")
	}
}

func breakSyntax(
	file source.File,
) (edits []source.Edit) {
	return []source.Edit{{Start: len(file.Content), End: len(file.Content), Text: "func {"}}
}

func neverSettle(
	file source.File,
) (edits []source.Edit) {
	return []source.Edit{{Start: 0, End: 0, Text: "\n"}}
}

func fixture(
	t *testing.T,
	root string,
	name string,
	content string,
) (file source.File) {
	t.Helper()
	path := filepath.Join(
		root,
		name,
	)
	err := os.WriteFile(
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

	file, err = source.Parse(
		root,
		path,
		[]byte(content),
	)
	if err != nil {
		t.Fatalf(
			"parse %s: %v",
			path,
			err,
		)
	}

	return file
}

func assertChanged(
	t *testing.T,
	got []string,
	want []string,
) {
	t.Helper()
	if !slices.Equal(
		got,
		want,
	) {
		t.Errorf(
			"Rewrite() changed = %v, want %v",
			got,
			want,
		)
	}
}

func assertContent(
	t *testing.T,
	path string,
	want string,
) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf(
			"read %s: %v",
			path,
			err,
		)
	}

	if string(content) != want {
		t.Errorf(
			"%s =\n%s\nwant\n%s",
			path,
			content,
			want,
		)
	}
}
