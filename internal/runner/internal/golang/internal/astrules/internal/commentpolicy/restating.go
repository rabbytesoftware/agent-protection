package commentpolicy

import (
	"fmt"
	"go/ast"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/lexicon"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	restatingSnippet = "doc comment of %s only restates its name: say what the name cannot (behavior, contract, edge cases)"
	wordSeparator    = " "
)

func restatingFindings(
	file source.File,
	decl ast.Decl,
) (findings []report.Finding) {
	switch typed := decl.(type) {
	case *ast.FuncDecl:
		return restatingFunc(
			file,
			typed,
		)
	case *ast.GenDecl:
		return restatingSpecs(
			file,
			typed,
		)
	}

	return nil
}

func restatingFunc(
	file source.File,
	decl *ast.FuncDecl,
) (findings []report.Finding) {
	isRestating := needsDoc(
		file,
		decl,
	) && hasDoc(decl.Doc) && restates(
		decl.Doc,
		decl.Name.Name,
		receiverName(decl),
	)
	if !isRestating {
		return nil
	}

	return restatingDoc(
		file,
		decl.Doc,
		decl.Name.Name,
	)
}

func restatingSpecs(
	file source.File,
	decl *ast.GenDecl,
) (findings []report.Finding) {
	for _, spec := range decl.Specs {
		findings = append(
			findings,
			restatingSpec(
				file,
				ownDoc(
					decl,
					spec,
				),
				exportedName(spec),
			)...,
		)
	}

	return findings
}

func restatingSpec(
	file source.File,
	doc *ast.CommentGroup,
	name string,
) (findings []report.Finding) {
	isRestating := name != "" && hasDoc(doc) && restates(
		doc,
		name,
	)
	if !isRestating {
		return nil
	}

	return restatingDoc(
		file,
		doc,
		name,
	)
}

func ownDoc(
	decl *ast.GenDecl,
	spec ast.Spec,
) (doc *ast.CommentGroup) {
	doc = specDoc(spec)
	isSoleSpec := doc == nil && len(decl.Specs) == 1
	if !isSoleSpec {
		return doc
	}

	return decl.Doc
}

func receiverName(
	decl *ast.FuncDecl,
) (name string) {
	if decl.Recv == nil {
		return ""
	}

	return embeddedName(decl.Recv.List[0].Type)
}

func restates(
	doc *ast.CommentGroup,
	names ...string,
) (restating bool) {
	own := stems(lexicon.Words(strings.Join(
		names,
		wordSeparator,
	)))
	said := stems(slices.DeleteFunc(
		lexicon.Words(doc.Text()),
		isFillerWord,
	))
	restating = true
	for _, root := range said {
		restating = restating && slices.Contains(
			own,
			root,
		)
	}

	return restating
}

func stems(
	words []string,
) (roots []string) {
	for _, word := range words {
		roots = append(
			roots,
			lexicon.Stem(word),
		)
	}

	return roots
}

func isFillerWord(
	word string,
) (filler bool) {
	return slices.Contains(
		fillerWords(),
		word,
	)
}

func restatingDoc(
	file source.File,
	doc *ast.CommentGroup,
	name string,
) (findings []report.Finding) {
	return []report.Finding{{
		Rule:     rule,
		Location: file.Location(doc),
		Snippet: fmt.Sprintf(
			restatingSnippet,
			name,
		),
	}}
}

func fillerWords() (words []string) {
	return []string{
		"a", "an", "the", "this", "that", "these", "those", "it", "its",
		"is", "are", "was", "be", "been", "being", "does", "do",
		"of", "for", "to", "in", "on", "at", "by", "with", "from", "into", "as",
		"and", "or", "whether", "if", "when",
		"returns", "return", "reports", "report", "new", "given",
		"value", "values", "instance", "object", "method", "function", "func",
		"type", "struct", "interface", "field", "constant", "variable",
		"implementation", "implements", "satisfies", "fulfils", "fulfills",
		"underlying", "backing", "current", "specified", "provided",
		"corresponding", "associated", "respective",
		"constructs", "creates", "builds", "makes", "initializes", "instantiates",
		"test", "double", "fake", "fakes", "stub", "mock", "code", "under",
	}
}
