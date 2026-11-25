package workspace

import (
	"fmt"
	"io/fs"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

const (
	gitBinary     = "git"
	hiddenPrefix  = "."
	ignoredPrefix = "_"
)

type discovery struct {
	root    string
	modules []string
}

func nestedModules(
	root string,
) (modules []string, err error) {
	found := &discovery{root: filepath.Clean(root)}
	err = filepath.WalkDir(
		found.root,
		found.visit,
	)
	if err != nil {
		return nil,
			fmt.Errorf(
				"find nested modules in %s: %w",
				root,
				err,
			)
	}

	return slices.DeleteFunc(
			found.modules,
			isGitIgnored(found.root),
		),
		nil
}

func (d *discovery) visit(
	path string,
	entry fs.DirEntry,
	err error,
) (walkErr error) {
	if err != nil {
		return err
	}

	isSkipped := entry.IsDir() && path != d.root && isSkippedDirectory(entry.Name())
	if isSkipped {
		return filepath.SkipDir
	}

	isNestedModule := !entry.IsDir() && entry.Name() == goModFile && filepath.Dir(path) != d.root
	if !isNestedModule {
		return nil
	}

	rel, err := filepath.Rel(
		d.root,
		filepath.Dir(path),
	)
	d.modules = append(
		d.modules,
		filepath.ToSlash(rel),
	)
	return err
}

func isSkippedDirectory(
	name string,
) (skipped bool) {
	return slices.Contains(
		skippedDirectories(),
		name,
	) || strings.HasPrefix(
		name,
		hiddenPrefix,
	) || strings.HasPrefix(
		name,
		ignoredPrefix,
	)
}

func isGitIgnored(
	root string,
) (isIgnored func(string) bool) {
	return func(
		module string,
	) (ignored bool) {
		command := exec.Command(
			gitBinary,
			"-C",
			root,
			"check-ignore",
			"-q",
			"--",
			module+"/",
		)
		return command.Run() == nil
	}
}

func skippedDirectories() (names []string) {
	return []string{"vendor", "testdata", "node_modules"}
}
