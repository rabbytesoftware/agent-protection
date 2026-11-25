package droppederror

import (
	"go/ast"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const rule = "R26"

func Droppederror(
	files []source.File,
) (findings []report.Finding) {
	callees := erroringCallees(files)
	directories := relativeDirectories(files)
	for _, file := range files {
		findings = append(
			findings,
			fileFindings(
				file,
				knownCallees(
					file,
					callees,
					directories,
				),
			)...,
		)
	}

	return findings
}

func fileFindings(
	file source.File,
	callees map[string]int,
) (findings []report.Finding) {
	for _, function := range source.Nodes[*ast.FuncDecl](file.Syntax) {
		findings = slices.Concat(
			findings,
			checkedFindings(
				file,
				function,
				callees,
			),
			loggedFindings(
				file,
				function,
				callees,
			),
		)
	}

	return slices.Concat(
		findings,
		blankFindings(
			file,
			callees,
		),
		wrappingFindings(
			file,
			callees,
		),
		recoverFindings(file),
		inlineFindings(
			file,
			callees,
		),
		ignoredFindings(file),
	)
}

func finding(
	file source.File,
	node ast.Node,
	snippet string,
) (built report.Finding) {
	return report.Finding{
		Rule:     rule,
		Location: file.Location(node),
		Snippet:  snippet,
	}
}
