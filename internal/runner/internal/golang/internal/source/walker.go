package source

import (
	"fmt"
	"io/fs"
	"os"
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

	if w.isSkipped(
		path,
		entry,
	) {
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

func (w *walker) isSkipped(
	path string,
	entry fs.DirEntry,
) (skipped bool) {
	isVendored := filepath.Dir(path) == w.root && slices.Contains(
		skippedDirectories(),
		entry.Name(),
	)
	isNestedModule := path != w.root && exists(filepath.Join(
		path,
		goModFile,
	))
	return entry.IsDir() && (isVendored || isNestedModule)
}

func skippedDirectories() (names []string) {
	return []string{"vendor", ".git", "node_modules"}
}

func HasSource(
	dir string,
) (hasSource bool, err error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false,
			fmt.Errorf(
				"read %s: %w",
				dir,
				err,
			)
	}

	return slices.ContainsFunc(
			entries,
			isSourceEntry,
		),
		nil
}

func Entries(
	dir string,
) (paths []string, err error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil,
			fmt.Errorf(
				"read %s: %w",
				dir,
				err,
			)
	}

	for _, entry := range slices.DeleteFunc(
		entries,
		isNotSourceEntry,
	) {
		paths = append(
			paths,
			filepath.Join(
				dir,
				entry.Name(),
			),
		)
	}

	return paths,
		nil
}

func isSourceEntry(
	entry os.DirEntry,
) (isSource bool) {
	return !entry.IsDir() && filepath.Ext(entry.Name()) == goExtension && !IsTest(entry.Name())
}

func isNotSourceEntry(
	entry os.DirEntry,
) (excluded bool) {
	return !isSourceEntry(entry)
}
