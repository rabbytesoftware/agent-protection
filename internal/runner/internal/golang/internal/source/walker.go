package source

import (
	"io/fs"
	"path/filepath"
	"slices"
)

type walker struct {
	root        string
	files       []string
	directories []string
}

func (w *walker) visit(
	path string,
	entry fs.DirEntry,
	err error,
) (walkErr error) {
	if err != nil {
		return err
	}

	isSkipped := entry.IsDir() && filepath.Dir(path) == w.root && slices.Contains(
		skippedDirectories(),
		entry.Name(),
	)
	if isSkipped {
		return filepath.SkipDir
	}

	if entry.IsDir() {
		w.directories = append(
			w.directories,
			path,
		)
		return nil
	}

	if filepath.Ext(path) == goExtension {
		w.files = append(
			w.files,
			path,
		)
	}

	return nil
}

func skippedDirectories() (names []string) {
	return []string{"vendor", ".git", "node_modules"}
}
