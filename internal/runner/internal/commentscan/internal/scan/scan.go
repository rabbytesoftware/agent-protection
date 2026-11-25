package scan

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
)

const (
	rule            = "R07"
	snippet         = "comment: delete it and make the code say it through names and structure"
	locationFormat  = "%s:%d"
	generatedWindow = 500
	gitBinary       = "git"
)

func Scan(
	root string,
	excluded []string,
	languages []lexeme.Language,
	fallback lexeme.Language,
) (findings []report.Finding, err error) {
	visitor := &walker{
		root:      root,
		excluded:  excluded,
		languages: languages,
		fallback:  fallback,
	}
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

	return filesFindings(
		root,
		unignored(
			root,
			visitor.files,
		),
	)
}

func filesFindings(
	root string,
	files []source,
) (findings []report.Finding, err error) {
	failures := []error{}
	for _, file := range files {
		fileFound, fileErr := fileFindings(
			root,
			file,
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
	file source,
) (findings []report.Finding, err error) {
	text, err := os.ReadFile(file.path)
	if err != nil {
		return nil,
			fmt.Errorf(
				"comments: %w",
				err,
			)
	}

	language, isLexed := file.resolve(text)
	isExempt := !isLexed ||
		(isGenerated(text) && hasGeneratedName(file.path))
	if isExempt {
		return nil,
			nil
	}

	rel, err := filepath.Rel(
		root,
		file.path,
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
			lexeme.Lines(
				text,
				language.Spans(text),
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

func hasGeneratedName(
	path string,
) (generated bool) {
	name := filepath.Base(path)
	for _, marker := range generatedNameMarkers() {
		generated = generated || strings.Contains(
			name,
			marker,
		)
	}

	return generated
}

func generatedNameMarkers() (markers []string) {
	return []string{
		".pb.",
		"_pb2.",
		"_pb2_grpc.",
		"_pb.",
		"_grpc_pb.",
		".min.",
		".generated.",
		".gen.",
		"_gen.",
		"_generated.",
	}
}
