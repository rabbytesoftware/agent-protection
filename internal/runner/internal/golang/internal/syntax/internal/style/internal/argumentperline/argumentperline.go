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
	callSnippet      = "call with several arguments: put every argument on its own line, after an opening paren that ends the line and before a closing paren on a line of its own"
)

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

	return append(
		findings,
		literalFindings(file)...,
	)
}

func callFindings(
	file source.File,
	call *ast.CallExpr,
) (findings []report.Finding) {
	isCompliant := len(call.Args) < minimumArguments || isOnePerLine(
		file,
		call.Lparen,
		argumentExtents(call),
		call.Rparen,
	)
	if isCompliant {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: file.Location(call),
		Snippet:  callSnippet,
	}}
}

func isOnePerLine(
	file source.File,
	opening token.Pos,
	items []source.Extent,
	closing token.Pos,
) (compliant bool) {
	spans := []span{pointSpan(
		file,
		opening,
	)}
	for _, extent := range items {
		spans = append(
			spans,
			span{
				first: lineOf(
					file,
					extent.Start,
				),
				last: lineOf(
					file,
					extent.End,
				),
			},
		)
	}

	spans = append(
		spans,
		pointSpan(
			file,
			closing,
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
			span{
				first: spans[index-1].last,
				last:  spans[index].first,
			},
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
