package packagehygiene_test

import (
	"go/parser"
	"go/token"
	"maps"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/packagehygiene"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestPackagehygiene(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "generic util package is reported",
			code: "package util\n\nfunc A() {}\n",
			want: []string{"R22 p.go:1"},
		},
		{
			name: "generic shared package is reported",
			code: "package shared\n\nfunc A() {}\n",
			want: []string{"R22 p.go:1"},
		},
		{
			name: "underscored package name is reported",
			code: "package my_ledger\n\nfunc A() {}\n",
			want: []string{"R21 p.go:1"},
		},
		{
			name: "mixed case package name is reported",
			code: "package myPkg\n\nfunc A() {}\n",
			want: []string{"R21 p.go:1"},
		},
		{
			name: "short lowercase noun passes",
			code: "package domain\n\nfunc A() {}\n",
			want: nil,
		},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			packagehygiene.Packagehygiene([]source.File{parsed(
				t,
				"p.go",
				tc.code,
			)}),
			tc.want,
		)
	}
}

func TestPackagehygieneDirectories(
	t *testing.T,
) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{
			name:  "directory named after its package passes, one finding per package",
			files: map[string]string{"internal/ledger/a.go": "package ledger\n", "internal/ledger/b.go": "package ledger\n"},
			want:  nil,
		},
		{
			name:  "directory not named after its package is reported",
			files: map[string]string{"internal/Helpers/handlers.go": "package handlers\n"},
			want:  []string{"R21 internal/Helpers/handlers.go:1", "R21 internal/Helpers", "R21 internal/Helpers", "R22 internal/Helpers"},
		},
		{
			name:  "directory with a generic underscored name is reported",
			files: map[string]string{"internal/common_utils/ledger.go": "package ledger\n"},
			want:  []string{"R21 internal/common_utils", "R21 internal/common_utils", "R22 internal/common_utils"},
		},
		{
			name:  "package main, the root directory and version directories keep their own names",
			files: map[string]string{"cmd/my-tool/main.go": "package main\n", "root.go": "package agent\n", "api/v2/api.go": "package api\n"},
			want:  nil,
		},
		{
			name:  "mock directories are left to the mock conventions",
			files: map[string]string{"internal/store/mocks/store.go": "package mock_store\n"},
			want:  nil,
		},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			packagehygiene.Packagehygiene(parsedAll(
				t,
				tc.files,
			)),
			tc.want,
		)
	}
}

func parsedAll(
	t *testing.T,
	files map[string]string,
) (parsedFiles []source.File) {
	t.Helper()
	for _, rel := range slices.Sorted(maps.Keys(files)) {
		parsedFiles = append(
			parsedFiles,
			parsed(
				t,
				rel,
				files[rel],
			),
		)
	}

	return parsedFiles
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

	return source.File{Path: rel, Rel: rel, FileSet: fileSet, Syntax: syntax}
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
