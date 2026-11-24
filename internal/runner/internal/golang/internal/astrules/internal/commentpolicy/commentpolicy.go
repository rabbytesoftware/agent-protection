package commentpolicy

import (
	"go/ast"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule            = "R07"
	snippet         = "comment: delete it and make the code say it through names and structure"
	cgoImport       = `"C"`
	examplePrefix   = "Example"
	outputMarker    = "Output:"
	unorderedMarker = "Unordered output:"
	lineMarker      = "//"
)

func Commentpolicy(
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
	exempt := exemptGroups(file)
	for _, group := range file.Syntax.Comments {
		findings = append(
			findings,
			groupFindings(
				file,
				group,
				exempt,
			)...,
		)
	}

	return findings
}

func groupFindings(
	file source.File,
	group *ast.CommentGroup,
	exempt []*ast.CommentGroup,
) (findings []report.Finding) {
	isProse := !slices.Contains(
		exempt,
		group,
	) && slices.ContainsFunc(
		group.List,
		isProseComment,
	)
	if !isProse {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: file.Location(group),
		Snippet:  snippet,
	}}
}

func exemptGroups(
	file source.File,
) (groups []*ast.CommentGroup) {
	for _, declaration := range file.Syntax.Decls {
		groups = append(
			groups,
			cgoPreamble(declaration)...,
		)
	}

	for _, function := range source.Nodes[*ast.FuncDecl](file.Syntax) {
		groups = append(
			groups,
			exampleOutputs(
				file,
				function,
			)...,
		)
	}

	return groups
}

func cgoPreamble(
	declaration ast.Decl,
) (groups []*ast.CommentGroup) {
	general, isGeneral := declaration.(*ast.GenDecl)
	isCgo := isGeneral && general.Doc != nil && slices.ContainsFunc(
		source.Nodes[*ast.ImportSpec](general),
		importsC,
	)
	if !isCgo {
		return nil
	}

	return []*ast.CommentGroup{general.Doc}
}

func importsC(
	spec *ast.ImportSpec,
) (matches bool) {
	return spec.Path.Value == cgoImport
}

func exampleOutputs(
	file source.File,
	function *ast.FuncDecl,
) (groups []*ast.CommentGroup) {
	isExample := source.IsTest(file.Path) &&
		function.Recv == nil &&
		function.Body != nil &&
		strings.HasPrefix(
			function.Name.Name,
			examplePrefix,
		)
	if !isExample {
		return nil
	}

	for _, group := range file.Syntax.Comments {
		groups = appendOutput(
			groups,
			group,
			function.Body,
		)
	}

	return groups
}

func appendOutput(
	groups []*ast.CommentGroup,
	group *ast.CommentGroup,
	body *ast.BlockStmt,
) (extended []*ast.CommentGroup) {
	if !isOutputWithin(
		group,
		body,
	) {
		return groups
	}

	return append(
		groups,
		group,
	)
}

func isOutputWithin(
	group *ast.CommentGroup,
	body *ast.BlockStmt,
) (within bool) {
	return group.Pos() > body.Lbrace &&
		group.End() < body.Rbrace &&
		startsWithOutput(group)
}

func startsWithOutput(
	group *ast.CommentGroup,
) (matches bool) {
	first := strings.TrimSpace(strings.TrimPrefix(
		group.List[0].Text,
		lineMarker,
	))
	return strings.HasPrefix(
		first,
		outputMarker,
	) || strings.HasPrefix(
		first,
		unorderedMarker,
	)
}

func isProseComment(
	comment *ast.Comment,
) (prose bool) {
	prose = true
	for _, prefix := range directivePrefixes() {
		prose = prose && !strings.HasPrefix(
			comment.Text,
			prefix,
		)
	}

	return prose
}

func directivePrefixes() (prefixes []string) {
	return []string{"//go:", "//line ", "//export ", "//extern ", "// +build"}
}
