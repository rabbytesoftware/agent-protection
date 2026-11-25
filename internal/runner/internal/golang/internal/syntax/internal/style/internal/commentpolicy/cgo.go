package commentpolicy

import (
	"go/ast"
	"regexp"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	preambleSnippet      = "C comment or disabled block in the cgo preamble: delete it; the preamble holds only preprocessor lines and C code"
	cLineCommentOpener   = "//"
	cBlockCommentOpener  = "/*"
	cBlockCommentCloser  = "*/"
	disabledBlockPattern = `^\s*#\s*if\s+0\b`
	cQuotes              = "'\""
	cEscape              = '\\'
	cEscapeWidth         = 2
	preambleLineBreak    = "\n"
)

func preambleFindings(
	file source.File,
) (findings []report.Finding) {
	for _, group := range cgoPreambles(file) {
		findings = append(
			findings,
			groupPreambleFindings(
				file,
				group,
			)...,
		)
	}

	return findings
}

func cgoPreambles(
	file source.File,
) (groups []*ast.CommentGroup) {
	for _, declaration := range file.Syntax.Decls {
		groups = append(
			groups,
			cgoPreamble(declaration)...,
		)
	}

	return groups
}

func groupPreambleFindings(
	file source.File,
	group *ast.CommentGroup,
) (findings []report.Finding) {
	for _, comment := range group.List {
		findings = append(
			findings,
			commentPreambleFindings(
				file,
				comment,
			)...,
		)
	}

	return findings
}

func commentPreambleFindings(
	file source.File,
	comment *ast.Comment,
) (findings []report.Finding) {
	if !slices.ContainsFunc(
		strings.Split(
			preambleBody(comment.Text),
			preambleLineBreak,
		),
		isNotCode,
	) {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: file.Location(comment),
		Snippet:  preambleSnippet,
	}}
}

func preambleBody(
	text string,
) (body string) {
	if strings.HasPrefix(
		text,
		cLineCommentOpener,
	) {
		return text[len(cLineCommentOpener):]
	}

	return strings.TrimSuffix(
		strings.TrimPrefix(
			text,
			cBlockCommentOpener,
		),
		cBlockCommentCloser,
	)
}

func isNotCode(
	line string,
) (notCode bool) {
	return regexp.MustCompile(disabledBlockPattern).MatchString(line) ||
		hasCComment(line)
}

func hasCComment(
	line string,
) (found bool) {
	cursor := 0
	for cursor < len(line) &&
		!opensCComment(
			line[cursor:],
		) {
		cursor = cTokenEnd(
			line,
			cursor,
		)
	}

	return cursor < len(line)
}

func opensCComment(
	rest string,
) (opens bool) {
	return strings.HasPrefix(
		rest,
		cLineCommentOpener,
	) || strings.HasPrefix(
		rest,
		cBlockCommentOpener,
	)
}

func cTokenEnd(
	line string,
	at int,
) (end int) {
	if strings.IndexByte(
		cQuotes,
		line[at],
	) < 0 {
		return at + 1
	}

	end = at + 1
	for end < len(line) &&
		line[end] != line[at] {
		end = cEscapedNext(
			line,
			end,
		)
	}

	return min(
		end+1,
		len(line),
	)
}

func cEscapedNext(
	line string,
	at int,
) (next int) {
	if line[at] == cEscape {
		return at + cEscapeWidth
	}

	return at + 1
}
