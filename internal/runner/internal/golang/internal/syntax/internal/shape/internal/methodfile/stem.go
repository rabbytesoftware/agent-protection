package methodfile

import (
	"path/filepath"
	"slices"
	"strings"
	"unicode"
)

const (
	goExtension   = ".go"
	wordSeparator = "_"
)

func stem(
	path string,
) (name string) {
	return strings.TrimSuffix(
		filepath.Base(path),
		goExtension,
	)
}

func portable(
	name string,
) (trimmed string) {
	words := strings.Split(
		name,
		wordSeparator,
	)
	for len(words) > 1 && slices.Contains(
		platforms(),
		words[len(words)-1],
	) {
		words = words[:len(words)-1]
	}

	return strings.Join(
		words,
		wordSeparator,
	)
}

func isNumbered(
	name string,
	siblings []string,
) (numbered bool) {
	words := strings.Split(
		name,
		wordSeparator,
	)
	lastWord := words[len(words)-1]
	isSegment := len(words) > 1 &&
		isDigits(lastWord)
	if isSegment {
		return true
	}

	base := strings.TrimRightFunc(
		name,
		unicode.IsDigit,
	)
	isSuffixed := base != name && base != ""
	return isSuffixed && (slices.Contains(
		siblings,
		base,
	) || slices.Contains(
		fillerWords(),
		base,
	))
}

func isDigits(
	word string,
) (digits bool) {
	return word != "" && strings.TrimFunc(
		word,
		unicode.IsDigit,
	) == ""
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

func fillerWords() (words []string) {
	return []string{
		"part",
		"file",
		"chunk",
		"piece",
		"section",
		"segment",
		"split",
		"extra",
		"more",
		"rest",
		"misc",
		"tmp",
		"temp",
		"copy",
		"new",
		"old",
	}
}

func platforms() (names []string) {
	return []string{
		"aix",
		"android",
		"darwin",
		"dragonfly",
		"freebsd",
		"illumos",
		"ios",
		"js",
		"linux",
		"netbsd",
		"openbsd",
		"plan9",
		"solaris",
		"wasip1",
		"windows",
		"unix",
		"386",
		"amd64",
		"arm",
		"arm64",
		"loong64",
		"mips",
		"mipsle",
		"mips64",
		"mips64le",
		"ppc64",
		"ppc64le",
		"riscv64",
		"s390x",
		"wasm",
	}
}
