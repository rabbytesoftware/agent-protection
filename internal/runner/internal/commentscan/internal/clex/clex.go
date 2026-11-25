package clex

import (
	"path/filepath"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
)

const (
	slashOpener     = "//"
	hashOpener      = "#"
	bothQuotes      = "\"'"
	doubleQuoteOnly = "\""
)

func Proto() (language lexeme.Language) {
	return lexeme.Language{
		Matches: hasExtension([]string{".proto"}),
		Spans: dialect{
			lineOpeners: []string{slashOpener},
			quotes:      bothQuotes,
		}.spans,
	}
}

func Css() (language lexeme.Language) {
	return lexeme.Language{
		Matches: hasExtension([]string{".css"}),
		Spans: dialect{
			quotes: bothQuotes,
		}.spans,
	}
}

func Scss() (language lexeme.Language) {
	return lexeme.Language{
		Matches: hasExtension([]string{
			".scss",
			".sass",
			".less",
			".styl",
		}),
		Spans: dialect{
			lineOpeners: []string{slashOpener},
			quotes:      bothQuotes,
		}.spans,
	}
}

func Jsonc() (language lexeme.Language) {
	return lexeme.Language{
		Matches: hasExtension([]string{".json", ".jsonc", ".json5"}),
		Spans: dialect{
			lineOpeners: []string{slashOpener},
			quotes:      bothQuotes,
		}.spans,
	}
}

func Hcl() (language lexeme.Language) {
	return lexeme.Language{
		Matches: hasExtension([]string{
			".tf",
			".tfvars",
			".hcl",
			".nomad",
		}),
		Spans: dialect{
			lineOpeners: []string{hashOpener, slashOpener},
			quotes:      doubleQuoteOnly,
		}.spans,
	}
}

func hasExtension(
	extensions []string,
) (matches func(string) bool) {
	return func(
		name string,
	) (matched bool) {
		return slices.Contains(
			extensions,
			strings.ToLower(filepath.Ext(name)),
		)
	}
}
