package assertion_test

import (
	"go/parser"
	"go/token"
	"maps"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/behavior/internal/assertion"
)

type fixture struct {
	name  string
	files map[string]string
	want  []string
}

func TestAssertion(
	t *testing.T,
) {
	assertCases(
		t,
		[]fixture{
			{
				name: "a test that discards its *testing.T is reported",
				files: map[string]string{
					"p/a_test.go": "package p\n\nimport \"testing\"\n\nfunc TestSquare(\n\t_ *testing.T,\n) {\n\tsquare(3)\n}\n",
				},
				want: []string{
					"R42 p/a_test.go:5",
				},
			},
			{
				name: "a test that only logs, cleans up and asks for a temp dir is reported",
				files: map[string]string{
					"p/a_test.go": "package p\n\nimport \"testing\"\n\nfunc TestSquare(\n\tt *testing.T,\n) {\n\tt.Helper()\n\tt.Parallel()\n\tt.Log(square(3))\n\tt.Cleanup(nothing)\n\t_ = t.TempDir()\n\tt.Skip(\"later\")\n}\n",
				},
				want: []string{
					"R42 p/a_test.go:5",
				},
			},
			{
				name: "a comparison that reports a mismatch passes",
				files: map[string]string{
					"p/a_test.go": "package p\n\nimport \"testing\"\n\nfunc TestSquare(\n\tt *testing.T,\n) {\n\tif square(3) != 9 {\n\t\tt.Errorf(\"square\")\n\t}\n}\n",
				},
			},
			{
				name: "examples, benchmarks, TestMain, lowercase names and methods are not tests",
				files: map[string]string{
					"p/a_test.go": "package p\n\nimport \"testing\"\n\nfunc Example() {\n\tsquare(3)\n}\n\nfunc BenchmarkSquare(\n\tb *testing.B,\n) {\n\tsquare(3)\n}\n\nfunc TestMain(\n\tm *testing.M,\n) {\n\tm.Run()\n}\n\nfunc Testify(\n\tt *testing.T,\n) {\n\tsquare(3)\n}\n\nfunc (s suite) TestSquare(\n\tt *testing.T,\n) {\n\tsquare(3)\n}\n\nfunc TestPair(\n\tt *testing.T,\n\tu *testing.T,\n) {\n}\n\nfunc TestValue(\n\tt testing.T,\n) {\n}\n",
				},
			},
			{
				name: "a fuzz target asserting inside f.Fuzz passes and one that only seeds is reported",
				files: map[string]string{
					"p/a_test.go": "package p\n\nimport \"testing\"\n\nfunc FuzzSquare(\n\tf *testing.F,\n) {\n\tf.Fuzz(func(t *testing.T, value int) {\n\t\tif square(value) < 0 {\n\t\t\tt.Fatal(value)\n\t\t}\n\t})\n}\n\nfunc FuzzCube(\n\tf *testing.F,\n) {\n\tf.Add(3)\n}\n\nfunc TestFuzzShaped(\n\tf *testing.F,\n) {\n\tf.Add(3)\n}\n",
				},
				want: []string{
					"R42 p/a_test.go:15",
				},
			},
			{
				name: "an aliased testing import is understood",
				files: map[string]string{
					"p/a_test.go": "package p\n\nimport tt \"testing\"\n\nfunc TestSquare(\n\tt *tt.T,\n) {\n\tif square(3) != 9 {\n\t\tt.Fatal(\"square\")\n\t}\n}\n\nfunc TestCube(\n\tt *tt.T,\n) {\n\tsquare(3)\n}\n",
				},
				want: []string{
					"R42 p/a_test.go:13",
				},
			},
			{
				name: "a test asserting through a helper in a non-test file passes",
				files: map[string]string{
					"p/a_test.go": "package p\n\nimport \"testing\"\n\nfunc TestSquare(\n\tt *testing.T,\n) {\n\tcheck(\n\t\tt,\n\t\tsquare(3),\n\t)\n}\n",
					"p/check.go":  "package p\n\nimport \"testing\"\n\nfunc check(\n\tt *testing.T,\n\tgot int,\n) {\n\tif got != 9 {\n\t\tt.FailNow()\n\t}\n}\n",
				},
			},
		},
	)
}

func TestAssertionSnippets(
	t *testing.T,
) {
	findings := checked(
		t,
		map[string]string{
			"p/a_test.go": "package p\n\nimport \"testing\"\n\nfunc TestBlank(\n\t_ *testing.T,\n) {\n}\n\nfunc TestUnnamed(\n\t*testing.T,\n) {\n}\n\nfunc TestSilent(\n\tt *testing.T,\n) {\n}\n",
		},
	)
	want := []string{
		"test TestBlank discards its *testing.T, so it can never fail: name the parameter t and report a wrong result with t.Errorf or t.Fatalf",
		"test TestUnnamed discards its *testing.T, so it can never fail: name the parameter t and report a wrong result with t.Errorf or t.Fatalf",
		"test TestSilent can never fail: it reaches no t.Error, t.Errorf, t.Fatal, t.Fatalf, t.Fail or t.FailNow, directly or through a helper that receives t; compare the result with the expected value and call t.Errorf when they differ",
	}
	got := []string{}
	for _, finding := range findings {
		got = append(
			got,
			finding.Snippet,
		)
	}

	if !slices.Equal(
		got,
		want,
	) {
		t.Errorf(
			"snippets = %q, want %q",
			got,
			want,
		)
	}
}

func assertCases(
	t *testing.T,
	cases []fixture,
) {
	t.Helper()
	for _, tc := range cases {
		assertFindings(
			t,
			tc,
			checked(
				t,
				tc.files,
			),
		)
	}
}

func assertFindings(
	t *testing.T,
	tc fixture,
	findings []report.Finding,
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
		tc.want,
	) && len(got)+len(tc.want) > 0
	if isMismatch {
		t.Errorf(
			"%s: findings = %v, want %v",
			tc.name,
			got,
			tc.want,
		)
	}
}

func checked(
	t *testing.T,
	contents map[string]string,
) (findings []report.Finding) {
	t.Helper()
	parsedFiles := []source.File{}
	for _, rel := range slices.Sorted(maps.Keys(contents)) {
		parsedFiles = append(
			parsedFiles,
			parsed(
				t,
				rel,
				contents[rel],
			),
		)
	}

	files := slices.DeleteFunc(
		slices.Clone(parsedFiles),
		isTestFile,
	)
	tests := slices.DeleteFunc(
		parsedFiles,
		isSourceFile,
	)
	return assertion.Assertion(
		files,
		tests,
	)
}

func isTestFile(
	file source.File,
) (test bool) {
	return source.IsTest(file.Rel)
}

func isSourceFile(
	file source.File,
) (production bool) {
	return !source.IsTest(file.Rel)
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
			"parse %s: %v",
			rel,
			err,
		)
	}

	return source.File{
		Path:    rel,
		Rel:     rel,
		FileSet: fileSet,
		Syntax:  syntax,
	}
}
