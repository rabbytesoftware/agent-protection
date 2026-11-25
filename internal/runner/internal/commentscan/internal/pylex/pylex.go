package pylex

import (
	"path/filepath"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
)

func Python() (language lexeme.Language) {
	return lexeme.Language{
		Matches:      isPython,
		Spans:        pythonCode,
		Interpreters: []string{"python", "pypy"},
	}
}

func Ruby() (language lexeme.Language) {
	return lexeme.Language{
		Matches:      isRuby,
		Spans:        rubyCode,
		Interpreters: []string{"ruby"},
	}
}

func Graphql() (language lexeme.Language) {
	return lexeme.Language{
		Matches: isGraphql,
		Spans:   hashSpans,
	}
}

func Toml() (language lexeme.Language) {
	return lexeme.Language{
		Matches: isToml,
		Spans:   tomlCode,
	}
}

func pythonCode(
	text []byte,
) (spans []lexeme.Span) {
	return lexeme.Without(
		text,
		hashSpans(text),
		isEncodingDeclaration,
	)
}

func rubyCode(
	text []byte,
) (spans []lexeme.Span) {
	return lexeme.Without(
		text,
		hashSpans(text),
		isMagicComment,
	)
}

func tomlCode(
	text []byte,
) (spans []lexeme.Span) {
	return lexeme.Without(
		text,
		hashSpans(text),
		isSchemaDirective,
	)
}

func isPython(
	name string,
) (matched bool) {
	return slices.Contains(
		[]string{".py", ".pyi"},
		filepath.Ext(name),
	)
}

func isRuby(
	name string,
) (matched bool) {
	return slices.Contains(
		[]string{".rb", ".rake", ".gemspec"},
		filepath.Ext(name),
	) || slices.Contains(
		[]string{"Rakefile", "Gemfile"},
		name,
	)
}

func isGraphql(
	name string,
) (matched bool) {
	return slices.Contains(
		[]string{".graphql", ".graphqls", ".gql"},
		filepath.Ext(name),
	)
}

func isToml(
	name string,
) (matched bool) {
	return filepath.Ext(name) == ".toml"
}
