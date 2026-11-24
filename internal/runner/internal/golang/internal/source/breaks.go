package source

import (
	"go/scanner"
	"go/token"
)

const (
	lineBreak         = "\n"
	commaAndLineBreak = ",\n"
)

func Breaks(
	file File,
	opening token.Pos,
	items []Extent,
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
				item.Start,
			)...,
		)
		previous = item.End
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
