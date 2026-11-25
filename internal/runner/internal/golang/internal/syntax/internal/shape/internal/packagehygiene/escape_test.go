package packagehygiene_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/shape/internal/packagehygiene"
)

func TestPackagehygieneEscapeHatches(
	t *testing.T,
) {
	cases := []struct {
		name string
		want int
	}{
		{name: "shim", want: 1},
		{name: "stub", want: 1},
		{name: "patch", want: 1},
		{name: "fixme", want: 1},
		{name: "hotfix", want: 1},
		{name: "interim", want: 1},
		{name: "stopgap", want: 1},
		{name: "quickfix", want: 1},
		{name: "bandaid", want: 1},
		{name: "kludge", want: 1},
		{name: "adhoc", want: 1},
		{name: "scratch", want: 1},
		{name: "draft", want: 1},
		{name: "sandbox", want: 1},
		{name: "playground", want: 1},
		{name: "experimental", want: 1},
		{name: "bypass", want: 1},
		{name: "shortcut", want: 1},
		{name: "workaround", want: 1},
		{name: "oldstore", want: 1},
		{name: "tempcache", want: 1},
		{name: "authshim", want: 1},
		{name: "hackfix", want: 1},
		{name: "temporarystore", want: 1},
		{name: "olddispatch", want: 1},
		{name: "template", want: 0},
		{name: "temperature", want: 0},
		{name: "temporal", want: 0},
		{name: "tmpl", want: 0},
		{name: "dispatch", want: 0},
		{name: "dispatcher", want: 0},
		{name: "jsonpatch", want: 0},
		{name: "threshold", want: 0},
		{name: "manifold", want: 0},
		{name: "scaffold", want: 0},
		{name: "golden", want: 0},
		{name: "news", want: 0},
		{name: "newsletter", want: 0},
		{name: "renew", want: 0},
		{name: "overdraft", want: 0},
		{name: "ledger", want: 0},
	}
	for _, tc := range cases {
		assertReported(
			t,
			tc.name,
			tc.want,
		)
	}
}

func assertReported(
	t *testing.T,
	name string,
	want int,
) {
	t.Helper()
	findings := packagehygiene.Packagehygiene(parsedAll(
		t,
		map[string]string{"internal/" + name + "/a.go": "package " + name + "\n"},
	))
	if len(findings) != want {
		t.Errorf(
			"package %s: findings = %v, want %d",
			name,
			findings,
			want,
		)
	}
}
