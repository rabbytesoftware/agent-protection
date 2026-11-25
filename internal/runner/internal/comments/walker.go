package comments

import (
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
)

const hiddenPrefix = "."

type walker struct {
	root       string
	extensions []string
	files      []string
}

func (w *walker) visit(
	path string,
	entry fs.DirEntry,
	err error,
) (walkErr error) {
	if err != nil {
		return err
	}

	if entry.IsDir() {
		return directoryVerdict(
			w.root,
			path,
			entry.Name(),
		)
	}

	if slices.Contains(
		w.extensions,
		filepath.Ext(path),
	) {
		w.files = append(
			w.files,
			path,
		)
	}

	return nil
}

func directoryVerdict(
	root string,
	path string,
	name string,
) (verdict error) {
	if path == root {
		return nil
	}

	isSkipped := strings.HasPrefix(
		name,
		hiddenPrefix,
	) || slices.Contains(
		skippedDirectories(),
		name,
	)
	if isSkipped {
		return filepath.SkipDir
	}

	return nil
}

func skippedDirectories() (names []string) {
	return []string{"node_modules", "target", "dist", "build", "out", "coverage", "vendor"}
}
