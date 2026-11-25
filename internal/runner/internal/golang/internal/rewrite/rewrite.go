package rewrite

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"os"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

const (
	maxPasses = 8
	fileMode  = 0o644
)

func Rewrite(
	root string,
	files []source.File,
	edits func(source.File) []source.Edit,
) (changed []string, err error) {
	failures := []error{}
	for _, file := range files {
		rewritten, fileErr := rewriteFile(
			root,
			file,
			edits,
		)
		changed = append(
			changed,
			rewritten...,
		)
		failures = append(
			failures,
			fileErr,
		)
	}

	return changed,
		errors.Join(failures...)
}

func rewriteFile(
	root string,
	file source.File,
	edits func(source.File) []source.Edit,
) (rewritten []string, err error) {
	content, err := settle(
		root,
		file.Path,
		file.Content,
		edits,
		maxPasses,
	)
	if err != nil {
		return nil,
			err
	}

	if bytes.Equal(
		content,
		file.Content,
	) {
		return nil,
			nil
	}

	err = os.WriteFile(
		file.Path,
		content,
		fileMode,
	)
	if err != nil {
		return nil,
			fmt.Errorf(
				"write %s: %w",
				file.Path,
				err,
			)
	}

	return []string{file.Rel},
		nil
}

func settle(
	root string,
	path string,
	content []byte,
	edits func(source.File) []source.Edit,
	passesLeft int,
) (settled []byte, err error) {
	file, err := formatted(
		root,
		path,
		content,
	)
	if err != nil {
		return nil,
			err
	}

	pending := edits(file)
	if len(pending) == 0 {
		return file.Content,
			nil
	}

	if passesLeft == 0 {
		return nil,
			unsettled(path)
	}

	return settle(
		root,
		path,
		source.Splice(
			file.Content,
			pending,
		),
		edits,
		passesLeft-1,
	)
}

func formatted(
	root string,
	path string,
	content []byte,
) (file source.File, err error) {
	canonical, err := format.Source(content)
	if err != nil {
		return source.File{},
			fmt.Errorf(
				"format %s: %w",
				path,
				err,
			)
	}

	return source.Parse(
		root,
		path,
		canonical,
	)
}

func unsettled(
	path string,
) (err error) {
	return fmt.Errorf(
		"%s: layout still changing after %d passes",
		path,
		maxPasses,
	)
}
