package source_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestFileLocation(
	t *testing.T,
) {
	file := parsedFile(
		t,
		"package p\n\nfunc f() {\n\treturn\n}\n",
	)
	statement := source.Nodes[*ast.ReturnStmt](file.Syntax)[0]
	location := file.Location(statement)
	if location != "dir/p.go:4" {
		t.Errorf(
			"Location() = %q, want dir/p.go:4",
			location,
		)
	}
}

func TestFileColumn(
	t *testing.T,
) {
	file := parsedFile(
		t,
		"package p\n\nfunc f() {\n\tif true {\n\t\treturn\n\t}\n}\n",
	)
	statement := source.Nodes[*ast.ReturnStmt](file.Syntax)[0]
	column := file.Column(statement)
	if column != 3 {
		t.Errorf(
			"Column() = %d, want 3 (two tabs deep)",
			column,
		)
	}
}

func parsedFile(
	t *testing.T,
	code string,
) (file source.File) {
	t.Helper()
	fileSet := token.NewFileSet()
	syntax, err := parser.ParseFile(
		fileSet,
		"p.go",
		code,
		parser.ParseComments,
	)
	if err != nil {
		t.Fatalf(
			"parse fixture: %v",
			err,
		)
	}

	return source.File{Path: "p.go", Rel: "dir/p.go", FileSet: fileSet, Syntax: syntax}
}
