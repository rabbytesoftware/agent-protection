package source

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
)

type File struct {
	Path        string
	Rel         string
	FileSet     *token.FileSet
	Syntax      *ast.File
	Content     []byte
	IsGenerated bool
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

const sentinelPrefix = "Err"

func Nodes[T ast.Node](
	root ast.Node,
) (nodes []T) {
	for node := range ast.Preorder(root) {
		nodes = appendMatching(
			nodes,
			node,
		)
	}

	return nodes
}

func DeclarationPackages(
	files []File,
) (directories []string) {
	behaving := []string{}
	for _, file := range files {
		directory := filepath.ToSlash(filepath.Dir(file.Rel))
		directories = append(
			directories,
			directory,
		)
		behaving = append(
			behaving,
			behavingDirectory(
				file,
				directory,
			)...,
		)
	}

	slices.Sort(directories)
	return withoutAny(
		slices.Compact(directories),
		behaving,
	)
}

func ErrorMethods() (names []string) {
	return []string{
		"Error",
		"Unwrap",
		"Is",
		"As",
		"String",
	}
}

func appendMatching[T ast.Node](
	nodes []T,
	node ast.Node,
) (matching []T) {
	typed, matches := node.(T)
	if !matches {
		return nodes
	}

	return append(
		nodes,
		typed,
	)
}

func behavingDirectory(
	file File,
	directory string,
) (behaving []string) {
	if !slices.ContainsFunc(
		file.Syntax.Decls,
		isBehavior,
	) {
		return nil
	}

	return []string{directory}
}

func isBehavior(
	declaration ast.Decl,
) (behaves bool) {
	switch typed := declaration.(type) {
	case *ast.FuncDecl:
		return typed.Recv == nil || !slices.Contains(
			ErrorMethods(),
			typed.Name.Name,
		)
	case *ast.GenDecl:
		return typed.Tok == token.VAR && slices.ContainsFunc(
			valueNames(typed),
			isStateName,
		)
	}

	return false
}

func isStateName(
	name *ast.Ident,
) (isState bool) {
	isSentinel := name.Name == "_" || name.IsExported() && strings.HasPrefix(
		name.Name,
		sentinelPrefix,
	)
	return !isSentinel
}

func valueNames(
	general *ast.GenDecl,
) (names []*ast.Ident) {
	for _, spec := range Nodes[*ast.ValueSpec](general) {
		names = append(
			names,
			spec.Names...,
		)
	}

	return names
}

func withoutAny(
	directories []string,
	excluded []string,
) (kept []string) {
	for _, directory := range directories {
		kept = append(
			kept,
			absentFrom(
				excluded,
				directory,
			)...,
		)
	}

	return kept
}

func absentFrom(
	excluded []string,
	directory string,
) (kept []string) {
	if slices.Contains(
		excluded,
		directory,
	) {
		return nil
	}

	return []string{directory}
}
