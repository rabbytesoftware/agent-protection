package commentpolicy

import (
	"slices"
	"strconv"
	"strings"
	"unicode"
)

const (
	goDirectivePrefix     = "//go:"
	lineDirectivePrefix   = "//line "
	exportDirectivePrefix = "//export "
	externDirectivePrefix = "//extern "
	legacyBuildPrefix     = "// +build "
	positionSeparator     = ":"
)

func isDirective(
	text string,
) (directive bool) {
	return isGoDirective(text) ||
		isLineDirective(text) ||
		isSymbolDirective(
			text,
			exportDirectivePrefix,
		) ||
		isSymbolDirective(
			text,
			externDirectivePrefix,
		) ||
		strings.HasPrefix(
			text,
			legacyBuildPrefix,
		)
}

func isGoDirective(
	text string,
) (directive bool) {
	rest, hasPrefix := strings.CutPrefix(
		text,
		goDirectivePrefix,
	)
	fields := strings.Fields(rest)
	return hasPrefix && len(fields) > 0 && slices.Contains(
		goDirectiveNames(),
		fields[0],
	)
}

func isLineDirective(
	text string,
) (directive bool) {
	rest, hasPrefix := strings.CutPrefix(
		text,
		lineDirectivePrefix,
	)
	fields := strings.Fields(rest)
	isSingleField := hasPrefix && len(fields) == 1
	if !isSingleField {
		return false
	}

	parts := strings.Split(
		fields[0],
		positionSeparator,
	)
	return len(parts) >= 2 && len(parts) <= 3 && slices.IndexFunc(
		parts[1:],
		isNotNumber,
	) < 0 && parts[0] != ""
}

func isSymbolDirective(
	text string,
	prefix string,
) (directive bool) {
	rest, hasPrefix := strings.CutPrefix(
		text,
		prefix,
	)
	fields := strings.Fields(rest)
	return hasPrefix && len(fields) == 1 && strings.IndexFunc(
		fields[0],
		isNotIdentifierRune,
	) < 0
}

func isNotNumber(
	part string,
) (notNumber bool) {
	_, err := strconv.Atoi(part)
	return err != nil
}

func isNotIdentifierRune(
	character rune,
) (notIdentifier bool) {
	return character != '_' && !unicode.IsLetter(character) && !unicode.IsDigit(character)
}

func goDirectiveNames() (names []string) {
	return []string{"build", "embed", "generate", "linkname", "noinline", "nosplit", "noescape", "norace", "nocheckptr", "uintptrescapes", "uintptrkeepalive", "systemstack", "nowritebarrier", "wasmimport", "wasmexport", "debug", "cgo_import_dynamic", "cgo_export_static", "cgo_export_dynamic", "cgo_ldflag"}
}
