package filelinelimit

import (
	"fmt"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule         = "R44"
	maxLines     = 500
	sourceFormat = "%d lines, over the %d line limit: split the file by responsibility"
	testFormat   = "%d lines, over the %d line limit, test files included: a test file this long means its source file holds more than one responsibility; split the source by responsibility and move each group of tests into the x_test.go of the file it exercises (benchmarks may go to x_bench_test.go)"
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
			formatFor(file.Rel),
			lines,
			maxLines,
		),
	}}
}

func formatFor(
	rel string,
) (format string) {
	if source.IsTest(rel) {
		return testFormat
	}

	return sourceFormat
}
