package onetypeperfile

import (
	"fmt"
	"go/ast"
	"path/filepath"
	"unicode"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule          = "R08"
	goExtension   = ".go"
	wordSeparator = "_"
)

func Onetypeperfile(
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
	specs := topLevelTypes(file.Syntax)
	if len(specs) == 0 {
		return nil
	}

	if len(specs) > 1 {
		return []report.Finding{{
			Rule:     rule,
			Location: file.Location(specs[1]),
			Snippet: fmt.Sprintf(
				"%d type declarations in one file: move each type into its own file named after it",
				len(specs),
			),
		}}
	}

	return misnamedFindings(
		file,
		specs[0],
	)
}

func misnamedFindings(
	file source.File,
	spec *ast.TypeSpec,
) (findings []report.Finding) {
	base := filepath.Base(file.Path)
	expected := snakeCase(spec.Name.Name) + goExtension
	isHome := base == expected || source.IsCompositionRoot(
		file.Path,
		file.Syntax.Name.Name,
	)
	if isHome {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: file.Location(spec),
		Snippet: fmt.Sprintf(
			"type %s lives in %s: rename the file to %s",
			spec.Name.Name,
			base,
			expected,
		),
	}}
}

func topLevelTypes(
	syntax *ast.File,
) (specs []*ast.TypeSpec) {
	for _, decl := range syntax.Decls {
		specs = append(
			specs,
			typeSpecs(decl)...,
		)
	}

	return specs
}

func typeSpecs(
	decl ast.Decl,
) (specs []*ast.TypeSpec) {
	general, isGeneral := decl.(*ast.GenDecl)
	if !isGeneral {
		return nil
	}

	return source.Nodes[*ast.TypeSpec](general)
}

func snakeCase(
	name string,
) (snake string) {
	letters := []rune(name)
	for index, letter := range letters {
		snake += separator(
			letters,
			index,
		) + string(unicode.ToLower(letter))
	}

	return snake
}

func separator(
	letters []rune,
	index int,
) (prefix string) {
	isWordStart := index > 0 && unicode.IsUpper(letters[index]) && (!unicode.IsUpper(letters[index-1]) || nextIsLower(
		letters,
		index,
	))
	if !isWordStart {
		return ""
	}

	return wordSeparator
}

func nextIsLower(
	letters []rune,
	index int,
) (lower bool) {
	return index+1 < len(letters) && unicode.IsLower(letters[index+1])
}
