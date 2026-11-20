package testmodule_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/rust/internal/testmodule"
)

func TestTestmodule(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{
			name: "inline test module covers its mod line through its closing brace",
			code: "fn real() {\n    panic!(\"a\");\n}\n\n#[cfg(test)]\nmod tests {\n    fn fake() {\n        panic!(\"b\");\n    }\n}\n",
			want: []int{6, 7, 8, 9, 10},
		},
		{
			name: "blank lines between the attribute and the module are skipped",
			code: "#[cfg(test)]\n\nmod tests {\n    fn fake() {}\n}\n\nfn real() {}\n",
			want: []int{3, 4, 5},
		},
		{
			name: "an unclosed module runs to the end of the file",
			code: "#[cfg(test)]\nmod tests {\n    fn fake() {}\n",
			want: []int{2, 3, 4},
		},
		{
			name: "an out-of-line module covers only its declaration",
			code: "#[cfg(test)]\nmod tests;\n\nfn real() {}\n",
			want: []int{2},
		},
		{
			name: "a file without test modules has no test lines",
			code: "fn real() {\n    panic!(\"a\");\n}\n",
			want: nil,
		},
	}
	for _, tc := range cases {
		assertLines(
			t,
			tc.name,
			written(
				t,
				tc.code,
			),
			tc.want,
		)
	}
}

func TestTestmoduleUnreadableFile(
	t *testing.T,
) {
	_, err := testmodule.Testmodule(filepath.Join(
		t.TempDir(),
		"missing.rs",
	))
	if err == nil {
		t.Error("Testmodule() error = nil, want a read error")
	}
}

func written(
	t *testing.T,
	code string,
) (path string) {
	t.Helper()
	path = filepath.Join(
		t.TempDir(),
		"lib.rs",
	)
	err := os.WriteFile(
		path,
		[]byte(code),
		0o600,
	)
	if err != nil {
		t.Fatalf(
			"write fixture: %v",
			err,
		)
	}
	return path
}

func assertLines(
	t *testing.T,
	name string,
	path string,
	want []int,
) {
	t.Helper()
	got, err := testmodule.Testmodule(path)
	if err != nil {
		t.Fatalf(
			"%s: Testmodule() error = %v",
			name,
			err,
		)
	}
	isMismatch := !slices.Equal(
		got,
		want,
	) && len(got)+len(want) > 0
	if isMismatch {
		t.Errorf(
			"%s: lines = %v, want %v",
			name,
			got,
			want,
		)
	}
}
