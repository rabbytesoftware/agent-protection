package commentpolicy

import (
	"fmt"
	"go/ast"
	"go/token"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule              = "R07"
	commentSnippet    = "%s: only exported declarations carry comments, as their doc comment; delete it and let names carry the meaning"
	missingDocSnippet = "exported %s has no doc comment"
	directivePrefix   = "//go:"
	lineDirective     = "//line "
	nolintPrefix      = "//nolint"
	legacyBuildPrefix = "// +build"
	generatedPrefix   = "// Code generated "
	generatedSuffix   = " DO NOT EDIT."
	lineComment       = "//"
	blockOpening      = "/*"
	blockClosing      = "*/"
	deprecatedPrefix  = "Deprecated:"
	plainLink         = "http://"
	secureLink        = "https://"
	testPrefix        = "Test"
	benchmarkPrefix   = "Benchmark"
	fuzzPrefix        = "Fuzz"
	examplePrefix     = "Example"
	testingPackage    = "testing"
	newline           = "\n"
)

// Commentpolicy reports every comment that is not the doc comment of an
// exported declaration, field or interface method, except directives,
// build or generated headers, deprecations and spec links, and every
// exported declaration missing one. Test functions need no doc comment.
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
	allowed := allowedDocs(file)
	for _, group := range file.Syntax.Comments {
		findings = append(
			findings,
			groupFindings(
				file,
				allowed,
				group,
			)...,
		)
	}

	for _, decl := range file.Syntax.Decls {
		findings = append(
			findings,
			undocumentedFindings(
				file,
				decl,
			)...,
		)
	}

	return findings
}

func groupFindings(
	file source.File,
	allowed []*ast.CommentGroup,
	group *ast.CommentGroup,
) (findings []report.Finding) {
	isPermitted := slices.Contains(
		allowed,
		group,
	) || isFileHeader(
		file.Syntax,
		group,
	) || !slices.ContainsFunc(
		group.List,
		isRestricted,
	)
	isSuppression := slices.ContainsFunc(
		group.List,
		isNolint,
	)
	if isPermitted &&
		!isSuppression {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: file.Location(group),
		Snippet: fmt.Sprintf(
			commentSnippet,
			strings.Split(
				group.List[0].Text,
				newline,
			)[0],
		),
	}}
}

func allowedDocs(
	file source.File,
) (groups []*ast.CommentGroup) {
	for _, decl := range file.Syntax.Decls {
		groups = append(
			groups,
			declDocs(
				file,
				decl,
			)...,
		)
	}

	return groups
}

func declDocs(
	file source.File,
	decl ast.Decl,
) (groups []*ast.CommentGroup) {
	switch typed := decl.(type) {
	case *ast.FuncDecl:
		return funcDocs(
			file,
			typed,
		)
	case *ast.GenDecl:
		return genDeclDocs(typed)
	}

	return nil
}

func funcDocs(
	file source.File,
	decl *ast.FuncDecl,
) (groups []*ast.CommentGroup) {
	if !needsDoc(
		file,
		decl,
	) {
		return nil
	}

	return []*ast.CommentGroup{decl.Doc}
}

func genDeclDocs(
	decl *ast.GenDecl,
) (groups []*ast.CommentGroup) {
	groups = memberDocs(decl)
	if !slices.ContainsFunc(
		decl.Specs,
		isExportedSpec,
	) {
		return groups
	}

	groups = append(
		groups,
		decl.Doc,
	)
	for _, spec := range decl.Specs {
		groups = append(
			groups,
			exportedSpecDoc(spec),
		)
	}

	return groups
}

func memberDocs(
	decl *ast.GenDecl,
) (groups []*ast.CommentGroup) {
	for _, structType := range source.Nodes[*ast.StructType](decl) {
		groups = append(
			groups,
			fieldDocs(structType.Fields)...,
		)
	}

	for _, interfaceType := range source.Nodes[*ast.InterfaceType](decl) {
		groups = append(
			groups,
			fieldDocs(interfaceType.Methods)...,
		)
	}

	return groups
}

