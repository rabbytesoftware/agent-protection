package comments

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rabbytesoftware/agent-protection/internal/report"
)

const (
	rule            = "R07"
	snippet         = "comment: delete it and make the code say it through names and structure"
	locationFormat  = "%s:%d"
	rustExtension   = ".rs"
	generatedWindow = 500
)

func Comments(
	root string,
	extensions []string,
) (findings []report.Finding, err error) {
	visitor := &walker{root: root, extensions: extensions}
	err = filepath.WalkDir(
		root,
		visitor.visit,
	)
	if err != nil {
		return nil,
			fmt.Errorf(
				"comments: %w",
				err,
			)
	}

	failures := []error{}
	for _, path := range visitor.files {
		fileFound, fileErr := fileFindings(
			root,
			path,
		)
		findings = append(
			findings,
			fileFound...,
		)
		failures = append(
			failures,
			fileErr,
		)
	}

	return findings,
		errors.Join(failures...)
}

func fileFindings(
	root string,
	path string,
) (findings []report.Finding, err error) {
	text, err := os.ReadFile(path)
	if err != nil {
		return nil,
			fmt.Errorf(
				"comments: %w",
				err,
			)
	}

	if isGenerated(text) {
		return nil,
			nil
	}

	rel, err := filepath.Rel(
		root,
		path,
	)
	if err != nil {
		return nil,
			fmt.Errorf(
				"comments: %w",
				err,
			)
	}

	return lineFindings(
			filepath.ToSlash(rel),
			commentLines(
				text,
				spansOf(
					path,
					text,
				),
			),
		),
		nil
}

func lineFindings(
	rel string,
	lines []int,
) (findings []report.Finding) {
	for _, line := range lines {
		findings = append(
			findings,
			report.Finding{
				Rule: rule,
				Location: fmt.Sprintf(
					locationFormat,
					rel,
					line,
				),
				Snippet: snippet,
			},
		)
	}

	return findings
}

func spansOf(
	path string,
	text []byte,
) (spans []span) {
	if filepath.Ext(path) == rustExtension {
		return rustSpans(text)
	}

	return scriptSpans(text)
}

func isGenerated(
	text []byte,
) (generated bool) {
	header := text[:min(
		len(text),
		generatedWindow,
	)]
	for _, marker := range generatedMarkers() {
		generated = generated || bytes.Contains(
			header,
			[]byte(marker),
		)
	}

	return generated
}

func generatedMarkers() (markers []string) {
	return []string{"Code generated", "@generated", "DO NOT EDIT"}
}

func ScriptExtensions() (extensions []string) {
	return []string{".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs", ".mts", ".cts"}
}

func RustExtensions() (extensions []string) {
	return []string{".rs"}
}
