package command

import (
	"fmt"
	"os"

	"github.com/rabbytesoftware/agent-protection/internal/report"
)

const (
	reportPath       = "ap-report.json"
	reportFileMode   = 0o644
	cleanExitCode    = 0
	findingsExitCode = 1
)

type check struct {
	run         func(root string, environment func(string) string) (findings []report.Finding, err error)
	docs        func() (docs map[string]report.RuleDoc, err error)
	manual      func() (guide string)
	environment func(key string) (value string)
}

func Check(
	run func(root string, environment func(string) string) (findings []report.Finding, err error),
	docs func() (docs map[string]report.RuleDoc, err error),
	manual func() (guide string),
	environment func(key string) (value string),
) (command *check) {
	return &check{
		run:         run,
		docs:        docs,
		manual:      manual,
		environment: environment,
	}
}

func (c *check) Run(
	root string,
) (code int) {
	output, err := c.checked(root)
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
	code = exitCode(output)
	if code != cleanExitCode {
		fmt.Print("\n" + c.manual())
	}

	return code
}

func (c *check) checked(
	root string,
) (output report.Output, err error) {
	findings, err := c.run(
		root,
		c.environment,
	)
	if err != nil {
		return report.Output{},
			err
	}

	registry, err := c.docs()
	if err != nil {
		return report.Output{},
			fmt.Errorf(
				"load rule registry: %w",
				err,
			)
	}

	return report.Report(
			findings,
			registry,
		),
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
