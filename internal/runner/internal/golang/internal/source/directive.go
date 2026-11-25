package source

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const (
	generateDirective  = "//go:generate "
	goCommand          = "go"
	runSubcommand      = "run"
	toolSubcommand     = "tool"
	flagPrefix         = "-"
	typeFlag           = "-type"
	typeFlagAssignment = "-type="
	typeSeparator      = ","
	stringerSuffix     = "_string.go"
	toolArgument       = 2
)

func isStringerOutput(
	path string,
) (proven bool) {
	directory := filepath.Dir(path)
	declared := declaredTypes(
		path,
		directory,
	)
	for _, command := range directives(
		path,
		[]string{directory},
	) {
		proven = proven || commandTool(command) == stringerTool && slices.ContainsFunc(
			stringerTypes(command),
			namesOutput(
				path,
				declared,
			),
		)
	}

	return proven
}

func invokesAny(
	generated string,
	directories []string,
	tools []string,
) (invokes bool) {
	for _, command := range directives(
		generated,
		directories,
	) {
		invokes = invokes || slices.Contains(
			tools,
			commandTool(command),
		)
	}

	return invokes
}

func directives(
	generated string,
	directories []string,
) (commands [][]string) {
	for _, path := range siblingSources(
		generated,
		directories,
	) {
		commands = append(
			commands,
			fileDirectives(path)...,
		)
	}

	return commands
}

func siblingSources(
	generated string,
	directories []string,
) (paths []string) {
	for _, directory := range directories {
		paths = append(
			paths,
			directorySources(directory)...,
		)
	}

	return slices.DeleteFunc(
		paths,
		isSelf(generated),
	)
}

func directorySources(
	directory string,
) (paths []string) {
	paths, err := Entries(directory)
	if err == nil {
		return paths
	}

	return nil
}

func isSelf(
	generated string,
) (excluded func(string) bool) {
	return func(
		path string,
	) (isGenerated bool) {
		return path == generated
	}
}

func fileDirectives(
	path string,
) (commands [][]string) {
	content, err := os.ReadFile(path)
	if err == nil {
		return generateCommands(string(content))
	}

	return nil
}

func generateCommands(
	content string,
) (commands [][]string) {
	for _, line := range strings.Split(
		content,
		"\n",
	) {
		commands = append(
			commands,
			lineCommand(line)...,
		)
	}

	return commands
}

func lineCommand(
	line string,
) (commands [][]string) {
	command, isDirective := strings.CutPrefix(
		line,
		generateDirective,
	)
	fields := strings.Fields(command)
	isCommand := isDirective && len(fields) > 0
	if !isCommand {
		return nil
	}

	return [][]string{fields}
}

func commandTool(
	command []string,
) (tool string) {
	isGoRun := len(command) > toolArgument && command[0] == goCommand && slices.Contains(
		[]string{runSubcommand, toolSubcommand},
		command[1],
	)
	if !isGoRun {
		return toolName(command[0])
	}

	index := slices.IndexFunc(
		command[toolArgument:],
		isNotFlag,
	)
	if index < 0 {
		return ""
	}

	return toolName(command[toolArgument+index])
}

func isNotFlag(
	field string,
) (notFlag bool) {
	return !strings.HasPrefix(
		field,
		flagPrefix,
	)
}

func stringerTypes(
	command []string,
) (names []string) {
	for index := range command {
		names = append(
			names,
			flagTypes(
				command,
				index,
			)...,
		)
	}

	return names
}

func flagTypes(
	command []string,
	index int,
) (names []string) {
	value, isAssigned := strings.CutPrefix(
		command[index],
		typeFlagAssignment,
	)
	isSeparate := command[index] == typeFlag && index+1 < len(command)
	if isSeparate {
		return strings.Split(
			command[index+1],
			typeSeparator,
		)
	}

	if !isAssigned {
		return nil
	}

	return strings.Split(
		value,
		typeSeparator,
	)
}

func namesOutput(
	path string,
	declared []string,
) (matches func(string) bool) {
	return func(
		name string,
	) (named bool) {
		return slices.Contains(
			declared,
			name,
		) && filepath.Base(path) == strings.ToLower(name)+stringerSuffix
	}
}

func declaredTypes(
	generated string,
	directory string,
) (names []string) {
	for _, path := range siblingSources(
		generated,
		[]string{directory},
	) {
		names = append(
			names,
			fileTypes(path)...,
		)
	}

	return names
}

func fileTypes(
	path string,
) (names []string) {
	syntax, err := parser.ParseFile(
		token.NewFileSet(),
		path,
		nil,
		parser.SkipObjectResolution,
	)
	if err == nil {
		return typeNames(syntax)
	}

	return nil
}

func typeNames(
	syntax *ast.File,
) (names []string) {
	for _, spec := range Nodes[*ast.TypeSpec](syntax) {
		names = append(
			names,
			spec.Name.Name,
		)
	}

	return names
}
