package scan

import (
	"bytes"
	"os/exec"
	"slices"
	"strings"
)

const pathSeparator = "\x00"

func unignored(
	root string,
	files []source,
) (kept []source) {
	ignored := ignoredPaths(
		root,
		files,
	)
	for _, file := range files {
		kept = append(
			kept,
			unlisted(
				file,
				ignored,
			)...,
		)
	}

	return kept
}

func unlisted(
	file source,
	ignored []string,
) (kept []source) {
	if slices.Contains(
		ignored,
		file.path,
	) {
		return nil
	}

	return []source{file}
}

func ignoredPaths(
	root string,
	files []source,
) (ignored []string) {
	paths := []string{}
	for _, file := range files {
		paths = append(
			paths,
			file.path,
		)
	}

	command := exec.Command(
		gitBinary,
		"-C",
		root,
		"check-ignore",
		"-z",
		"--stdin",
	)
	command.Stdin = strings.NewReader(strings.Join(
		paths,
		pathSeparator,
	))
	output := &bytes.Buffer{}
	command.Stdout = output
	hasIgnored := command.Run() == nil
	if !hasIgnored {
		return nil
	}

	return strings.Split(
		strings.TrimSuffix(
			output.String(),
			pathSeparator,
		),
		pathSeparator,
	)
}
