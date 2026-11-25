package returnperline

import (
	"go/ast"
	"go/token"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule          = "R54"
	minimumValues = 2
	snippet       = "return with several values: keep the first value on the return line and put every following value on its own line"
)

func Returnperline(
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
		source.Nodes[*ast.ReturnStmt](file.Syntax),
		hasSingleValue,
	)
	for _, statement := range statements {
		findings = append(
			findings,
			statementFindings(
				file,
				statement,
			)...,
		)
	}

	return findings
}

func statementFindings(
	file source.File,
	statement *ast.ReturnStmt,
) (findings []report.Finding) {
	collapsed := []ast.Expr{}
	for index := 1; index < len(statement.Results); index++ {
		collapsed = append(
			collapsed,
			collapsedValue(
				file,
				statement.Results[index-1],
				statement.Results[index],
			)...,
		)
	}

	if len(collapsed) == 0 {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: file.Location(statement),
		Snippet:  snippet,
	}}
}

func collapsedValue(
	file source.File,
	previous ast.Expr,
	value ast.Expr,
) (collapsed []ast.Expr) {
	isOwnLine := lineOf(
		file,
		value.Pos(),
	) > lineOf(
		file,
		previous.End(),
	)
	if isOwnLine {
		return nil
	}

	return []ast.Expr{value}
}

func hasSingleValue(
	statement *ast.ReturnStmt,
) (single bool) {
	return len(statement.Results) < minimumValues
}

func lineOf(
	file source.File,
	position token.Pos,
) (line int) {
	return file.FileSet.Position(position).Line
}
