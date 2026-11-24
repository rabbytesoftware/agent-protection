package report_test

import (
	"maps"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
)

func TestReport(
	t *testing.T,
) {
	cases := []struct {
		name          string
		findings      []report.Finding
		wantLocations []string
		wantDocs      []string
	}{
		{
			name:          "sorts by rule then location",
			findings:      []report.Finding{{Rule: "R33", Location: "b.go:1"}, {Rule: "R32", Location: "z.go:1"}, {Rule: "R32", Location: "a.go:1"}},
			wantLocations: []string{"a.go:1", "z.go:1", "b.go:1"},
			wantDocs:      []string{"R32", "R33"},
		},
		{
			name:          "documents each distinct rule once",
			findings:      []report.Finding{{Rule: "R32", Location: "a.go:1"}, {Rule: "R32", Location: "b.go:2"}},
			wantLocations: []string{"a.go:1", "b.go:2"},
			wantDocs:      []string{"R32"},
		},
		{
			name:          "keeps unregistered rules without docs",
			findings:      []report.Finding{{Rule: "TS-ESLINT:plugin-rule", Location: "x.ts:1"}},
			wantLocations: []string{"x.ts:1"},
			wantDocs:      nil,
		},
	}
	for _, tc := range cases {
		assertReport(
			t,
			tc.name,
			tc.findings,
			tc.wantLocations,
			tc.wantDocs,
		)
	}
}

func assertReport(
	t *testing.T,
	name string,
	findings []report.Finding,
	wantLocations []string,
	wantDocs []string,
) {
	t.Helper()
	output, err := report.Report(findings)
	if err != nil {
		t.Fatalf(
			"%s: Report() error = %v",
			name,
			err,
		)
	}

	gotLocations := locations(output.Findings)
	if !slices.Equal(
		gotLocations,
		wantLocations,
	) {
		t.Errorf(
			"%s: locations = %v, want %v",
			name,
			gotLocations,
			wantLocations,
		)
	}

	gotDocs := slices.Sorted(maps.Keys(output.RuleDocs))
	if !slices.Equal(
		gotDocs,
		wantDocs,
	) {
		t.Errorf(
			"%s: documented rules = %v, want %v",
			name,
			gotDocs,
			wantDocs,
		)
	}
}

func locations(
	findings []report.Finding,
) (result []string) {
	for _, finding := range findings {
		result = append(
			result,
			finding.Location,
		)
	}

	return result
}
