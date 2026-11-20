package main

import (
	"os"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
)

func TestRootArgument(
	t *testing.T,
) {
	cases := []struct {
		arguments []string
		wantRoot  string
		wantErr   bool
	}{
		{arguments: []string{"ap", "check", "."}, wantRoot: "."},
		{arguments: []string{"ap", "check"}, wantErr: true},
		{arguments: []string{"ap", "lint", "."}, wantErr: true},
		{arguments: []string{"ap", "check", ".", "extra"}, wantErr: true},
	}
	for _, tc := range cases {
		assertRootArgument(
			t,
			tc.arguments,
			tc.wantRoot,
			tc.wantErr,
		)
	}
}

func TestWriteReport(
	t *testing.T,
) {
	t.Chdir(t.TempDir())
	err := writeReport(report.Output{})
	if err != nil {
		t.Fatalf(
			"writeReport() error = %v",
			err,
		)
	}
	_, err = os.Stat(reportPath)
	if err != nil {
		t.Errorf(
			"report file missing: %v",
			err,
		)
	}
}

func TestExitCode(
	t *testing.T,
) {
	clean := exitCode(report.Output{})
	failing := exitCode(report.Output{Findings: []report.Finding{{Rule: "R03"}}})
	isCorrect := clean == cleanExitCode && failing == findingsExitCode
	if !isCorrect {
		t.Errorf(
			"exitCode() = %d clean and %d failing, want %d and %d",
			clean,
			failing,
			cleanExitCode,
			findingsExitCode,
		)
	}
}

func assertRootArgument(
	t *testing.T,
	arguments []string,
	wantRoot string,
	wantErr bool,
) {
	t.Helper()
	root, err := rootArgument(arguments)
	isMismatch := root != wantRoot || (err != nil) != wantErr
	if isMismatch {
		t.Errorf(
			"rootArgument(%v) = %q, %v; want %q, error %v",
			arguments,
			root,
			err,
			wantRoot,
			wantErr,
		)
	}
}
