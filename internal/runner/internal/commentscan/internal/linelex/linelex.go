package linelex

import (
	"path/filepath"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
)

const (
	hashMarker      = "#"
	semicolonMarker = ";"
	bangMarker      = "!"
)

func Ignore() (language lexeme.Language) {
	return lexeme.Language{
		Matches: isIgnoreFile,
		Spans: markers{
			openers: []string{hashMarker},
		}.spans,
	}
}

func Ini() (language lexeme.Language) {
	return lexeme.Language{
		Matches: isIniFile,
		Spans: markers{
			openers: []string{hashMarker, semicolonMarker},
		}.spans,
	}
}

func Properties() (language lexeme.Language) {
	return lexeme.Language{
		Matches: isPropertiesFile,
		Spans: markers{
			openers: []string{hashMarker, bangMarker},
		}.spans,
	}
}

func Fallback() (language lexeme.Language) {
	return lexeme.Language{
		Matches: isNothing,
		Spans:   fallbackSpans,
	}
}

func fallbackSpans(
	text []byte,
) (spans []lexeme.Span) {
	return lexeme.Without(
		text,
		markers{
			openers:  fallbackOpeners(),
			bareRule: dashRule,
		}.spans(text),
		isShebang,
	)
}

func isShebang(
	text []byte,
	span lexeme.Span,
) (shebang bool) {
	return span.Start == 0 && lexeme.HasPrefix(
		text,
		0,
		lexeme.Shebang,
	)
}

func isIgnoreFile(
	name string,
) (matched bool) {
	return slices.Contains(
		[]string{
			".gitignore",
			".dockerignore",
			".gitattributes",
			".gitmodules",
			".npmignore",
			".prettierignore",
			".eslintignore",
			".helmignore",
			".gcloudignore",
			".vercelignore",
			"CODEOWNERS",
		},
		name,
	)
}

func isIniFile(
	name string,
) (matched bool) {
	return slices.Contains(
		[]string{
			".ini",
			".cfg",
			".conf",
			".editorconfig",
		},
		strings.ToLower(filepath.Ext(name)),
	) || slices.Contains(
		[]string{
			".npmrc",
			".gitconfig",
			".pylintrc",
			".flake8",
		},
		name,
	)
}

func isPropertiesFile(
	name string,
) (matched bool) {
	return strings.EqualFold(
		filepath.Ext(name),
		".properties",
	)
}

func isNothing(
	_ string,
) (matched bool) {
	return false
}

func fallbackOpeners() (openers []string) {
	return []string{
		hashMarker,
		"//",
		"-- ",
		semicolonMarker,
		"/*",
		"<!--",
	}
}