func fieldDocs(
	fields *ast.FieldList,
) (groups []*ast.CommentGroup) {
	for _, field := range slices.DeleteFunc(
		slices.Clone(fields.List),
		isUnexportedField,
	) {
		groups = append(
			groups,
			field.Doc,
		)
	}

	return groups
}

func isExportedSpec(
	spec ast.Spec,
) (exported bool) {
	return exportedName(spec) != ""
}

func exportedSpecDoc(
	spec ast.Spec,
) (doc *ast.CommentGroup) {
	if !isExportedSpec(spec) {
		return nil
	}

	return specDoc(spec)
}

func specDoc(
	spec ast.Spec,
) (doc *ast.CommentGroup) {
	switch typed := spec.(type) {
	case *ast.TypeSpec:
		return typed.Doc
	case *ast.ValueSpec:
		return typed.Doc
	}

	return nil
}

func exportedName(
	spec ast.Spec,
) (name string) {
	switch typed := spec.(type) {
	case *ast.TypeSpec:
		return firstExported([]*ast.Ident{typed.Name})
	case *ast.ValueSpec:
		return firstExported(typed.Names)
	}

	return ""
}

func firstExported(
	names []*ast.Ident,
) (name string) {
	index := slices.IndexFunc(
		names,
		(*ast.Ident).IsExported,
	)
	if index < 0 {
		return ""
	}

	return names[index].Name
}

func isUnexportedField(
	field *ast.Field,
) (unexported bool) {
	if len(field.Names) > 0 {
		return firstExported(field.Names) == ""
	}

	return !token.IsExported(embeddedName(field.Type))
}

func embeddedName(
	expr ast.Expr,
) (name string) {
	switch typed := expr.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.SelectorExpr:
		return typed.Sel.Name
	case *ast.StarExpr:
		return embeddedName(typed.X)
	case *ast.IndexExpr:
		return embeddedName(typed.X)
	case *ast.IndexListExpr:
		return embeddedName(typed.X)
	}

	return ""
}

func hasDoc(
	group *ast.CommentGroup,
) (documented bool) {
	return group != nil && slices.ContainsFunc(
		group.List,
		isProse,
	)
}

func isFileHeader(
	syntax *ast.File,
	group *ast.CommentGroup,
) (isHeader bool) {
	isBeforePackage := group.End() < syntax.Package
	isHeaderContent := isGeneratedMarker(group.List[0]) || !slices.ContainsFunc(
		group.List,
		isNotBuildConstraint,
	)
	return isBeforePackage && isHeaderContent
}

func isGeneratedMarker(
	comment *ast.Comment,
) (isMarker bool) {
	return strings.HasPrefix(
		comment.Text,
		generatedPrefix,
	) && strings.HasSuffix(
		comment.Text,
		generatedSuffix,
	)
}

func isNotBuildConstraint(
	comment *ast.Comment,
) (notConstraint bool) {
	return !isDirective(comment) && !strings.HasPrefix(
		comment.Text,
		legacyBuildPrefix,
	)
}

func isRestricted(
	comment *ast.Comment,
) (restricted bool) {
	body := strings.TrimSpace(strings.TrimSuffix(
		strings.TrimPrefix(
			strings.TrimPrefix(
				comment.Text,
				lineComment,
			),
			blockOpening,
		),
		blockClosing,
	))
	isLink := strings.Contains(
		comment.Text,
		plainLink,
	) || strings.Contains(
		comment.Text,
		secureLink,
	)
	return !isDirective(comment) && !strings.HasPrefix(
		body,
		deprecatedPrefix,
	) && !isLink
}

func isProse(
	comment *ast.Comment,
) (prose bool) {
	return !isDirective(comment)
}

func isDirective(
	comment *ast.Comment,
) (directive bool) {
	return strings.HasPrefix(
		comment.Text,
		directivePrefix,
	) || strings.HasPrefix(
		comment.Text,
		lineDirective,
	)
}

func isNolint(
	comment *ast.Comment,
) (suppression bool) {
	return strings.HasPrefix(
		comment.Text,
		nolintPrefix,
	)
}
