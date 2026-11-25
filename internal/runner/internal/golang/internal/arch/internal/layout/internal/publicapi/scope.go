package publicapi

import (
	"fmt"
	"path/filepath"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	mainPackage   = "main"
	rootDirectory = "."
	internalChild = "internal/"
)

type scope struct {
	packageName  string
	label        string
	internal     string
	structs      []string
	vocabularies []string
	api          bool
	vocabulary   bool
}

func scopeOf(
	file source.File,
	sourceDirectories []string,
	supportDirectories []string,
	structs map[string][]string,
	vocabularies []string,
) (found scope) {
	directory := filepath.ToSlash(filepath.Dir(file.Rel))
	packageName := file.Syntax.Name.Name
	isModule := source.IsModule(
		filepath.Dir(file.Path),
		sourceDirectories,
	)
	isExempt := packageName == mainPackage || source.IsExample(file.Rel) || slices.Contains(
		supportDirectories,
		directory,
	)
	return scope{
		packageName: packageName,
		label: label(
			file,
			sourceDirectories,
		),
		internal:     internalPath(directory),
		structs:      structs[directory],
		vocabularies: vocabularies,
		api:          !isExempt && (isModule || !source.IsInternalTree(directory)),
		vocabulary: slices.Contains(
			vocabularyNames(),
			packageName,
		),
	}
}

func label(
	file source.File,
	sourceDirectories []string,
) (described string) {
	if source.IsModule(
		filepath.Dir(file.Path),
		sourceDirectories,
	) {
		return "module root " + file.Syntax.Name.Name
	}

	return "public package " + file.Syntax.Name.Name
}

func internalPath(
	directory string,
) (path string) {
	if directory == rootDirectory {
		return internalChild
	}

	return fmt.Sprintf(
		"%s/%s",
		directory,
		internalChild,
	)
}
