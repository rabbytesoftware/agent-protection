package genericname_test

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/style/internal/genericname"
)

func TestGenericnameEscapeWords(
	t *testing.T,
) {
	cases := []struct {
		name string
		rel  string
		code string
		want []string
	}{
		{name: "escape-hatch function words are reported", rel: "p.go", code: "package p\n\nfunc hackSquare() {}\n\nfunc legacyFactor() {}\n\nfunc tempWorkaround() {}\n", want: []string{"R56 p.go:3", "R56 p.go:5", "R56 p.go:7"}},
		{name: "escape-hatch types and methods are reported", rel: "p.go", code: "package p\n\ntype authShim struct{}\n\nfunc (authShim) BypassCheck() {}\n", want: []string{"R56 p.go:5", "R56 p.go:3"}},
		{name: "escape-hatch package values are reported", rel: "p.go", code: "package p\n\nvar legacyMode = true\n\nconst kludgeFactor = 2\n", want: []string{"R56 p.go:3", "R56 p.go:5"}},
		{name: "smushed escape words are reported", rel: "p.go", code: "package p\n\nfunc applyHotfixes() {}\n\nfunc tempcacheGet() {}\n", want: []string{"R56 p.go:3", "R56 p.go:5"}},
		{name: "words that only embed an escape word pass", rel: "p.go", code: "package p\n\nfunc templateEnd() {}\n\nfunc readTemperature() {}\n\nfunc isCompatible() {}\n\nfunc wipeCache() {}\n\nfunc attemptSend() {}\n", want: nil},
		{name: "before and after values, constructors and verbs pass", rel: "p.go", code: "package p\n\nfunc NewServer() {}\n\nfunc swapOld() {}\n\nfunc removeUnused() {}\n\nfunc handlePatch() {}\n\nfunc saveDraft() {}\n", want: nil},
		{name: "test doubles in test files pass", rel: "p_test.go", code: "package p\n\ntype stubClock struct{}\n\nvar tempRoot = 1\n", want: nil},
		{name: "test doubles in mock files pass", rel: "mocks/p.go", code: "package mocks\n\ntype StubReadModel struct{}\n", want: nil},
		{name: "a closure value is reported once", rel: "p.go", code: "package p\n\nvar hackRun = func() {}\n", want: []string{"R56 p.go:3"}},
		{name: "plain values that say what they hold pass", rel: "p.go", code: "package p\n\nvar maxRetries, timeout = 3, 4\n", want: nil},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			genericname.Genericname([]source.File{parsedAt(
				t,
				tc.rel,
				tc.code,
			)}),
			tc.want,
		)
	}
}

func parsedAt(
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

	return source.File{
		Path:    rel,
		Rel:     rel,
		FileSet: fileSet,
		Syntax:  syntax,
	}
}
