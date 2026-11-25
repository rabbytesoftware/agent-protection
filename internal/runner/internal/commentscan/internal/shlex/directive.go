package shlex

import (
	"regexp"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
)

const (
	shellcheckPattern      = `^#\s*shellcheck(?:\s+(?:disable|enable|source|source-path|shell|external-sources)=\S+)+\s*$`
	schemaPattern          = `^#\s*yaml-language-server:\s*\$schema=\S+\s*$`
	parserDirectivePattern = `^#\s*(?i:syntax|escape|check)\s*=\s*\S+\s*$`
)

func isToolDirective(
	text []byte,
	span lexeme.Span,
) (directive bool) {
	comment := text[span.Start:span.End]
	return lexeme.IsLineLeading(
		text,
		span,
	) && (regexp.MustCompile(shellcheckPattern).Match(comment) ||
		regexp.MustCompile(schemaPattern).Match(comment))
}

func isRecipeShebang(
	text []byte,
	span lexeme.Span,
) (directive bool) {
	return lexeme.IsLineLeading(
		text,
		span,
	) && lexeme.HasPrefix(
		text,
		span.Start,
		lexeme.Shebang,
	)
}

func parserDirectivesEnd(
	text []byte,
) (end int) {
	directive := regexp.MustCompile(parserDirectivePattern)
	for end < len(text) &&
		directive.Match(text[end:lexeme.LineEnd(
			text,
			end,
		)]) {
		end = lexeme.NextLine(
			text,
			end,
		)
	}

	return end
}

func startingFrom(
	spans []lexeme.Span,
	at int,
) (kept []lexeme.Span) {
	for _, span := range spans {
		kept = append(
			kept,
			spanFrom(
				span,
				at,
			)...,
		)
	}

	return kept
}

func spanFrom(
	span lexeme.Span,
	at int,
) (kept []lexeme.Span) {
	if span.Start < at {
		return nil
	}

	return []lexeme.Span{span}
}
