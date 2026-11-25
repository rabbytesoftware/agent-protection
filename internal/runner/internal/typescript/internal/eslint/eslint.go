package eslint

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	bun         = "bun"
	jsonOpeners = "[{"
)

func Eslint(
	root string,
	arguments ...string,
) (results []FileResult, err error) {
	output, err := run(
		root,
		arguments,
	)
	if err != nil {
		return nil,
			err
	}

	return decode(
		root,
		output,
	)
}

func run(
	root string,
	arguments []string,
) (output []byte, err error) {
	commandArguments := append(
		[]string{
			"x",
			"eslint",
			".",
			"--format=json",
		},
		arguments...,
	)
	binary, err := bunBinary()
	if err != nil {
		return nil,
			err
	}

	command := exec.Command(
		binary,
		commandArguments...,
	)
	command.Dir = root
	output, err = command.CombinedOutput()
	var exitError *exec.ExitError
	couldNotRun := err != nil && !errors.As(
		err,
		&exitError,
	)
	if couldNotRun {
		return nil,
			fmt.Errorf(
				"run eslint through %s: %w",
				bun,
				err,
			)
	}

	return output,
		nil
}

func decode(
	root string,
	output []byte,
) (results []FileResult, err error) {
	err = json.Unmarshal(
		jsonPayload(output),
		&results,
	)
	if err != nil {
		return nil,
			fmt.Errorf(
				"eslint printed no JSON report (is eslint installed?): %w\n%s",
				err,
				output,
			)
	}

	failures := []error{}
	for index := range results {
		rel, relErr := relPath(
			root,
			results[index].FilePath,
		)
		results[index].FilePath = rel
		failures = append(
			failures,
			relErr,
		)
	}

	err = errors.Join(failures...)
	if err != nil {
		return nil,
			err
	}

	return results,
		nil
}

func bunBinary() (binary string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "",
			fmt.Errorf(
				"locate %s: %w",
				bun,
				err,
			)
	}

	candidate := filepath.Join(
		home,
		"."+bun,
		"bin",
		bun,
	)
	_, err = os.Stat(candidate)
	if errors.Is(
		err,
		fs.ErrNotExist,
	) {
		return bun,
			nil
	}

	if err != nil {
		return "",
			fmt.Errorf(
				"locate %s: %w",
				bun,
				err,
			)
	}

	return candidate,
		nil
}

func jsonPayload(
	output []byte,
) (payload []byte) {
	start := strings.IndexAny(
		string(output),
		jsonOpeners,
	)
	if start < 0 {
		return output
	}

	return output[start:]
}

func relPath(
	root string,
	path string,
) (rel string, err error) {
	rel, err = filepath.Rel(
		root,
		path,
	)
	if err != nil {
		return "",
			fmt.Errorf(
				"eslint file path %s: %w",
				path,
				err,
			)
	}

	return rel,
		nil
}
