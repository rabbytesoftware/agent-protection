package noelse

import (
	"go/ast"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const rule = "R03"

func Noelse(
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
	statements := slices.DeleteFunc(
		source.Nodes[*ast.IfStmt](file.Syntax),
		hasNoElse,
	)
	for _, statement := range statements {
		findings = append(
			findings,
			report.Finding{
				Rule:     rule,
				Location: file.Location(statement.Else),
				Snippet:  "else branch: invert the condition into a guard clause that returns early",
			},
		)
	}

	return findings
}

func hasNoElse(
	statement *ast.IfStmt,
) (noElse bool) {
	return statement.Else == nil
}
