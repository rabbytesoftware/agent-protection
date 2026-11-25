package source

import (
	"cmp"
	"go/ast"
	"go/scanner"
	"go/token"
	"slices"
)

type Edit struct {
	Start int
	End   int
	Text  string
}

func Splice(
	content []byte,
	edits []Edit,
) (spliced []byte) {
	cursor := 0
	for _, edit := range disjoint(edits) {
		spliced = slices.Concat(
			spliced,
			content[cursor:edit.Start],
			[]byte(edit.Text),
		)
		cursor = edit.End
	}

	return slices.Concat(
		spliced,
		content[cursor:],
	)
}

func disjoint(
	edits []Edit,
) (kept []Edit) {
	for _, edit := range slices.SortedStableFunc(
		slices.Values(edits),
		byPosition,
	) {
		kept = appendDisjoint(
			kept,
			edit,
		)
	}

	return kept
}

func appendDisjoint(
	kept []Edit,
	edit Edit,
) (extended []Edit) {
	isOverlapping := len(kept) > 0 && edit.Start < kept[len(kept)-1].End
	if isOverlapping {
		return kept
	}

	return append(
		kept,
		edit,
	)
}

func byPosition(
	first Edit,
	second Edit,
) (order int) {
	return cmp.Or(
		cmp.Compare(
			first.Start,
			second.Start,
		),
		cmp.Compare(
			first.End,
			second.End,
		),
	)
}

const (
	lineBreak         = "\n"
	commaAndLineBreak = ",\n"
)

func Breaks(
	file File,
	opening token.Pos,
	items []ast.Node,
	closing token.Pos,
) (edits []Edit) {
	if len(items) == 0 {
		return nil
	}

	previous := opening
	for _, item := range items {
		edits = append(
			edits,
			breakBefore(
				file,
				previous,
				item.Pos(),
			)...,
		)
		previous = item.End()
	}

	return append(
		edits,
		closingBreak(
			file,
			previous,
			closing,
		)...,
	)
}

func breakBefore(
	file File,
	previous token.Pos,
	next token.Pos,
) (edits []Edit) {
	if startsBelow(
		file,
		previous,
		next,
	) {
		return nil
	}

	return []Edit{insertion(
		file,
		next,
		lineBreak,
	)}
}

func closingBreak(
	file File,
	last token.Pos,
	closing token.Pos,
) (edits []Edit) {
	if startsBelow(
		file,
		last,
		closing,
	) {
		return nil
	}

	if hasComma(file.Content[file.Offset(last):file.Offset(closing)]) {
		return []Edit{insertion(
			file,
			closing,
			lineBreak,
		)}
	}

	return []Edit{insertion(
		file,
		closing,
		commaAndLineBreak,
	)}
}

func startsBelow(
	file File,
	previous token.Pos,
	next token.Pos,
) (below bool) {
	return file.FileSet.Position(next).Line > file.FileSet.Position(previous).Line
}

func insertion(
	file File,
	position token.Pos,
	text string,
) (edit Edit) {
	offset := file.Offset(position)
	return Edit{Start: offset, End: offset, Text: text}
}

func hasComma(
	between []byte,
) (found bool) {
	var tokens scanner.Scanner
	tokens.Init(
		token.NewFileSet().AddFile(
			"",
			-1,
			len(between),
		),
		between,
		nil,
		0,
	)
	for kind := scan(&tokens); kind != token.EOF; kind = scan(&tokens) {
		found = found || kind == token.COMMA
	}

	return found
}

func scan(
	tokens *scanner.Scanner,
) (kind token.Token) {
	_, kind, _ = tokens.Scan()
	return kind
}
