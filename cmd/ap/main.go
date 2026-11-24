package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner"
)

const (
	checkCommand      = "check"
	reportPath        = "ap-report.json"
	reportFileMode    = 0o644
	cleanExitCode     = 0
	findingsExitCode  = 1
	usageExitCode     = 2
	expectedArguments = 3
)

func main() {
	root, err := rootArgument(os.Args)
	if err != nil {
		fmt.Fprintln(
			os.Stderr,
			err,
		)
		os.Exit(usageExitCode)
	}
	output, err := runner.Runner(root)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"check failed: %v\n",
			err,
		)
		os.Exit(findingsExitCode)
	}
	err = writeReport(output)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"write report: %v\n",
			err,
		)
		os.Exit(findingsExitCode)
	}
	fmt.Print(output.Text())
	os.Exit(exitCode(output))
}

func rootArgument(
	arguments []string,
) (root string, err error) {
	isCheckInvocation := len(arguments) == expectedArguments && arguments[1] == checkCommand
	if !isCheckInvocation {
		return "",
			errors.New("usage: ap check <root-dir>")
	}
	return arguments[2],
		nil
}

func writeReport(
	output report.Output,
) (err error) {
	document, err := output.JSON()
	if err != nil {
		return err
	}
	return os.WriteFile(
		reportPath,
		[]byte(document),
		reportFileMode,
	)
}

func exitCode(
	output report.Output,
) (code int) {
	if len(output.Findings) > 0 {
		return findingsExitCode
	}
	return cleanExitCode
}
