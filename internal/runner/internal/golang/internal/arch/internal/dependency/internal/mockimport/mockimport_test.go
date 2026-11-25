package mockimport_test

import (
	"maps"
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

func TestMockimportUsedMocks(
	t *testing.T,
) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{
			name: "mock file whose declarations production calls is reported",
			files: map[string]string{
				"fake.go":      "package arith\n\nfunc newFake() {}\n",
				"container.go": "package arith\n\nfunc Arith() {\n\tnewFake()\n}\n",
			},
			want: []string{"R63 fake.go"},
		},
		{
			name: "mock-named file imported by a production package is reported at the mock",
			files: map[string]string{
				"internal/report/mock_wiring.go": "package report\n\nfunc Wired() {}\n",
				"internal/app/app.go":            "package app\n\nimport \"example.com/m/internal/report\"\n\nfunc Run() {\n\treport.Wired()\n}\n",
			},
			want: []string{"R63 internal/report/mock_wiring.go"},
		},
		{
			name: "mock package imported by production is reported once, at the import",
			files: map[string]string{
				"store/mocks/store.go": "package mocks\n\ntype Store struct{}\n",
				"app/app.go":           "package app\n\nimport \"example.com/m/store/mocks\"\n\nvar _ = mocks.Store{}\n",
			},
			want: []string{"R63 app/app.go:3"},
		},
		{
			name: "mock referenced only from tests passes",
			files: map[string]string{
				"fake.go":       "package arith\n\nfunc newFake() {}\n",
				"arith_test.go": "package arith\n\nfunc TestArith() {\n\tnewFake()\n}\n",
			},
			want: nil,
		},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			mockimport.Mockimport(parsedFiles(
				t,
				tc.files,
			)),
			tc.want,
		)
	}
}

func parsedFiles(
	t *testing.T,
	sources map[string]string,
) (files []source.File) {
	t.Helper()
	for _, rel := range slices.Sorted(maps.Keys(sources)) {
		files = append(
			files,
			parsed(
				t,
				rel,
				sources[rel],
			),
		)
	}

	return files
}

func TestMockimportUsedMockSnippet(
	t *testing.T,
) {
	findings := mockimport.Mockimport([]source.File{
		parsed(
			t,
			"fake.go",
			"package arith\n\nfunc newFake() {}\n",
		),
		parsed(
			t,
			"container.go",
			"package arith\n\nfunc Arith() {\n\tnewFake()\n}\n",
		),
	})
	want := "fake.go is named like a mock but production file container.go uses its declarations: mocks and fakes exist for tests only, so it is checked as production code; name the file after what it implements and give it its own test, or reference it only from _test.go files"
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
