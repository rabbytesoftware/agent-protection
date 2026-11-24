package runner

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/rust"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/typescript"
)

const (
	goMarker         = "go.mod"
	rustMarker       = "Cargo.toml"
	typescriptMarker = "package.json"
)

// Runner checks the single-language component at root against its rule set
// and returns the report. The language comes from the marker file directly
// inside root: go.mod, Cargo.toml or package.json, in that order.
func Runner(
	root string,
) (output report.Output, err error) {
	check, err := checkFor(root)
	if err != nil {
		return report.Output{},
			err
	}
	findings, err := check(root)
	if err != nil {
		return report.Output{},
			err
	}
	return report.Report(findings)
}

func checkFor(
	root string,
) (check func(string) ([]report.Finding, error), err error) {
	switch {
	case hasMarker(
		root,
		goMarker,
	):
		return golang.Golang,
			nil
	case hasMarker(
		root,
		rustMarker,
	):
		return rust.Rust,
			nil
	case hasMarker(
		root,
		typescriptMarker,
	):
		return typescript.Typescript,
			nil
	}
	return nil,
		fmt.Errorf(
			"no %s, %s or %s directly in %s: point ap at a single-language component folder",
			goMarker,
			rustMarker,
			typescriptMarker,
			root,
		)
}

func hasMarker(
	root string,
	marker string,
) (present bool) {
	_, err := os.Stat(filepath.Join(
		root,
		marker,
	))
	return err == nil
}
