package dispatch

import (
	"errors"
	"fmt"
	"os"
	"slices"
)

const (
	checkCommand      = "check"
	fixCommand        = "fix"
	markdownCommand   = "markdown"
	guideCommand      = "guide"
	guideArguments    = 2
	usageExitCode     = 2
	expectedArguments = 3
	usage             = "usage: ap check <root-dir> | ap fix <root-dir> | ap markdown <ap-report.json> | ap guide"
)

func Dispatch(
	arguments []string,
	check Command,
	fix Command,
	markdown Command,
	guide Command,
) (code int) {
	isGuide := len(arguments) == guideArguments && arguments[1] == guideCommand
	if isGuide {
		return guide.Run("")
	}

	command, target, err := commandArguments(arguments)
	if err != nil {
		fmt.Fprintln(
			os.Stderr,
			err,
		)
		return usageExitCode
	}

	switch command {
	case markdownCommand:
		return markdown.Run(target)
	case fixCommand:
		return fix.Run(target)
	}

	return check.Run(target)
}

func commandArguments(
	arguments []string,
) (command string, target string, err error) {
	isKnownCommand := len(arguments) == expectedArguments && slices.Contains(
		[]string{
			checkCommand,
			fixCommand,
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
