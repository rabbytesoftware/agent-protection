package source

import (
	"fmt"
	"go/ast"
	"go/token"
)

type File struct {
	Path    string
	Rel     string
	FileSet *token.FileSet
	Syntax  *ast.File
	Content []byte
}

func (f File) Location(
	node ast.Node,
) (location string) {
	return fmt.Sprintf(
		"%s:%d",
		f.Rel,
		f.Line(node),
	)
}

func (f File) Line(
	node ast.Node,
) (line int) {
	return f.FileSet.Position(node.Pos()).Line
}

func (f File) Column(
	node ast.Node,
) (column int) {
	return f.FileSet.Position(node.Pos()).Column
}

func (f File) Offset(
	position token.Pos,
) (offset int) {
	return f.FileSet.Position(position).Offset
}

func (f File) LineOffset(
	line int,
) (offset int) {
	tokenFile := f.FileSet.File(f.Syntax.Pos())
	return tokenFile.Offset(tokenFile.LineStart(line))
}
