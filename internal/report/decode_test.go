package report_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
)

func TestDecode(
	t *testing.T,
) {
	cases := []struct {
		document     string
		wantFindings int
		wantErr      bool
	}{
		{document: `{"rule_docs":{},"findings":[{"rule":"R32","location":"a.go:1","snippet":"x"}]}`, wantFindings: 1},
		{document: `{"rule_docs":{},"findings":[]}`, wantFindings: 0},
		{document: "not json", wantErr: true},
	}
	for _, tc := range cases {
		assertDecode(
			t,
			tc.document,
			tc.wantFindings,
			tc.wantErr,
		)
	}
}

func assertDecode(
	t *testing.T,
	document string,
	wantFindings int,
	wantErr bool,
) {
	t.Helper()
	output, err := report.Decode([]byte(document))
	isMismatch := len(output.Findings) != wantFindings || (err != nil) != wantErr
	if isMismatch {
		t.Errorf(
			"Decode(%q) = %d findings, error %v; want %d, error %v",
			document,
			len(output.Findings),
			err,
			wantFindings,
			wantErr,
		)
	}
}
