package command

import (
	"fmt"
	"os"
)

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
