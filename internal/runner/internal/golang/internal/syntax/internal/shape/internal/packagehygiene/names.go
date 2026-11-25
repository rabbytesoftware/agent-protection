package packagehygiene

import (
	"regexp"
	"slices"
	"strings"
	"unicode"
)

const (
	wellFormedPattern = `^[a-z][a-z0-9]*$`
	versionPattern    = `^v[0-9]+$`
	pluralSuffix      = "s"
	shortestPlural    = 4
)

func isMalformed(
	name string,
) (malformed bool) {
	return !regexp.MustCompile(wellFormedPattern).MatchString(name)
}

func isPlural(
	name string,
) (plural bool) {
	isCandidate := !isMalformed(name) && len(name) >= shortestPlural && strings.HasSuffix(
		name,
		pluralSuffix,
	)
	return isCandidate && !slices.ContainsFunc(
		singularEndings(),
		suffixOf(name),
	) && !slices.Contains(
		singularNames(),
		name,
	)
}

func isGeneric(
	name string,
) (generic bool) {
	for _, segment := range strings.FieldsFunc(
		strings.ToLower(name),
		isSeparator,
	) {
		generic = generic || slices.Contains(
			genericNames(),
			segment,
		) || hasGenericAffix(segment)
	}

	return generic
}

func hasGenericAffix(
	segment string,
) (affixed bool) {
	for _, affix := range genericAffixes() {
		affixed = affixed || strings.HasPrefix(
			segment,
			affix,
		) || strings.HasSuffix(
			segment,
			affix,
		)
	}

	return affixed
}

func isNumbered(
	name string,
) (numbered bool) {
	last := []rune(name)[len([]rune(name))-1]
	return unicode.IsDigit(last) && !isVersion(name) && !slices.Contains(
		numberedTerms(),
		strings.ToLower(name),
	)
}

func isVersion(
	name string,
) (version bool) {
	return regexp.MustCompile(versionPattern).MatchString(name)
}

func suffixOf(
	name string,
) (matches func(string) bool) {
	return func(
		ending string,
	) (hasEnding bool) {
		return strings.HasSuffix(
			name,
			ending,
		)
	}
}

func isSeparator(
	character rune,
) (separates bool) {
	return !unicode.IsLetter(character) && !unicode.IsDigit(character)
}

func genericNames() (names []string) {
	return slices.Concat(
		bucketNames(),
		vagueNames(),
	)
}

func bucketNames() (names []string) {
	return []string{
		"util",
		"utils",
		"common",
		"helpers",
		"helper",
		"base",
		"misc",
		"shared",
		"core",
		"kit",
		"lib",
		"libs",
		"support",
		"toolbox",
		"impl",
		"logic",
		"pkg",
	}
}

func vagueNames() (names []string) {
	return []string{
		"stuff",
		"processing",
		"manager",
		"managers",
		"part",
		"parts",
		"extra",
		"extras",
		"more",
		"other",
		"others",
		"data",
		"generic",
		"generics",
		"general",
		"various",
		"glue",
	}
}

func genericAffixes() (affixes []string) {
	return []string{
		"util",
		"utils",
		"helper",
		"helpers",
	}
}

func singularEndings() (endings []string) {
	return []string{
		"ss",
		"us",
		"is",
		"ics",
	}
}

func singularNames() (names []string) {
	return []string{
		"mocks",
		"fakes",
		"examples",
		"news",
		"series",
		"species",
		"alias",
		"atlas",
		"canvas",
		"chaos",
		"lens",
		"postgres",
		"kubernetes",
		"windows",
		"macos",
		"https",
		"cors",
		"jwks",
		"models",
		"contracts",
		"types",
		"ports",
		"options",
		"entities",
	}
}

func numberedTerms() (terms []string) {
	return []string{
		"oauth1",
		"oauth2",
		"s3",
		"ec2",
		"sha1",
		"sha256",
		"sha512",
		"md5",
		"base32",
		"base58",
		"base64",
		"utf8",
		"utf16",
		"http2",
		"http3",
		"h2",
		"x509",
		"ipv4",
		"ipv6",
		"crc32",
		"crc64",
		"adler32",
		"ed25519",
		"x25519",
		"argon2",
		"bzip2",
		"pkcs7",
		"pkcs8",
		"pkcs12",
		"p256",
	}
}
