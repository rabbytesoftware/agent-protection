package scan

import (
	"io/fs"
	"os/exec"
	"path/filepath"
	"slices"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/lexeme"
)

type walker struct {
	root      string
	excluded  []string
	languages []lexeme.Language
	fallback  lexeme.Language
	files     []source
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
		w.excluded,
		filepath.Ext(path),
	) {
		return nil
	}

	w.files = append(
		w.files,
		w.sourcesOf(
			path,
			entry.Name(),
		)...,
	)
	return nil
}

func (w *walker) sourcesOf(
	path string,
	name string,
) (sources []source) {
	language, found := languageOf(
		w.languages,
		name,
	)
	if found {
		return []source{{
			path:     path,
			language: language,
		}}
	}

	if !isCandidate(name) {
		return nil
	}

	return []source{{
		path:       path,
		candidates: w.languages,
		fallback:   w.fallback,
	}}
}

func languageOf(
	languages []lexeme.Language,
	name string,
) (language lexeme.Language, found bool) {
	index := 0
	for index < len(languages) && !languages[index].Matches(name) {
		index++
	}

	if index == len(languages) {
		return lexeme.Language{},
			false
	}

	return languages[index],
		true
}

func directoryVerdict(
	root string,
	path string,
	name string,
) (verdict error) {
	isSkipped := path != root &&
		(slices.Contains(
			skippedDirectories(),
			name,
		) || isIgnored(
			root,
			path,
		))
	if isSkipped {
		return filepath.SkipDir
	}

	return nil
}

func skippedDirectories() (names []string) {
	return []string{
		"vendor",
		"node_modules",
		".git",
		"testdata",
	}
}

func isIgnored(
	root string,
	path string,
) (ignored bool) {
	return exec.Command(
		gitBinary,
		"-C",
		root,
		"check-ignore",
		"--quiet",
		path,
	).Run() == nil
}
