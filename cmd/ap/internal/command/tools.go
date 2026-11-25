package command

import (
	"fmt"
	"os"

	"github.com/rabbytesoftware/agent-protection/internal/report"
)

type Tools struct {
	Run     func(root string) (findings []report.Finding, err error)
	Rewrite func(root string) (changed []string, err error)
	Docs    func() (docs map[string]report.RuleDoc, err error)
	Manual  func() (guide string)
}

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

func (t Tools) Fix(
	root string,
) (code int) {
	changed, err := t.Rewrite(root)
	for _, path := range changed {
		fmt.Println(path)
	}

	fmt.Printf(
		"files changed: %d\n",
		len(changed),
	)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"fix failed: %v\n",
			err,
		)
		return findingsExitCode
	}

	return cleanExitCode
}

func (t Tools) Guide() (code int) {
	fmt.Print(t.Manual())
	return cleanExitCode
}

func (t Tools) Markdown(
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
