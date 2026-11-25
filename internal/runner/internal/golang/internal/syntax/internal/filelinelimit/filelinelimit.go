package filelinelimit

import (
	"fmt"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule     = "R44"
	maxLines = 500
)

func Filelinelimit(
	files []source.File,
) (findings []report.Finding) {
	for _, file := range files {
		findings = append(
			findings,
			fileFindings(file)...,
		)
	}

	return findings
}

func fileFindings(
	file source.File,
) (findings []report.Finding) {
	lines := file.FileSet.File(file.Syntax.Pos()).LineCount()
	if lines <= maxLines {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: file.Location(file.Syntax),
		Snippet: fmt.Sprintf(
			"%d lines, over the %d line limit: split the file by responsibility",
			lines,
			maxLines,
		),
	}}
}
