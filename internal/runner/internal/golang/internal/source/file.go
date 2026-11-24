package source

import (
	"fmt"
	"go/ast"
	"go/token"
)

// File is one parsed Go file: its path as walked, its path relative to the
// checked root (the form every finding reports), its syntax tree with the
// file set that resolves the tree's positions, and the bytes it was parsed
// from, against which edits are computed.
type File struct {
	Path    string
	Rel     string
	FileSet *token.FileSet
	Syntax  *ast.File
	Content []byte
}

// Location renders the position of node as "relative/path.go:line", the
// location format every finding uses.
func (f File) Location(
	node ast.Node,
) (location string) {
	return fmt.Sprintf(
		"%s:%d",
		f.Rel,
		f.Line(node),
	)
}

// Line returns the 1-based line node starts on.
func (f File) Line(
	node ast.Node,
) (line int) {
	return f.FileSet.Position(node.Pos()).Line
}

// Column returns the 1-based byte column node starts at. In gofmt-formatted
// code every indentation level is one tab, so the column of a statement
// minus the column of its enclosing declaration is its nesting depth.
func (f File) Column(
	node ast.Node,
) (column int) {
	return f.FileSet.Position(node.Pos()).Column
}

// Offset returns the byte offset of position within the file's content.
func (f File) Offset(
	position token.Pos,
) (offset int) {
	return f.FileSet.Position(position).Offset
}

// LineOffset returns the byte offset at which the 1-based line starts.
func (f File) LineOffset(
	line int,
) (offset int) {
	tokenFile := f.FileSet.File(f.Syntax.Pos())
	return tokenFile.Offset(tokenFile.LineStart(line))
}
