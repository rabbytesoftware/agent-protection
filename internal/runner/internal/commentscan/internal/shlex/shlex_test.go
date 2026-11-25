package shlex_test

import (
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/shlex"
)

func TestMakefile(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "comment is reported", code: "# build\nall:\n\tgo build\n", want: []int{1}},
		{name: "recipe comment is reported", code: "all:\n\tgo build # why\n", want: []int{2}},
		{name: "hash in quotes passes", code: "all:\n\techo '#'\n", want: nil},
		{name: "hash inside a make function passes", code: "X := $(shell echo a#b)\n", want: nil},
	}
	for _, tc := range cases {
		assertLines(
			t,
			tc.name,
			shlex.Makefile(),
			tc.code,
			tc.want,
		)
	}
}

func TestShellMatches(
	t *testing.T,
) {
	assertMatches(
		t,
		shlex.Shell(),
		[]string{
			"a.sh",
			"b.bash",
			"c.zsh",
			"d.ksh",
			".env",
			".envrc",
			".env.local",
		},
		[]string{"a.py", "sh", "a.env.go"},
	)
}

func TestJustfile(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "comment is reported", code: "# legacy\nbuild:\n\tgo build ./...\n", want: []int{1}},
		{name: "recipe shebang passes", code: "test:\n    #!/usr/bin/env bash\n    go test\n", want: nil},
		{name: "recipe comment is reported", code: "test:\n    go test # why\n", want: []int{2}},
	}
	for _, tc := range cases {
		assertLines(
			t,
			tc.name,
			shlex.Justfile(),
			tc.code,
			tc.want,
		)
	}
}

func TestJustfileMatches(
	t *testing.T,
) {
	assertMatches(
		t,
		shlex.Justfile(),
		[]string{
			"Justfile",
			"justfile",
			".justfile",
			"ci.just",
		},
		[]string{"Justfile.md", "just"},
	)
}

func TestYamlMatches(
	t *testing.T,
) {
	assertMatches(
		t,
		shlex.Yaml(),
		[]string{"a.yml", "b.yaml", ".github/workflows/ci.yml"},
		[]string{"a.yamllint", "yml"},
	)
}

func TestMakefileMatches(
	t *testing.T,
) {
	assertMatches(
		t,
		shlex.Makefile(),
		[]string{
			"Makefile",
			"makefile",
			"GNUmakefile",
			"rules.mk",
		},
		[]string{"Makefile.go", "make"},
	)
}

func TestDockerfileMatches(
	t *testing.T,
) {
	assertMatches(
		t,
		shlex.Dockerfile(),
		[]string{
			"Dockerfile",
			"Dockerfile.dev",
			"Containerfile",
			"api.Dockerfile",
			"api.dockerfile",
		},
		[]string{"dockerfile.go", "Docker"},
	)
}

func assertLines(
	t *testing.T,
	name string,
	language lexeme.Language,
	code string,
	want []int,
) {
	t.Helper()
	got := lexeme.Lines(
		[]byte(code),
		language.Spans([]byte(code)),
	)
	if !slices.Equal(
		got,
		want,
	) {
		t.Errorf(
			"%s: comment lines = %v, want %v",
			name,
			got,
			want,
		)
	}
}

func assertMatches(
	t *testing.T,
	language lexeme.Language,
	matching []string,
	other []string,
) {
	t.Helper()
	got := []string{}
	for _, name := range slices.Concat(
		matching,
		other,
	) {
		got = append(
			got,
			matched(
				language,
				name,
			)...,
		)
	}

	if !slices.Equal(
		got,
		matching,
	) {
		t.Errorf(
			"matched names = %v, want %v",
			got,
			matching,
		)
	}
}

func matched(
	language lexeme.Language,
	name string,
) (names []string) {
	if !language.Matches(name) {
		return nil
	}

	return []string{name}
}
