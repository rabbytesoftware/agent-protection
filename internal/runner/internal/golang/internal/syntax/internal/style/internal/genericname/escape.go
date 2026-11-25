package genericname

import (
	"fmt"
	"go/ast"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	escapeFormat = "carries the escape-hatch word %q, which says how provisional it is instead of what it does"
	wordGap      = " "
)

func escapeReason(
	name string,
) (reason string) {
	index := slices.IndexFunc(
		Words(name),
		isEscapeWord,
	)
	if index < 0 {
		return ""
	}

	return fmt.Sprintf(
		escapeFormat,
		Words(name)[index],
	)
}

func isEscapeWord(
	word string,
) (escape bool) {
	if slices.Contains(
		escapeWords(),
		word,
	) {
		return true
	}

	for _, part := range strings.Fields(withoutEmbeddings(word)) {
		escape = escape || slices.ContainsFunc(
			escapeWords(),
			isAffixOf(part),
		)
	}

	return escape
}

func withoutEmbeddings(
	word string,
) (stripped string) {
	stripped = word
	for _, embedding := range embeddingWords() {
		stripped = strings.ReplaceAll(
			stripped,
			embedding,
			wordGap,
		)
	}

	return stripped
}

func isAffixOf(
	part string,
) (matches func(string) bool) {
	return func(
		word string,
	) (isAffix bool) {
		return strings.HasPrefix(
			part,
			word,
		) || strings.HasSuffix(
			part,
			word,
		)
	}
}

func isTestSupport(
	rel string,
) (support bool) {
	return source.IsTest(rel) || source.IsMock(rel)
}

func packageValueNames(
	syntax *ast.File,
) (identifiers []*ast.Ident) {
	for _, declaration := range syntax.Decls {
		identifiers = append(
			identifiers,
			declarationValues(declaration)...,
		)
	}

	return identifiers
}

func declarationValues(
	declaration ast.Decl,
) (identifiers []*ast.Ident) {
	general, isGeneral := declaration.(*ast.GenDecl)
	if !isGeneral {
		return nil
	}

	for _, spec := range general.Specs {
		identifiers = append(
			identifiers,
			plainValues(spec)...,
		)
	}

	return identifiers
}

func plainValues(
	spec ast.Spec,
) (identifiers []*ast.Ident) {
	value, isValue := spec.(*ast.ValueSpec)
	if !isValue {
		return nil
	}

	for index, name := range value.Names {
		identifiers = append(
			identifiers,
			plainValue(
				value,
				index,
				name,
			)...,
		)
	}

	return identifiers
}

func plainValue(
	value *ast.ValueSpec,
	index int,
	name *ast.Ident,
) (identifiers []*ast.Ident) {
	isClosure := index < len(value.Values) &&
		isFuncLiteral(value.Values[index])
	if isClosure {
		return nil
	}

	return []*ast.Ident{name}
}

func isFuncLiteral(
	value ast.Expr,
) (closure bool) {
	_, closure = value.(*ast.FuncLit)
	return closure
}

func escapeWords() (words []string) {
	return []string{
		"legacy",
		"compat",
		"tmp",
		"temp",
		"temporary",
		"wip",
		"todo",
		"junk",
		"interim",
		"stopgap",
		"hack",
		"kludge",
		"workaround",
		"shim",
		"stub",
		"fixme",
		"hotfix",
		"quickfix",
		"bandaid",
		"bypass",
		"adhoc",
	}
}

func embeddingWords() (words []string) {
	return []string{
		"templ",
		"temper",
		"tempest",
		"temporal",
		"contemp",
		"tmpl",
		"tmpfs",
		"wipe",
		"compatib",
		"junket",
		"shimmer",
		"stubborn",
		"stubble",
	}
}
