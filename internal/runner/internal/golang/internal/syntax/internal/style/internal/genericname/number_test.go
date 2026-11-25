package genericname_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/style/internal/genericname"
)

func TestNumberedNames(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "exported name with a trailing number is reported",
			code: "package p\n\nfunc ParseOrder2() {}\n\ntype Ledger3 struct{}\n",
			want: []string{"R56 p.go:3", "R56 p.go:5"},
		},
		{
			name: "domain terms ending in digits pass",
			code: "package p\n\nfunc EncodeBase64() {}\n\nfunc hashSHA256() {}\n\nfunc readInt64() {}\n\nfunc parseX509() {}\n\nfunc uploadS3() {}\n\ntype ClientOAuth2 struct{}\n\nfunc dialIPv6() {}\n\nfunc normalVec3() {}\n",
			want: nil,
		},
		{
			name: "version and percentile markers pass",
			code: "package p\n\ntype v0 struct{}\n\nfunc ParseV2() {}\n\nfunc latencyP99() {}\n",
			want: nil,
		},
		{
			name: "status codes and benchmark sizes pass",
			code: "package p\n\nfunc TestGet_Returns404() {}\n\nfunc BenchmarkChain10() {}\n",
			want: nil,
		},
		{
			name: "numbers outside the status range are reported",
			code: "package p\n\nfunc TestGet_Returns700() {}\n\nfunc TestGet_Case10() {}\n",
			want: []string{"R56 p.go:3", "R56 p.go:5"},
		},
		{
			name: "letter and number labels are reported",
			code: "package p\n\nfunc deliveryA1() {}\n\nfunc handleData2() {}\n",
			want: []string{"R56 p.go:3", "R56 p.go:5"},
		},
		{
			name: "domain term glued into a lowercase word is reported",
			code: "package p\n\nfunc match2() {}\n\nfunc encodebase64() {}\n",
			want: []string{"R56 p.go:3", "R56 p.go:5"},
		},
		{
			name: "digits inside a name pass",
			code: "package p\n\nfunc http2Server() {}\n\nfunc split3Ways() {}\n\nfunc base64() {}\n\nfunc parse_int64() {}\n",
			want: nil,
		},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			genericname.Genericname([]source.File{parsed(
				t,
				tc.code,
			)}),
			tc.want,
		)
	}
}
