package command

import "testing"

func TestFix(
	t *testing.T,
) {
	failing := fixtureModule(
		t,
		"failing",
		"package failing\n\nfunc F(a int, b int) {}\n",
	)
	cases := []struct {
		root     string
		wantCode int
	}{
		{root: failing, wantCode: cleanExitCode},
		{root: t.TempDir(), wantCode: findingsExitCode},
	}
	for _, tc := range cases {
		assertCode(
			t,
			"Fix("+tc.root+")",
			realTools().Fix(tc.root),
			tc.wantCode,
		)
	}
}
