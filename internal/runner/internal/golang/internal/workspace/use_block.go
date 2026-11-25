package workspace

import "path/filepath"

type useBlock struct {
	members []string
	inBlock bool
}

func (b *useBlock) read(
	fields []string,
) {
	if len(fields) == 0 {
		return
	}

	if b.inBlock {
		b.inBlock = fields[0] != closeBlock
		b.add(fields[0])
		return
	}

	isUse := fields[0] == useDirective && len(fields) > 1
	if !isUse {
		return
	}

	b.inBlock = fields[1] == openBlock
	b.add(fields[1])
}

func (b *useBlock) add(
	field string,
) {
	isDelimiter := field == openBlock || field == closeBlock
	if isDelimiter {
		return
	}

	member := unquoted(field)
	if filepath.IsAbs(member) {
		return
	}

	b.members = append(
		b.members,
		filepath.ToSlash(filepath.Clean(member)),
	)
}
