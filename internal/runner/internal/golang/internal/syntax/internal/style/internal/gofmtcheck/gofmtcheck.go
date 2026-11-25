package gofmtcheck

import (
	"bytes"
	"go/format"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule        = "R62"
	firstLine   = ":1"
	unformatted = "file is not gofmt-formatted: run gofmt (or ap fix)"
)

func Gofmtcheck(
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
	canonical, err := format.Source(file.Content)
	isFormatted := err != nil || bytes.Equal(
		canonical,
		file.Content,
	)
	if isFormatted {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: file.Rel + firstLine,
		Snippet:  unformatted,
	}}
}
