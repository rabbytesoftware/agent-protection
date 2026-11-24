package main

import (
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner"
)

const (
	checkCommand      = "check"
	markdownCommand   = "markdown"
	reportPath        = "ap-report.json"
	reportFileMode    = 0o644
	cleanExitCode     = 0
	findingsExitCode  = 1
	usageExitCode     = 2
	expectedArguments = 3
	usage             = "usage: ap check <root-dir> | ap markdown <ap-report.json>"
)

func main() {
	os.Exit(run(os.Args))
}

func run(
	arguments []string,
) (code int) {
	command, target, err := commandArguments(arguments)
	if err != nil {
		fmt.Fprintln(
			os.Stderr,
			err,
		)
		return usageExitCode
	}

	if command == markdownCommand {
		return markdown(target)
	}

	return check(target)
}

func commandArguments(
	arguments []string,
) (command string, target string, err error) {
	isKnownCommand := len(arguments) == expectedArguments && slices.Contains(
		[]string{
			checkCommand,
			markdownCommand,
		},
		arguments[1],
	)
	if !isKnownCommand {
		return "",
			"",
			errors.New(usage)
	}

	return arguments[1],
		arguments[2],
		nil
}

func check(
	root string,
) (code int) {
	output, err := runner.Runner(root)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"check failed: %v\n",
			err,
		)
		return findingsExitCode
	}

	err = writeReport(output)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"write report: %v\n",
			err,
		)
		return findingsExitCode
	}

	fmt.Print(output.Text())
	return exitCode(output)
}

func markdown(
	path string,
) (code int) {
	document, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"read report: %v\n",
			err,
		)
		return findingsExitCode
	}

	output, err := report.Decode(document)
	if err != nil {
		fmt.Fprintln(
			os.Stderr,
			err,
		)
		return findingsExitCode
	}

	fmt.Print(output.Markdown())
	return cleanExitCode
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
