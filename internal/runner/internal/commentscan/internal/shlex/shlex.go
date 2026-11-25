package shlex

import (
	"path/filepath"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
)

const (
	makefileExtension   = ".mk"
	dockerfileName      = "Dockerfile"
	dockerfileSeparator = "."
)

func Shell() (language lexeme.Language) {
	return lexeme.Language{
		Matches: isShell,
		Spans:   shellCode,
	}
}

func Yaml() (language lexeme.Language) {
	return lexeme.Language{
		Matches: isYaml,
		Spans:   yamlCode,
	}
}

func Makefile() (language lexeme.Language) {
	return lexeme.Language{
		Matches: isMakefile,
		Spans:   shellSpans,
	}
}

func Dockerfile() (language lexeme.Language) {
	return lexeme.Language{
		Matches: isDockerfile,
		Spans:   dockerfileCode,
	}
}

func shellCode(
	text []byte,
) (spans []lexeme.Span) {
	return lexeme.Without(
		text,
		shellSpans(text),
		isToolDirective,
	)
}

func yamlCode(
	text []byte,
) (spans []lexeme.Span) {
	return lexeme.Without(
		text,
		yamlSpans(text),
		isToolDirective,
	)
}

func dockerfileCode(
	text []byte,
) (spans []lexeme.Span) {
	return startingFrom(
		shellSpans(text),
		parserDirectivesEnd(text),
	)
}

func isShell(
	name string,
) (matched bool) {
	return slices.Contains(
		[]string{
			".sh",
			".bash",
			".zsh",
			".ksh",
		},
		filepath.Ext(name),
	)
}

func isYaml(
	name string,
) (matched bool) {
	return slices.Contains(
		[]string{".yaml", ".yml"},
		filepath.Ext(name),
	)
}

func isMakefile(
	name string,
) (matched bool) {
	return slices.Contains(
		[]string{"Makefile", "makefile", "GNUmakefile"},
		name,
	) || filepath.Ext(name) == makefileExtension
}

func isDockerfile(
	name string,
) (matched bool) {
	return slices.Contains(
		[]string{dockerfileName, "Containerfile"},
		name,
	) || strings.HasPrefix(
		name,
		dockerfileName+dockerfileSeparator,
	) || strings.EqualFold(
		filepath.Ext(name),
		dockerfileSeparator+dockerfileName,
	)
}
