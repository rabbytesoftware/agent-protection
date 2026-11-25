package workspace

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	goWorkFile    = "go.work"
	useDirective  = "use"
	openBlock     = "("
	closeBlock    = ")"
	commentMarker = "//"
)

func workMembers(
	root string,
) (members []string, err error) {
	content, err := os.ReadFile(filepath.Join(
		root,
		goWorkFile,
	))
	if errors.Is(
		err,
		fs.ErrNotExist,
	) {
		return nil,
			nil
	}

	if err != nil {
		return nil,
			fmt.Errorf(
				"read %s: %w",
				goWorkFile,
				err,
			)
	}

	parsed := &useBlock{}
	for _, line := range strings.Split(
		string(content),
		"\n",
	) {
		parsed.read(strings.Fields(withoutComment(line)))
	}

	return parsed.members,
		nil
}

func unquoted(
	field string,
) (text string) {
	text, err := strconv.Unquote(field)
	if err == nil {
		return text
	}

	return field
}

func withoutComment(
	line string,
) (code string) {
	code, _, _ = strings.Cut(
		line,
		commentMarker,
	)
	return code
}
