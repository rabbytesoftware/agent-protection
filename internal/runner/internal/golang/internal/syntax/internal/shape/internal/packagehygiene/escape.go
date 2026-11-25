package packagehygiene

import (
	"slices"
	"strings"
)

const wordGap = " "

func isEscapeHatch(
	name string,
) (escape bool) {
	for _, segment := range strings.FieldsFunc(
		strings.ToLower(name),
		isSeparator,
	) {
		escape = escape || slices.Contains(
			escapeNames(),
			segment,
		) || hasEscapeAffix(segment)
	}

	return escape
}

func hasEscapeAffix(
	segment string,
) (affixed bool) {
	for _, part := range strings.Fields(withoutEmbeddings(segment)) {
		affixed = affixed || slices.ContainsFunc(
			escapeNames(),
			isAffixOf(part),
		)
	}

	return affixed
}

func withoutEmbeddings(
	segment string,
) (stripped string) {
	stripped = segment
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

func escapeNames() (names []string) {
	return slices.Concat(
		provisionalNames(),
		patchNames(),
		trialNames(),
	)
}

func provisionalNames() (names []string) {
	return []string{
		"legacy",
		"compat",
		"tmp",
		"temp",
		"temporary",
		"wip",
		"old",
		"new",
		"deprecated",
		"unused",
		"todo",
		"junk",
		"interim",
		"stopgap",
	}
}

func patchNames() (names []string) {
	return []string{
		"hack",
		"kludge",
		"workaround",
		"shim",
		"stub",
		"patch",
		"fixme",
		"hotfix",
		"quickfix",
		"bandaid",
		"bypass",
		"shortcut",
		"adhoc",
	}
}

func trialNames() (names []string) {
	return []string{
		"scratch",
		"draft",
		"sandbox",
		"playground",
		"experimental",
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
		"hold",
		"fold",
		"bold",
		"cold",
		"gold",
		"mold",
		"sold",
		"told",
		"news",
		"newline",
		"newton",
		"renew",
		"anew",
		"knew",
		"sinew",
		"junket",
		"shimmer",
		"stubborn",
		"stubble",
		"dispatch",
		"jsonpatch",
		"mergepatch",
		"overdraft",
	}
}
