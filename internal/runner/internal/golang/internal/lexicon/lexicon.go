package lexicon

import (
	"strings"
	"unicode"
)

const (
	minimumRoot     = 3
	pluralSuffix    = "s"
	doubledPlural   = "ss"
	pluralY         = "ies"
	singularY       = "y"
	silentE         = "e"
	gerundSuffix    = "ing"
	pastSuffix      = "ed"
	agentSuffix     = "er"
	agentSuffixLong = "or"
	vowels          = "aeiou"
)

// Words splits text into lowercase words. Runs of letters and digits are
// cut again at camelCase humps, at the end of an acronym and at every
// letter-digit edge, so HTTPServer yields http and server, and handler2
// yields handler and 2. Everything else separates words.
func Words(
	text string,
) (words []string) {
	for _, token := range strings.FieldsFunc(
		text,
		isSeparator,
	) {
		words = append(
			words,
			tokenWords([]rune(token))...,
		)
	}

	return words
}

// Stem reduces an English word to a crude root so that its inflections
// compare equal: reads, reader and reading all become read, created and
// creates both become creat. A suffix is only stripped while three letters
// remain, and the root is meant for comparing words, never for showing them.
func Stem(
	word string,
) (root string) {
	root = uninflected(
		singular(word),
		[]string{gerundSuffix, pastSuffix, agentSuffix, agentSuffixLong},
	)
	if len(root) <= minimumRoot {
		return root
	}

	return strings.TrimSuffix(
		root,
		silentE,
	)
}

func isSeparator(
	character rune,
) (separator bool) {
	return !unicode.IsLetter(character) && !unicode.IsDigit(character)
}

func tokenWords(
	runes []rune,
) (words []string) {
	cut := firstBoundary(runes)
	word := strings.ToLower(string(runes[:cut]))
	if cut == len(runes) {
		return []string{word}
	}

	return append(
		[]string{word},
		tokenWords(runes[cut:])...,
	)
}

func firstBoundary(
	runes []rune,
) (cut int) {
	cut = 1
	for cut < len(runes) && !isBoundary(
		runes,
		cut,
	) {
		cut++
	}

	return cut
}

func isBoundary(
	runes []rune,
	index int,
) (boundary bool) {
	previous := runes[index-1]
	current := runes[index]
	isHump := unicode.IsLower(previous) && unicode.IsUpper(current)
	isNextLower := index+1 < len(runes) && unicode.IsLower(runes[index+1])
	isAcronymEnd := unicode.IsUpper(previous) && unicode.IsUpper(current) && isNextLower
	isDigitEdge := unicode.IsDigit(previous) != unicode.IsDigit(current)
	return isHump || isAcronymEnd || isDigitEdge
}

func singular(
	word string,
) (noun string) {
	isLong := len(word) > minimumRoot
	isPluralY := isLong && strings.HasSuffix(
		word,
		pluralY,
	)
	if isPluralY {
		return strings.TrimSuffix(
			word,
			pluralY,
		) + singularY
	}

	isPlural := isLong && strings.HasSuffix(
		word,
		pluralSuffix,
	) && !strings.HasSuffix(
		word,
		doubledPlural,
	)
	if !isPlural {
		return word
	}

	return strings.TrimSuffix(
		word,
		pluralSuffix,
	)
}

func uninflected(
	word string,
	suffixes []string,
) (root string) {
	if len(suffixes) == 0 {
		return word
	}

	trimmed := strings.TrimSuffix(
		word,
		suffixes[0],
	)
	isInflected := trimmed != word && len(trimmed) >= minimumRoot
	if isInflected {
		return undoubled(trimmed)
	}

	return uninflected(
		word,
		suffixes[1:],
	)
}

func undoubled(
	root string,
) (single string) {
	last := len(root) - 1
	isDoubled := len(root) > minimumRoot && root[last] == root[last-1] && !strings.ContainsRune(
		vowels,
		rune(root[last]),
	)
	if !isDoubled {
		return root
	}

	return root[:last]
}
