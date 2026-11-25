package packageset

import (
	"regexp"
	"slices"
	"strings"
)

const (
	coveredShare    = 4
	wholeShare      = 5
	shortestStem    = 3
	versionPattern  = `v?[0-9]+$`
	versionDivision = ""
)

func isLookalike(
	name string,
	other string,
) (lookalike bool) {
	shorter := min(
		len(name),
		len(other),
	)
	isCoveredPrefix := sharedPrefix(
		name,
		other,
	)*wholeShare >= shorter*coveredShare
	stem := conceptStem(name)
	return isCoveredPrefix || len(stem) >= shortestStem && stem == conceptStem(other)
}

func sharedPrefix(
	name string,
	other string,
) (shared int) {
	limit := min(
		len(name),
		len(other),
	)
	for shared < limit && name[shared] == other[shared] {
		shared++
	}

	return shared
}

func conceptStem(
	name string,
) (stem string) {
	return uninflected(unversioned(name))
}

func unversioned(
	name string,
) (bare string) {
	bare = regexp.MustCompile(versionPattern).ReplaceAllString(
		name,
		versionDivision,
	)
	for _, marker := range versionMarkers() {
		bare = trimmedAffix(
			bare,
			marker,
		)
	}

	return bare
}

func trimmedAffix(
	name string,
	marker string,
) (bare string) {
	bare = strings.TrimSuffix(
		strings.TrimPrefix(
			name,
			marker,
		),
		marker,
	)
	if len(bare) < shortestStem {
		return name
	}

	return bare
}

func uninflected(
	name string,
) (stem string) {
	index := slices.IndexFunc(
		conceptSuffixes(),
		trimsTo(name),
	)
	if index < 0 {
		return name
	}

	return strings.TrimSuffix(
		name,
		conceptSuffixes()[index],
	)
}

func trimsTo(
	name string,
) (trims func(string) bool) {
	return func(
		suffix string,
	) (trimmed bool) {
		return strings.HasSuffix(
			name,
			suffix,
		) && len(name)-len(suffix) >= shortestStem
	}
}

func versionMarkers() (markers []string) {
	return []string{
		"old",
		"new",
		"legacy",
		"next",
		"prev",
	}
}

func conceptSuffixes() (suffixes []string) {
	return []string{
		"ions",
		"ment",
		"age",
		"ing",
		"ion",
		"es",
		"er",
		"or",
		"s",
		"e",
	}
}
