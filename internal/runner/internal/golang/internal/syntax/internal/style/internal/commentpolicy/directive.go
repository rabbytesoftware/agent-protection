package commentpolicy

import (
	"go/ast"
	"go/build/constraint"
	"path"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	goDirectivePrefix     = "//go:"
	lineDirectivePrefix   = "//line "
	exportDirectivePrefix = "//export "
	externDirectivePrefix = "//extern "
	plusBuildPrefix       = "// +build "
	buildDirectivePrefix  = "//go:build "
	positionSeparator     = ":"
	settingSeparator      = "="
	aliasFlag             = "-command"
	buildName             = "build"
	generateName          = "generate"
	linknameName          = "linkname"
	wasmimportName        = "wasmimport"
	wasmexportName        = "wasmexport"
	debugName             = "debug"
	aliasCommandIndex     = 2
	linknameMaxArguments  = 2
	wasmimportArguments   = 2
)

func isDirective(
	file source.File,
	comment *ast.Comment,
) (directive bool) {
	return isGoDirective(
		file,
		comment,
	) || isLineDirective(comment.Text) ||
		isSymbolDirective(
			comment.Text,
			exportDirectivePrefix,
		) ||
		isSymbolDirective(
			comment.Text,
			externDirectivePrefix,
		) ||
		isPlusBuild(
			file,
			comment,
		)
}

func isGoDirective(
	file source.File,
	comment *ast.Comment,
) (directive bool) {
	rest, hasPrefix := strings.CutPrefix(
		comment.Text,
		goDirectivePrefix,
	)
	fields := strings.Fields(rest)
	if !hasPrefix ||
		len(fields) == 0 {
		return false
	}

	return hasValidArguments(
		file,
		comment,
		fields[0],
		fields[1:],
	)
}

func hasValidArguments(
	file source.File,
	comment *ast.Comment,
	name string,
	arguments []string,
) (valid bool) {
	switch {
	case slices.Contains(
		bareDirectiveNames(),
		name,
	):
		return len(arguments) == 0
	case name == generateName:
		return isGenerate(
			file,
			comment,
			arguments,
		)
	case name == buildName:
		return isConstraint(
			file,
			comment,
		)
	case name == linknameName:
		return len(arguments) > 0 &&
			len(arguments) <= linknameMaxArguments
	}

	return hasValidSettings(
		name,
		arguments,
	)
}

func hasValidSettings(
	name string,
	arguments []string,
) (valid bool) {
	switch name {
	case wasmimportName:
		return len(arguments) == wasmimportArguments
	case wasmexportName:
		return len(arguments) == 1
	case debugName:
		return len(arguments) == 1 &&
			strings.Contains(
				arguments[0],
				settingSeparator,
			)
	}

	return len(arguments) > 0 && slices.Contains(
		argumentDirectiveNames(),
		name,
	)
}

func isGenerate(
	file source.File,
	comment *ast.Comment,
	arguments []string,
) (generate bool) {
	command := generateCommand(arguments)
	return command != "" &&
		!slices.Contains(
			proseCommands(),
			path.Base(command),
		) &&
		isTopLevel(
			file,
			comment,
		)
}

func generateCommand(
	arguments []string,
) (command string) {
	if len(arguments) == 0 {
		return ""
	}

	if arguments[0] != aliasFlag {
		return arguments[0]
	}

	if len(arguments) <= aliasCommandIndex {
		return ""
	}

	return arguments[aliasCommandIndex]
}

func isTopLevel(
	file source.File,
	comment *ast.Comment,
) (topLevel bool) {
	topLevel = true
	for _, declaration := range file.Syntax.Decls {
		topLevel = topLevel && !encloses(
			declaration,
			comment,
		)
	}

	return topLevel
}

func encloses(
	declaration ast.Decl,
	comment *ast.Comment,
) (enclosed bool) {
	return declaration.Pos() <= comment.Pos() &&
		comment.End() <= declaration.End()
}

func isConstraint(
	file source.File,
	comment *ast.Comment,
) (valid bool) {
	_, err := constraint.Parse(comment.Text)
	return err == nil &&
		comment.End() < file.Syntax.Package
}

func isPlusBuild(
	file source.File,
	comment *ast.Comment,
) (legacy bool) {
	return strings.HasPrefix(
		comment.Text,
		plusBuildPrefix,
	) && isConstraint(
		file,
		comment,
	) && hasBuildDirective(file)
}

func hasBuildDirective(
	file source.File,
) (found bool) {
	for _, group := range file.Syntax.Comments {
		found = found || slices.ContainsFunc(
			group.List,
			isBuildLine,
		)
	}

	return found
}

func isBuildLine(
	comment *ast.Comment,
) (build bool) {
	return strings.HasPrefix(
		comment.Text,
		buildDirectivePrefix,
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

func bareDirectiveNames() (names []string) {
	return []string{
		"noinline",
		"nosplit",
		"noescape",
		"norace",
		"nocheckptr",
		"uintptrescapes",
		"uintptrkeepalive",
		"systemstack",
		"nowritebarrier",
		"nowritebarrierrec",
		"yeswritebarrierrec",
		"notinheap",
	}
}

func argumentDirectiveNames() (names []string) {
	return []string{
		"embed",
		"cgo_import_dynamic",
		"cgo_import_static",
		"cgo_export_static",
		"cgo_export_dynamic",
		"cgo_ldflag",
		"cgo_unsafe_args",
		"cgo_dynamic_linker",
	}
}

func proseCommands() (commands []string) {
	return []string{
		"echo",
		"true",
		"false",
		":",
		"printf",
	}
}
