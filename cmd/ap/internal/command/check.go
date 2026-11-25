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

func (t Tools) Check(
	root string,
) (code int) {
	output, err := t.checked(root)
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
		fmt.Print("\n" + t.Manual())
	}

	return code
}

func (t Tools) checked(
	root string,
) (output report.Output, err error) {
	findings, err := t.Run(root)
	if err != nil {
		return report.Output{},
			err
	}

	registry, err := t.Docs()
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
