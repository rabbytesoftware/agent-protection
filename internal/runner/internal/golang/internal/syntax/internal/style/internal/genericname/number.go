package genericname

import (
	"regexp"
	"slices"
	"strings"
	"unicode"
)

const (
	versionMarker    = "V"
	percentileMarker = "P"
	benchmarkPrefix  = "Benchmark"
	statusPattern    = `^[1-5][0-9][0-9]$`
)

func isNumbered(
	name string,
) (numbered bool) {
	head := unnumbered(name)
	isUnnumbered := head == name || head == ""
	if isUnnumbered {
		return false
	}

	isSized := strings.HasPrefix(
		name,
		benchmarkPrefix,
	)
	isStatus := regexp.MustCompile(statusPattern).MatchString(name[len(head):])
	return !isSized && !isStatus && !isMarkedNumber(head) && !isDomainTerm(name)
}

func isMarkedNumber(
	head string,
) (marked bool) {
	return slices.ContainsFunc(
		[]string{versionMarker, percentileMarker},
		markerOf(head),
	)
}

func markerOf(
	head string,
) (matches func(string) bool) {
	return func(
		marker string,
	) (isMarker bool) {
		return strings.HasSuffix(
			head,
			marker,
		) || strings.EqualFold(
			head,
			marker,
		)
	}
}

func isDomainTerm(
	name string,
) (term bool) {
	for _, known := range numberedTerms() {
		term = term || endsWithWord(
			name,
			known,
		)
	}

	return term
}

func endsWithWord(
	name string,
	word string,
) (ends bool) {
	start := len(name) - len(word)
	isSuffix := start >= 0 && strings.EqualFold(
		name[start:],
		word,
	)
	if !isSuffix {
		return false
	}

	return start == 0 || unicode.IsUpper(rune(name[start])) || !unicode.IsLetter(rune(name[start-1]))
}

func numberedTerms() (terms []string) {
	return slices.Concat(
		encodingTerms(),
		hashTerms(),
		cipherTerms(),
		protocolTerms(),
		numericTerms(),
		platformTerms(),
		mediaTerms(),
	)
}

func encodingTerms() (terms []string) {
	return []string{
		"base32",
		"base58",
		"base64",
		"b64",
		"utf8",
		"utf16",
		"utf32",
		"x509",
		"pkcs1",
		"pkcs7",
		"pkcs8",
		"pkcs12",
		"rfc3339",
		"rfc1123",
		"rfc822",
		"iso8601",
		"e164",
	}
}

func hashTerms() (terms []string) {
	return []string{
		"sha1",
		"sha224",
		"sha256",
		"sha384",
		"sha512",
		"md4",
		"md5",
		"crc32",
		"crc64",
		"adler32",
		"fnv32",
		"fnv64",
		"blake2",
		"argon2",
	}
}

func cipherTerms() (terms []string) {
	return []string{
		"ed25519",
		"x25519",
		"p256",
		"p384",
		"p521",
		"aes128",
		"aes256",
		"chacha20",
		"poly1305",
	}
}

func protocolTerms() (terms []string) {
	return []string{
		"http1",
		"http2",
		"http3",
		"h2",
		"h2c",
		"ipv4",
		"ipv6",
		"ip4",
		"ip6",
		"tls12",
		"tls13",
		"oauth1",
		"oauth2",
	}
}

func numericTerms() (terms []string) {
	return []string{
		"int8",
		"int16",
		"int32",
		"int64",
		"uint8",
		"uint16",
		"uint32",
		"uint64",
		"float32",
		"float64",
		"complex64",
		"complex128",
		"vec2",
		"vec3",
		"vec4",
		"mat3",
		"mat4",
	}
}

func platformTerms() (terms []string) {
	return []string{
		"s3",
		"ec2",
		"amd64",
		"arm64",
		"x86",
		"win32",
		"win64",
	}
}

func mediaTerms() (terms []string) {
	return []string{
		"bzip2",
		"h264",
		"h265",
		"mp3",
		"mp4",
	}
}
