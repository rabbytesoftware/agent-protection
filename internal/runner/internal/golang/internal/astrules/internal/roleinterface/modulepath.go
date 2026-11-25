package roleinterface

import (
	"maps"
	"slices"
	"strings"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func derivedModulePath(
	files []source.File,
) (modulePath string) {
	directories := packageDirectories(files)
	votes := map[string]int{}
	for _, candidate := range candidates(
		files,
		directories,
	) {
		votes[candidate]++
	}

	for _, candidate := range slices.Sorted(maps.Keys(votes)) {
		modulePath = moreVoted(
			modulePath,
			candidate,
			votes,
		)
	}

	return modulePath
}

func moreVoted(
	current string,
	candidate string,
	votes map[string]int,
) (winner string) {
	if votes[candidate] > votes[current] {
		return candidate
	}

	return current
}

func candidates(
	files []source.File,
	directories []string,
) (found []string) {
	for _, file := range files {
		found = append(
			found,
			fileCandidates(
				file,
				directories,
			)...,
		)
	}

	return found
}

func fileCandidates(
	file source.File,
	directories []string,
) (found []string) {
	for _, spec := range file.Syntax.Imports {
		found = append(
			found,
			modulePathCandidate(
				importedPath(spec),
				directories,
			)...,
		)
	}

	return found
}

func modulePathCandidate(
	importPath string,
	directories []string,
) (found []string) {
	longest := ""
	for _, directory := range directories {
		longest = longerSuffix(
			longest,
			directory,
			importPath,
		)
	}

	if longest == "" {
		return nil
	}

	return []string{strings.TrimSuffix(
		importPath,
		pathSeparator+longest,
	)}
}

func longerSuffix(
	longest string,
	directory string,
	importPath string,
) (longer string) {
	isLonger := len(directory) > len(longest) && strings.HasSuffix(
		importPath,
		pathSeparator+directory,
	)
	if isLonger {
		return directory
	}

	return longest
}

func packageDirectories(
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
		isCurrentDirectory,
	)
}

func isCurrentDirectory(
	directory string,
) (isCurrent bool) {
	return directory == currentDirectory
}
