package rewrite_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
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

func TestRewritePreservesSyntaxAndSettles(
	t *testing.T,
) {
	root := t.TempDir()
	_, err := rewrite.Rewrite(
		root,
		[]source.File{fixture(
			t,
			root,
			"join.go",
			kitchenSink,
		)},
		syntax.Edits,
	)
	if err != nil {
		t.Fatalf(
			"first Rewrite() error = %v",
			err,
		)
	}

	fixed := readContent(
		t,
		filepath.Join(
			root,
			"join.go",
		),
	)
	assertSameSyntax(
		t,
		kitchenSink,
		fixed,
	)
	changed, err := rewrite.Rewrite(
		root,
		[]source.File{fixture(
			t,
			root,
			"join.go",
			fixed,
		)},
		syntax.Edits,
	)
	isChanged := err != nil || len(changed) > 0
	if isChanged {
		t.Errorf(
			"second Rewrite() = %v, %v, want no change",
			changed,
			err,
		)
	}
}

func readContent(
	t *testing.T,
	path string,
) (content string) {
	t.Helper()
	bytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf(
			"read %s: %v",
			path,
			err,
		)
	}

	return string(bytes)
}

func assertSameSyntax(
	t *testing.T,
	original string,
	fixed string,
) {
	t.Helper()
	if syntaxDump(
		t,
		original,
	) != syntaxDump(
		t,
		fixed,
	) {
		t.Errorf(
			"Rewrite() changed more than positions:\n%s",
			fixed,
		)
	}
}

func syntaxDump(
	t *testing.T,
	content string,
) (dump string) {
	t.Helper()
	syntax, err := parser.ParseFile(
		token.NewFileSet(),
		"join.go",
		content,
		parser.ParseComments|parser.SkipObjectResolution,
	)
	if err != nil {
		t.Fatalf(
			"parse: %v",
			err,
		)
	}

	printed := strings.Builder{}
	err = ast.Fprint(
		&printed,
		nil,
		syntax,
		withoutPositions,
	)
	if err != nil {
		t.Fatalf(
			"print syntax: %v",
			err,
		)
	}

	return printed.String()
}

func withoutPositions(
	name string,
	value reflect.Value,
) (kept bool) {
	return value.Type() != reflect.TypeFor[token.Pos]() && ast.NotNilFilter(
		name,
		value,
	)
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

const kitchenSink = `package fixture

import (
	"fmt"
	"strings"
)

// Handler handles.
type Handler func(name string, count int) error

// Store stores.
type Store interface {
	Put(key, value string) error
}

// Join joins.
func Join[K comparable, V any](key K, value V, rest ...string) (joined string) {

	joined = fmt.Sprintf("%v=%v", key, value)

	if joined == "" {
		return ""
	}
	for _, part := range rest {
		joined = strings.Join([]string{joined, fmt.Sprint(part, ` + "`" + `raw
string` + "`" + `)}, ",")
	}
	run := func(a int, b int) int { return a + b }
	fmt.Println(run(1, 2), /* why, not */ joined)
	joined = strings.Join(append([]string{joined}, rest...), "")
	return joined

}
`
