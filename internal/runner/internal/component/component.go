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
	markerSeparator  = ", "
)

func Check(
	root string,
	checkers []Checker,
) (findings []report.Finding, err error) {
	index := slices.IndexFunc(
		checkers,
		func(
			checker Checker,
		) (present bool) {
			return hasMarker(
				root,
				checker.Marker,
			)
		},
	)
	if index < 0 {
		return nil,
			unsupported(
				root,
				checkers,
			)
	}

	return checkers[index].Check(root)
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
