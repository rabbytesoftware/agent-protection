package clippy

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

const (
	cargo        = "cargo"
	rule         = "RUST-PANIC-RISK"
	testFragment = "test"
	allMatches   = -1
)

const (
	wholeMatch = iota
	pathGroup
	lineGroup
	textGroup
)

func Clippy(
	root string,
) (warnings []Warning, err error) {
	command := exec.Command(
		cargo,
		"clippy",
		"--all-targets",
		"--message-format=short",
		"--",
		"-W",
		"clippy::unwrap_used",
		"-W",
		"clippy::expect_used",
		"-W",
		"clippy::panic",
	)
	command.Dir = root
	output, err := command.CombinedOutput()
	if err != nil {
		return nil,
			fmt.Errorf(
				"cargo clippy in %s: %w\n%s",
				root,
				err,
				output,
			)
	}

	return parse(string(output))
}

func parse(
	output string,
) (warnings []Warning, err error) {
	matches := slices.DeleteFunc(
		warningPattern().FindAllStringSubmatch(
			output,
			allMatches,
		),
		isIgnored,
	)
	failures := []error{}
	for _, match := range matches {
		warning, convertErr := newWarning(match)
		warnings = append(
			warnings,
			warning,
		)
		failures = append(
			failures,
			convertErr,
		)
	}

	err = errors.Join(failures...)
	if err != nil {
		return nil,
			err
	}

	return warnings,
		nil
}

func newWarning(
	match []string,
) (warning Warning, err error) {
	line, err := strconv.Atoi(match[lineGroup])
	if err != nil {
		return Warning{},
			fmt.Errorf(
				"clippy line number in %q: %w",
				match[wholeMatch],
				err,
			)
	}

	return Warning{Path: match[pathGroup], Line: line, Rule: rule, Text: match[textGroup]},
		nil
}

func isIgnored(
	match []string,
) (ignored bool) {
	return strings.Contains(
		match[pathGroup],
		testFragment,
	) || !panicRiskPattern().MatchString(match[textGroup])
}

func warningPattern() (pattern *regexp.Regexp) {
	return regexp.MustCompile(`(?m)^(.+?\.rs):(\d+):\d+: warning: (.+?)\r?$`)
}

func panicRiskPattern() (pattern *regexp.Regexp) {
	return regexp.MustCompile("`(unwrap|expect|panic)(_err)?\\(|(unwrap|expect|panic)` should not be present")
}
