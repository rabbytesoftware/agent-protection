package command

import (
	"fmt"
	"os"

	"github.com/rabbytesoftware/agent-protection/internal/report"
)

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
