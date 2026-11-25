package nostutter_test

import (
	"go/parser"
	"go/token"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/nostutter"
)

func TestNostutter(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "type prefixed with the package name is reported",
			code: "package vault\n\ntype VaultManager struct{}\n",
			want: []string{"p.go:3"},
		},
		{
			name: "function prefixed with the package name is reported",
			code: "package vault\n\nfunc VaultOpen() {}\n",
			want: []string{"p.go:3"},
		},
		{
			name: "type named exactly after the package passes",
			code: "package vault\n\ntype Vault struct{}\n",
			want: nil,
		},
		{
			name: "entry function named exactly after the package passes",
			code: "package noelse\n\nfunc Noelse() {}\n",
			want: nil,
		},
		{
			name: "method prefixed with the package name passes",
			code: "package vault\n\ntype Store struct{}\n\nfunc (s Store) VaultKey() {}\n",
			want: nil,
		},
		{
			name: "unexported type prefixed with the package name passes",
			code: "package vault\n\ntype vaultEntry struct{}\n",
			want: nil,
		},
		{
			name: "name without the package prefix passes",
			code: "package user\n\ntype Service struct{}\n",
			want: nil,
		},
	}
	for _, tc := range cases {
		assertLocations(
			t,
			tc.name,
			nostutter.Nostutter([]source.File{parsed(
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
