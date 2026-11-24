package droppederror

import (
	"go/ast"
	"path/filepath"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const rule = "R26"

// Droppederror reports failures that are observed and then dropped (R26):
// an error compared with nil whose if block leaves without using it, an
// error or recovered panic assigned to the blank identifier, and a
// recover() whose value is discarded or only compared with nil. With no
// type checker, an error is a variable named err or ending in Err, one
// declared as error, or one bound to the error result of a function or
// method declared in the same directory.
func Droppederror(
	files []source.File,
) (findings []report.Finding) {
	callees := erroringCallees(files)
	for _, file := range files {
		findings = append(
			findings,
			fileFindings(
				file,
				callees[filepath.Dir(file.Path)],
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
		findings = append(
			findings,
			checkedFindings(
				file,
				function,
				callees,
			)...,
		)
	}

	return slices.Concat(
		findings,
		blankFindings(
			file,
			callees,
		),
		recoverFindings(file),
	)
}

func finding(
	file source.File,
	node ast.Node,
	snippet string,
) (built report.Finding) {
	return report.Finding{Rule: rule, Location: file.Location(node), Snippet: snippet}
}
