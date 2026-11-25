package onetypeperfile

import (
	"fmt"
	"go/ast"
	"path/filepath"
	"slices"
	"unicode"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule             = "R08"
	goExtension      = ".go"
	wordSeparator    = "_"
	localFormat      = "type %s is declared inside a function body: declare it at package level in its own file %s"
	anonymousSnippet = "anonymous struct type outside a test file: declare it as a named type in its own file named after it"
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
	return slices.Concat(
		topLevelFindings(file),
		localFindings(file),
		anonymousFindings(file),
	)
}

func localFindings(
	file source.File,
) (findings []report.Finding) {
	topLevel := topLevelTypes(file.Syntax)
	for _, spec := range source.Nodes[*ast.TypeSpec](file.Syntax) {
		findings = appendLocal(
			findings,
			file,
			spec,
			topLevel,
		)
	}

	return findings
}

func appendLocal(
	findings []report.Finding,
	file source.File,
	spec *ast.TypeSpec,
	topLevel []*ast.TypeSpec,
) (extended []report.Finding) {
	if slices.Contains(
		topLevel,
		spec,
	) {
		return findings
	}

	return append(
		findings,
		report.Finding{
			Rule:     rule,
			Location: file.Location(spec),
			Snippet: fmt.Sprintf(
				localFormat,
				spec.Name.Name,
				snakeCase(spec.Name.Name)+goExtension,
			),
		},
	)
}

func anonymousFindings(
	file source.File,
) (findings []report.Finding) {
	specs := source.Nodes[*ast.TypeSpec](file.Syntax)
	for _, structType := range source.Nodes[*ast.StructType](file.Syntax) {
		findings = appendAnonymous(
			findings,
			file,
			structType,
			specs,
		)
	}

	return findings
}

func appendAnonymous(
	findings []report.Finding,
	file source.File,
	structType *ast.StructType,
	specs []*ast.TypeSpec,
) (extended []report.Finding) {
	isAnonymous := structType.Fields.NumFields() > 0 && !isDeclared(
		structType,
		specs,
	)
	if !isAnonymous {
		return findings
	}

	return append(
		findings,
		report.Finding{
			Rule:     rule,
			Location: file.Location(structType),
			Snippet:  anonymousSnippet,
		},
	)
}

func isDeclared(
	node ast.Node,
	specs []*ast.TypeSpec,
) (declared bool) {
	for _, spec := range specs {
		declared = declared || spec.Pos() <= node.Pos() && node.End() <= spec.End()
	}

	return declared
}

func topLevelFindings(
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
