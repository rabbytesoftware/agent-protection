package conditionperline

import (
	"go/ast"
	"go/token"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule    = "R53"
	snippet = "if condition joins several conditions on one line: end each line with && or || and put the next condition on its own line"
)

func Conditionperline(
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
	for _, statement := range source.Nodes[*ast.IfStmt](file.Syntax) {
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
	statement *ast.IfStmt,
) (findings []report.Finding) {
	collapsed := collapsedJoins(
		file,
		statement,
	)
	if len(collapsed) == 0 {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: file.Location(statement.Cond),
		Snippet:  snippet,
	}}
}

func collapsedJoins(
	file source.File,
	statement *ast.IfStmt,
) (collapsed []*ast.BinaryExpr) {
	joins := slices.DeleteFunc(
		source.Nodes[*ast.BinaryExpr](statement.Cond),
		isNotLogical,
	)
	for _, join := range joins {
		collapsed = append(
			collapsed,
			collapsedJoin(
				file,
				join,
			)...,
		)
	}

	return collapsed
}

func collapsedJoin(
	file source.File,
	join *ast.BinaryExpr,
) (collapsed []*ast.BinaryExpr) {
	isSplit := lineOf(
		file,
		join.Y.Pos(),
	) > lineOf(
		file,
		join.X.End(),
	)
	if isSplit {
		return nil
	}

	return []*ast.BinaryExpr{join}
}

func isNotLogical(
	expression *ast.BinaryExpr,
) (notLogical bool) {
	return expression.Op != token.LAND && expression.Op != token.LOR
}

func lineOf(
	file source.File,
	position token.Pos,
) (line int) {
	return file.FileSet.Position(position).Line
}
