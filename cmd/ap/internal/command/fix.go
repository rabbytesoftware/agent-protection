package command

import (
	"fmt"
	"os"
)

type fix struct {
	rewrite func(root string) (changed []string, err error)
}

func Fix(
	rewrite func(root string) (changed []string, err error),
) (command *fix) {
	return &fix{rewrite: rewrite}
}

func (f *fix) Run(
	root string,
) (code int) {
	changed, err := f.rewrite(root)
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
