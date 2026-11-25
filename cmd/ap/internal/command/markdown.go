package command

import (
	"fmt"
	"os"

	"github.com/rabbytesoftware/agent-protection/internal/report"
)

type markdown struct {
	read func(path string) (document []byte, err error)
}

func Markdown(
	read func(path string) (document []byte, err error),
) (command *markdown) {
	return &markdown{read: read}
}

func (m *markdown) Run(
	path string,
) (code int) {
	document, err := m.read(path)
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
