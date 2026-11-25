package scan

import (
	"path/filepath"
	"slices"
	"strings"
)

func isCandidate(
	name string,
) (candidate bool) {
	return !slices.Contains(
		skippedExtensions(),
		strings.ToLower(filepath.Ext(name)),
	) && !slices.Contains(
		skippedNames(),
		name,
	) && !slices.ContainsFunc(
		noticePrefixes(),
		isPrefixOf(strings.ToUpper(name)),
	)
}

func isPrefixOf(
	name string,
) (matches func(string) bool) {
	return func(
		prefix string,
	) (isPrefix bool) {
		return strings.HasPrefix(
			name,
			prefix,
		)
	}
}

func skippedExtensions() (extensions []string) {
	return []string{
		".md",
		".markdown",
		".mdx",
		".txt",
		".rst",
		".adoc",
		".sum",
		".lock",
		".lockb",
	}
}

func skippedNames() (names []string) {
	return []string{
		".git",
		"package-lock.json",
		"npm-shrinkwrap.json",
		"pnpm-lock.yaml",
	}
}

func noticePrefixes() (prefixes []string) {
	return []string{
		"LICENSE",
		"LICENCE",
		"NOTICE",
		"COPYING",
		"AUTHORS",
		"CONTRIBUTORS",
		"PATENTS",
	}
}
