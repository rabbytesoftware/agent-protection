package main

import (
	"fmt"
	"os"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner"
)

const (
	minArgCount       = 3
	rootArgIndex      = 2
	reportFileMode    = 0o644
	usageExitCode     = 2
	checkFailExitCode = 1
)

// main is a thin CLI shell: parse arguments, delegate everything else to
// the runner package (AP's real composition root), print and exit.
func main() {
	if len(os.Args) < minArgCount || os.Args[1] != "check" {
		fmt.Fprintln(os.Stderr, "usage: ap check <root-dir>")
		os.Exit(usageExitCode)
	}
	root := os.Args[rootArgIndex]

	out, err := runner.Run(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "check failed: %v\n", err)
		os.Exit(checkFailExitCode)
	}

	writeReportAndExit(out)
}

// writeReportAndExit writes the report to ap-report.json, prints it, and
// exits non-zero if any findings were reported.
func writeReportAndExit(out report.Output) {
	j, _ := out.JSON()
	if werr := os.WriteFile("ap-report.json", []byte(j), reportFileMode); werr != nil {
		fmt.Fprintf(os.Stderr, "write report: %v\n", werr)
	}
	fmt.Print(out.Text())
	if len(out.Findings) > 0 {
		os.Exit(checkFailExitCode)
	}
}
