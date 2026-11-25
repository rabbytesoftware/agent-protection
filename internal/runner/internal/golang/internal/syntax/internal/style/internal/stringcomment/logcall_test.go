package stringcomment_test

import (
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/style/internal/stringcomment"
)

func TestStringcommentLogCalls(
	t *testing.T,
) {
	cases := []struct {
		name string
		path string
		code string
		want []string
	}{
		{name: "TODO in a slog message is reported", path: "p.go", code: "package p\n\nfunc F() {\n\tslog.Debug(\"TODO: factor is ignored\")\n}\n", want: []string{"R07 p.go:4"}},
		{name: "escape-hatch word in a log.Printf format is reported", path: "p.go", code: "package p\n\nfunc F() {\n\tlog.Printf(\"using a workaround for %s\", name)\n}\n", want: []string{"R07 p.go:4"}},
		{name: "for now in a logger method is reported", path: "p.go", code: "package p\n\nfunc F() {\n\tlogger.Warn(\"retries disabled \" + \"for now\")\n}\n", want: []string{"R07 p.go:4"}},
		{name: "a Log-prefixed method with a marker is reported", path: "p.go", code: "package p\n\nfunc F() {\n\tlogger.LogAttrs(ctx, level, \"XXX unscaled\")\n}\n", want: []string{"R07 p.go:4"}},
		{name: "a marker in t.Skip is reported", path: "p_test.go", code: "package p\n\nfunc TestF(t *testing.T) {\n\tt.Skip(\"FIXME flaky\")\n}\n", want: []string{"R07 p_test.go:4"}},
		{name: "a sentence in t.Log is reported", path: "p_test.go", code: "package p\n\nfunc TestF(t *testing.T) {\n\tt.Logf(\"factor stays at one until billing migrates %d\", 1)\n}\n", want: []string{"R07 p_test.go:4"}},
		{name: "a sentence with a marker in t.Log is reported once", path: "p_test.go", code: "package p\n\nfunc TestF(t *testing.T) {\n\tt.Log(\"TODO: this is a hack because factor is ignored\")\n}\n", want: []string{"R07 p_test.go:4"}},
		{name: "a nested logger selector with a marker is reported", path: "p.go", code: "package p\n\nfunc F() {\n\ts.logger.Info(\"HACK around the cache\")\n}\n", want: []string{"R07 p.go:4"}},
		{name: "a plain log sentence passes", path: "p.go", code: "package p\n\nfunc F() {\n\tslog.Info(\"server listening on the configured address and port\", \"addr\", addr)\n\tlogger.Log(\"connection closed by the remote peer after timeout\")\n}\n", want: nil},
		{name: "a sentence in t.Log outside a test file passes", path: "p.go", code: "package p\n\nfunc F(t T) {\n\tt.Log(\"factor stays at one until billing migrates\")\n}\n", want: nil},
		{name: "t.Skip with a reason and short t.Log pass", path: "p_test.go", code: "package p\n\nfunc TestF(t *testing.T) {\n\tt.Skip(\"requires docker, run with the integration tag set\")\n\tt.Log(\"got\", got)\n}\n", want: nil},
		{name: "embedded words, non-log calls and non-string arguments pass", path: "p.go", code: "package p\n\nfunc F() {\n\tslog.Info(\"hackathon started\", \"todo\", todos)\n\tfmt.Println(\"TODO later\")\n\tlog.Print(count)\n\tf()()\n}\n", want: nil},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			stringcomment.Stringcomment([]source.File{parsed(
				t,
				tc.path,
				tc.code,
			)}),
			tc.want,
		)
	}
}

func TestStringcommentLogSnippets(
	t *testing.T,
) {
	findings := stringcomment.Stringcomment([]source.File{parsed(
		t,
		"p_test.go",
		"package p\n\nfunc TestF(t *testing.T) {\n\tslog.Debug(\"TODO: factor is a hack until billing migrates, squares stay unscaled for now\")\n\tt.Log(\"factor stays at one until billing migrates\")\n}\n",
	)})
	want := []report.Finding{
		{Rule: "R07", Location: "p_test.go:4", Snippet: "slog.Debug message \"TODO: factor is a hack until billing migrates, s...\" is a note about provisional code (a TODO marker or an escape-hatch word): a log line says what happened, not what is left to fix; delete the note and fix the code, and track the work outside the source"},
		{Rule: "R07", Location: "p_test.go:5", Snippet: "t.Log message \"factor stays at one until billing migrates\" is a sentence explaining the test: a comment moved into a log line; delete it and let the test name and assertions say it"},
	}
	if !slices.Equal(
		findings,
		want,
	) {
		t.Errorf(
			"Stringcomment() = %v, want %v",
			findings,
			want,
		)
	}
}
