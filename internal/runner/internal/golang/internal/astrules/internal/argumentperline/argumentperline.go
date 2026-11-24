package argumentperline

import (
	"go/ast"
	"go/token"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule             = "R52"
	minimumArguments = 2
	snippet          = "call with several arguments: put every argument on its own line, after an opening paren that ends the line and before a closing paren on a line of its own"
)

// Argumentperline reports every call with two or more arguments whose
// arguments do not each sit on their own line between an opening paren
// ending its line and a closing paren on a line of its own. Calls with zero
// or one argument may stay inline.
func Argumentperline(
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
	for _, call := range source.Nodes[*ast.CallExpr](file.Syntax) {
		findings = append(
			findings,
			callFindings(
				file,
				call,
			)...,
		)
	}

	return findings
}

func callFindings(
	file source.File,
	call *ast.CallExpr,
) (findings []report.Finding) {
	isCompliant := len(call.Args) < minimumArguments || hasArgumentPerLine(
		file,
		call,
	)
	if isCompliant {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: file.Location(call),
		Snippet:  snippet,
	}}
}

func hasArgumentPerLine(
	file source.File,
	call *ast.CallExpr,
) (compliant bool) {
	spans := []span{pointSpan(
		file,
		call.Lparen,
	)}
	for _, extent := range argumentExtents(call) {
		spans = append(
			spans,
			span{first: lineOf(
				file,
				extent.Start,
			), last: lineOf(
				file,
				extent.End,
			)},
		)
	}

	spans = append(
		spans,
		pointSpan(
			file,
			call.Rparen,
		),
	)
	return !slices.ContainsFunc(
		gaps(spans),
		isCollapsed,
	)
}

func gaps(
	spans []span,
) (between []span) {
	for index := 1; index < len(spans); index++ {
		between = append(
			between,
			span{first: spans[index-1].last, last: spans[index].first},
		)
	}

	return between
}

func isCollapsed(
	gap span,
) (collapsed bool) {
	return gap.last <= gap.first
}

func pointSpan(
	file source.File,
	position token.Pos,
) (point span) {
	line := lineOf(
		file,
		position,
	)
	return span{first: line, last: line}
}

func lineOf(
	file source.File,
	position token.Pos,
) (line int) {
	return file.FileSet.Position(position).Line
}
