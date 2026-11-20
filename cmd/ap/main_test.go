package main

import (
	"os"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
)

func TestWriteReportAndExit_WritesReportFileOnTheZeroFindingsPath(t *testing.T) {
	// writeReportAndExit calls os.Exit when there ARE findings, which this
	// test process can't survive -- that end-to-end behavior is exercised
	// via exit code assertions in a real subprocess test elsewhere isn't
	// worth the complexity here; this covers the zero-findings path, which
	// is the only path that returns normally.
	t.Chdir(t.TempDir())
	out := report.Build(nil)
	writeReportAndExit(out)

	if _, err := os.Stat("ap-report.json"); err != nil {
		t.Errorf("expected ap-report.json to be written, stat failed: %v", err)
	}
}
