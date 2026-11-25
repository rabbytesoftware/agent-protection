package interfacesize

import (
	"go/ast"
	"path"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	rootDirectory = "."
	importQuotes  = "\"`"
	blankImport   = "_"
	dotImport     = "."
)

func modulePath(
	files []source.File,
) (module string) {
	directories := nestedDirectories(files)
	tally := map[string]int{}
	for _, file := range files {
		countImports(
			tally,
			file,
			directories,
		)
	}

	best := 0
	for _, candidate := range sortedCandidates(tally) {
		module, best = leader(
			module,
			best,
			candidate,
			tally[candidate],
		)
	}

	return module
}

func leader(
	module string,
	best int,
	candidate string,
	votes int,
) (winner string, winnerVotes int) {
	if votes <= best {
		return module,
			best
	}

	return candidate,
		votes
}

func sortedCandidates(
	tally map[string]int,
) (candidates []string) {
	for candidate := range tally {
		candidates = append(
			candidates,
			candidate,
		)
	}

	slices.Sort(candidates)
	return candidates
}

func countImports(
	tally map[string]int,
	file source.File,
	directories []string,
) {
	for _, spec := range file.Syntax.Imports {
		countImport(
			tally,
			importedPath(spec),
			directories,
		)
	}
}

func countImport(
	tally map[string]int,
	importPath string,
	directories []string,
) {
	matched := ""
	for _, directory := range directories {
		matched = longerMatch(
			matched,
			directory,
			importPath,
		)
	}

	if matched == "" {
		return
	}

	tally[strings.TrimSuffix(
		importPath,
		pathSeparator+matched,
	)]++
}

func longerMatch(
	matched string,
	directory string,
	importPath string,
) (longer string) {
	isLonger := len(directory) > len(matched) && strings.HasSuffix(
		importPath,
		pathSeparator+directory,
	)
	if isLonger {
		return directory
	}

	return matched
}

func nestedDirectories(
	files []source.File,
) (directories []string) {
	for _, file := range files {
		directories = append(
			directories,
			packageDirectory(file),
		)
	}

	slices.Sort(directories)
	return slices.DeleteFunc(
		slices.Compact(directories),
		isRootDirectory,
	)
}

func isRootDirectory(
	directory string,
) (isRoot bool) {
	return directory == rootDirectory
}

func fileAliases(
	file source.File,
	module string,
	packageNames map[string]string,
) (aliases map[string]string) {
	aliases = map[string]string{}
	for _, spec := range file.Syntax.Imports {
		addAlias(
			aliases,
			spec,
			module,
			packageNames,
		)
	}

	return aliases
}

func addAlias(
	aliases map[string]string,
	spec *ast.ImportSpec,
	module string,
	packageNames map[string]string,
) {
	importPath := importedPath(spec)
	directory, isLocal := localDirectory(
		importPath,
		module,
	)
	name := aliasName(
		spec,
		importPath,
		packageNames[directory],
	)
	isNamed := isLocal && name != blankImport && name != dotImport
	if isNamed {
		aliases[name] = directory
	}
}

func localDirectory(
	importPath string,
	module string,
) (directory string, isLocal bool) {
	if module == "" {
		return "",
			false
	}

	if importPath == module {
		return rootDirectory,
			true
	}

	return strings.CutPrefix(
		importPath,
		module+pathSeparator,
	)
}

func aliasName(
	spec *ast.ImportSpec,
	importPath string,
	declared string,
) (name string) {
	if spec.Name != nil {
		return spec.Name.Name
	}

	if declared != "" {
		return declared
	}

	return path.Base(importPath)
}

func importedPath(
	spec *ast.ImportSpec,
) (importPath string) {
	return strings.Trim(
		spec.Path.Value,
		importQuotes,
	)
}
