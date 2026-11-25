package genericname

import (
	"fmt"
	"go/ast"
	"slices"
	"strings"
	"unicode"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rule              = "R56"
	snippet           = "%s %s: name what it does or holds"
	prefixReason      = "opens with %q, which tells nothing about it"
	contentFreeReason = "is built only from content-free words"
	numberedReason    = "differs from a sibling name only by its number"
	programEntry      = "main"
	packageSetup      = "init"
	blank             = "_"
)

func Genericname(
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
	identifiers := slices.DeleteFunc(
		declaredNames(file.Syntax),
		isExempt,
	)
	siblings := []string{}
	for _, identifier := range identifiers {
		siblings = append(
			siblings,
			identifier.Name,
		)
	}

	for _, identifier := range identifiers {
		findings = append(
			findings,
			identifierFindings(
				file,
				siblings,
				identifier,
			)...,
		)
	}

	return findings
}

func identifierFindings(
	file source.File,
	siblings []string,
	identifier *ast.Ident,
) (findings []report.Finding) {
	reason := genericReason(
		identifier.Name,
		siblings,
	)
	if reason == "" {
		return nil
	}

	return []report.Finding{{
		Rule:     rule,
		Location: file.Location(identifier),
		Snippet: fmt.Sprintf(
			snippet,
			identifier.Name,
			reason,
		),
	}}
}

func genericReason(
	name string,
	siblings []string,
) (reason string) {
	words := Words(name)
	switch {
	case hasContentFreePrefix(words):
		return fmt.Sprintf(
			prefixReason,
			words[0],
		)
	case isContentFree(words):
		return contentFreeReason
	case isNumberedSibling(
		name,
		siblings,
	):
		return numberedReason
	}

	return ""
}

func hasContentFreePrefix(
	words []string,
) (prefixed bool) {
	return len(words) > 1 && slices.Contains(
		prefixWords(),
		words[0],
	)
}

func isContentFree(
	words []string,
) (contentFree bool) {
	isSpecific := slices.ContainsFunc(
		words,
		isMeaningful,
	)
	if isSpecific {
		return false
	}

	return len(words) > 1 || slices.ContainsFunc(
		words,
		isFiller,
	)
}

func isMeaningful(
	word string,
) (meaningful bool) {
	return !isNumber(word) && !isFiller(word) && !slices.Contains(
		roleWords(),
		word,
	)
}

func isFiller(
	word string,
) (filler bool) {
	return slices.Contains(
		fillerWords(),
		word,
	)
}

func isNumberedSibling(
	name string,
	siblings []string,
) (numbered bool) {
	stem := unnumbered(name)
	isNumbered := stem != name && stem != "" && unicode.IsLower(rune(stem[len(stem)-1]))
	if !isNumbered {
		return false
	}

	for _, sibling := range siblings {
		isSibling := sibling != name && unnumbered(sibling) == stem
		numbered = numbered || isSibling
	}

	return numbered
}

func unnumbered(
	name string,
) (stem string) {
	return strings.TrimRightFunc(
		name,
		unicode.IsDigit,
	)
}

func isNumber(
	word string,
) (number bool) {
	return unnumbered(word) == ""
}

func declaredNames(
	syntax *ast.File,
) (identifiers []*ast.Ident) {
	for _, function := range source.Nodes[*ast.FuncDecl](syntax) {
		identifiers = append(
			identifiers,
			function.Name,
		)
	}

	for _, spec := range source.Nodes[*ast.TypeSpec](syntax) {
		identifiers = append(
			identifiers,
			spec.Name,
		)
	}

	for _, assignment := range source.Nodes[*ast.AssignStmt](syntax) {
		identifiers = append(
			identifiers,
			closureTargets(
				assignment.Lhs,
				assignment.Rhs,
			)...,
		)
	}

	for _, spec := range source.Nodes[*ast.ValueSpec](syntax) {
		identifiers = append(
			identifiers,
			closureNames(spec)...,
		)
	}

	return identifiers
}

func closureNames(
	spec *ast.ValueSpec,
) (identifiers []*ast.Ident) {
	targets := []ast.Expr{}
	for _, name := range spec.Names {
		targets = append(
			targets,
			name,
		)
	}

	return closureTargets(
		targets,
		spec.Values,
	)
}

func closureTargets(
	targets []ast.Expr,
	values []ast.Expr,
) (identifiers []*ast.Ident) {
	for index, value := range values {
		identifiers = append(
			identifiers,
			closureTarget(
				targets[index],
				value,
			)...,
		)
	}

	return identifiers
}

func closureTarget(
	target ast.Expr,
	value ast.Expr,
) (identifiers []*ast.Ident) {
	_, isClosure := value.(*ast.FuncLit)
	identifier, isIdentifier := target.(*ast.Ident)
	isNamedClosure := isClosure && isIdentifier
	if !isNamedClosure {
		return nil
	}

	return []*ast.Ident{identifier}
}

func isExempt(
	identifier *ast.Ident,
) (exempt bool) {
	return identifier.IsExported() || slices.Contains(
		[]string{programEntry, packageSetup, blank},
		identifier.Name,
	)
}

func prefixWords() (words []string) {
	return []string{"impl", "helper", "helpers"}
}

func fillerWords() (words []string) {
	return []string{
		"fn", "func", "f", "cb", "helper", "helpers", "impl", "do",
		"tmp", "temp", "stuff", "thing", "things", "data", "util", "utils",
		"misc", "aux", "body", "work", "logic", "part", "wrap", "wrapper",
		"inner", "outer", "foo", "bar", "baz",
	}
}

func roleWords() (words []string) {
	return []string{
		"handler", "handle", "callback", "run", "make", "exec", "execute",
		"perform", "process", "call", "invoke",
	}
}
