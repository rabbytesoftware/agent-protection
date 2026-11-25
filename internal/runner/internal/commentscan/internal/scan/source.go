package scan

import (
	"bytes"
	"path"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
)

const (
	binaryWindow      = 8192
	nul               = 0
	envCommand        = "env"
	optionPrefix      = "-"
	assignment        = "="
	versionCharacters = "0123456789."
)

type source struct {
	path       string
	language   lexeme.Language
	candidates []lexeme.Language
	fallback   lexeme.Language
}

func (s source) resolve(
	text []byte,
) (language lexeme.Language, found bool) {
	if s.language.Spans != nil {
		return s.language,
			true
	}

	if isBinary(text) {
		return lexeme.Language{},
			false
	}

	index := slices.IndexFunc(
		s.candidates,
		interprets(interpreterOf(text)),
	)
	if index < 0 {
		return s.fallback,
			s.fallback.Spans != nil
	}

	return s.candidates[index],
		true
}

func isBinary(
	text []byte,
) (binary bool) {
	return bytes.IndexByte(
		text[:min(
			len(text),
			binaryWindow,
		)],
		nul,
	) >= 0
}

func interprets(
	interpreter string,
) (matches func(lexeme.Language) bool) {
	return func(
		language lexeme.Language,
	) (interpreted bool) {
		return slices.Contains(
			language.Interpreters,
			interpreter,
		)
	}
}

func interpreterOf(
	text []byte,
) (interpreter string) {
	end := lexeme.ShebangEnd(text)
	commands := slices.DeleteFunc(
		strings.Fields(string(text[min(
			len(lexeme.Shebang),
			end,
		):end])),
		isCommandArgument,
	)
	if len(commands) == 0 {
		return ""
	}

	command := path.Base(commands[0])
	isWrapped := command == envCommand && len(commands) > 1
	if isWrapped {
		command = path.Base(commands[1])
	}

	return strings.TrimRight(
		command,
		versionCharacters,
	)
}

func isCommandArgument(
	field string,
) (argument bool) {
	return strings.HasPrefix(
		field,
		optionPrefix,
	) || strings.Contains(
		field,
		assignment,
	)
}
