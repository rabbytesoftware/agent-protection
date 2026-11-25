package packagehygiene_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/shape/internal/packagehygiene"
)

func TestPackagehygieneChecks(
	t *testing.T,
) {
	cases := []struct {
		name string
		rel  string
		code string
		want []string
	}{
		{
			name: "a malformed name is reported once, without a plural finding",
			rel:  "internal/myPkgs/a.go",
			code: "package myPkgs\n",
			want: []string{"R21 internal/myPkgs/a.go:1"},
		},
		{
			name: "a malformed generic name is reported for both rules",
			rel:  "internal/my_utils/a.go",
			code: "package my_utils\n",
			want: []string{"R21 internal/my_utils/a.go:1", "R22 internal/my_utils/a.go:1"},
		},
		{
			name: "an external test package keeps its directory",
			rel:  "internal/ledger/a.go",
			code: "package ledger_test\n",
			want: []string{"R21 internal/ledger/a.go:1"},
		},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			packagehygiene.Packagehygiene(parsedAll(
				t,
				map[string]string{tc.rel: tc.code},
			)),
			tc.want,
		)
	}
}
