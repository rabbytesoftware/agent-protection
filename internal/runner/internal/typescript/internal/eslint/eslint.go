package eslint

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	bun         = "bun"
	jsonOpeners = "[{"
)

// Eslint runs eslint through bun over root with any extra arguments and
// returns its per-file JSON report, every FilePath made relative to root.
// A non-zero exit still yields the report; it fails when bun cannot start
// or the output holds no parseable JSON report.
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
		[]string{"x", "eslint", ".", "--format=json"},
		arguments...,
	)
	command := exec.Command(
		bunBinary(),
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
	for index := range results {
		results[index].FilePath = relPath(
			root,
			results[index].FilePath,
		)
	}
	return results,
		nil
}

func bunBinary() (binary string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return bun
	}
	candidate := filepath.Join(
		home,
		"."+bun,
		"bin",
		bun,
	)
	_, err = os.Stat(candidate)
	if err != nil {
		return bun
	}
	return candidate
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
) (rel string) {
	rel, err := filepath.Rel(
		root,
		path,
	)
	if err != nil {
		return path
	}
	return rel
}
