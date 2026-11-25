package component

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
)

const (
	GoMarker         = "go.mod"
	RustMarker       = "Cargo.toml"
	TypescriptMarker = "package.json"
	GoSource         = ".go"
	markerSeparator  = ", "
)

func Check(
	root string,
	environment func(string) string,
	checkers []Checker,
	comments func(string, []string) ([]report.Finding, error),
) (findings []report.Finding, err error) {
	index := markedIndex(
		root,
		checkers,
	)
	if index < 0 {
		return nil,
			unsupported(
				root,
				checkers,
			)
	}

	own, err := checkers[index].Check(
		root,
		environment,
	)
	if err != nil {
		return nil,
			err
	}

	commented, err := comments(
		root,
		checkers[index].Owned,
	)
	if err != nil {
		return nil,
			err
	}

	return append(
			own,
			commented...,
		),
		nil
}

func markedIndex(
	root string,
	checkers []Checker,
) (index int) {
	return slices.Index(
		presence(
			root,
			checkers,
		),
		true,
	)
}

func presence(
	root string,
	checkers []Checker,
) (present []bool) {
	for _, checker := range checkers {
		present = append(
			present,
			hasMarker(
				root,
				checker.Marker,
			),
		)
	}

	return present
}

func Fix(
	root string,
	fix func(string) ([]string, error),
) (changed []string, err error) {
	if !hasMarker(
		root,
		GoMarker,
	) {
		return nil,
			fmt.Errorf(
				"ap fix supports Go only: %s holds no %s",
				root,
				GoMarker,
			)
	}

	return fix(root)
}

func unsupported(
	root string,
	checkers []Checker,
) (err error) {
	markers := []string{}
	for _, checker := range checkers {
		markers = append(
			markers,
			checker.Marker,
		)
	}

	return fmt.Errorf(
		"no %s directly in %s: point ap at a single-language component folder",
		strings.Join(
			markers,
			markerSeparator,
		),
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
