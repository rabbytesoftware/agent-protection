package mockimport_test

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch/internal/dependency/internal/mockimport"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	fixtureRoot = "/repo"
	fakesImport = "package arith\n\nimport (\n\t\"fmt\"\n\n\t\"example.com/m/internal/fakes\"\n)\n\nvar _ = fmt.Sprint\n\nvar _ = fakes.NewMemo\n"
)

func TestMockimport(
	t *testing.T,
) {
	cases := []struct {
		name string
		rel  string
		code string
		want []string
	}{
		{name: "production file importing a fakes package is reported", rel: "container.go", code: fakesImport, want: []string{"R63 container.go:6"}},
		{name: "production file importing a mocks package is reported", rel: "app/app.go", code: "package app\n\nimport \"example.com/m/store/mocks\"\n", want: []string{"R63 app/app.go:3"}},
		{name: "production file importing a mock framework is reported", rel: "app/app.go", code: "package app\n\nimport \"github.com/stretchr/testify/mock\"\n", want: []string{"R63 app/app.go:3"}},
		{name: "production file importing a mock_ package is reported", rel: "app/app.go", code: "package app\n\nimport \"example.com/m/mock_store\"\n", want: []string{"R63 app/app.go:3"}},
		{name: "mock file importing another mock passes", rel: "store/mocks/store.go", code: "package mocks\n\nimport \"example.com/m/clock/fakes\"\n", want: nil},
		{name: "mock-named file passes", rel: "store/mock_store.go", code: "package store\n\nimport \"example.com/m/clock/fakes\"\n", want: nil},
		{name: "test helper importing testing passes", rel: "storetest/storetest.go", code: "package storetest\n\nimport (\n\t\"testing\"\n\n\t\"example.com/m/store/mocks\"\n)\n", want: nil},
		{name: "test file passes", rel: "app/app_test.go", code: "package app\n\nimport \"example.com/m/store/mocks\"\n", want: nil},
		{name: "look-alike names pass", rel: "app/app.go", code: "package app\n\nimport (\n\t\"example.com/m/mockery\"\n\t\"example.com/m/fakeroot\"\n\t\"net/http/httptest\"\n)\n", want: nil},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			mockimport.Mockimport([]source.File{parsed(
				t,
				tc.rel,
				tc.code,
			)}),
			tc.want,
		)
	}
}

func TestMockimportSnippet(
	t *testing.T,
) {
	findings := mockimport.Mockimport([]source.File{parsed(
		t,
		"container.go",
		fakesImport,
	)})
	want := "production file imports mock package example.com/m/internal/fakes: mocks and fakes exist for tests only; wire the real implementation here and import example.com/m/internal/fakes only from _test.go files"
	isMatch := len(findings) == 1 && findings[0].Snippet == want
	if !isMatch {
		t.Errorf(
			"Mockimport() = %v, want one finding with snippet %q",
			findings,
			want,
		)
	}
}

func parsed(
	t *testing.T,
	rel string,
	code string,
) (file source.File) {
	t.Helper()
	file, err := source.Parse(
		fixtureRoot,
		filepath.Join(
			fixtureRoot,
			rel,
		),
		[]byte(code),
	)
	if err != nil {
		t.Fatalf(
			"parse %s: %v",
			rel,
			err,
		)
	}

	return file
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

	if !slices.Equal(
		got,
		append(
			[]string{},
			want...,
		),
	) {
		t.Errorf(
			"%s: findings = %v, want %v",
			name,
			got,
			want,
		)
	}
}
