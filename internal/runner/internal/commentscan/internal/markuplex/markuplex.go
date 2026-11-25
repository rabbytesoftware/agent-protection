package markuplex

import (
	"path/filepath"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
)

func Markup(
	script lexeme.Language,
	style lexeme.Language,
) (language lexeme.Language) {
	return lexeme.Language{
		Matches: isMarkup,
		Spans: document{
			script: script,
			style:  style,
		}.spans,
	}
}

func isMarkup(
	name string,
) (matched bool) {
	return slices.Contains(
		markupExtensions(),
		strings.ToLower(filepath.Ext(name)),
	)
}

func markupExtensions() (extensions []string) {
	return []string{
		".html",
		".htm",
		".xhtml",
		".tmpl",
		".gohtml",
		".gotmpl",
		".tpl",
		".vue",
		".svelte",
		".xml",
		".svg",
		".xsd",
		".xsl",
		".xslt",
		".plist",
	}
}
