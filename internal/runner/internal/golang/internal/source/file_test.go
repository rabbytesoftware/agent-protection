package source_test

import (
	"go/ast"
	"slices"
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

func TestFileOffsets(
	t *testing.T,
) {
	file := parsedFile(
		t,
		"package p\n\nfunc f() {\n}\n",
	)
	declaration := source.Nodes[*ast.FuncDecl](file.Syntax)[0]
	offsets := []int{
		file.Offset(declaration.Pos()),
		file.LineOffset(3),
		file.LineOffset(4),
	}
	if !slices.Equal(
		offsets,
		[]int{11, 11, 22},
	) {
		t.Errorf(
			"offsets = %v, want [11 11 22]",
			offsets,
		)
	}
}

func parsedFile(
	t *testing.T,
	code string,
) (file source.File) {
	t.Helper()
	file, err := source.Parse(
		"",
		"dir/p.go",
		[]byte(code),
	)
	if err != nil {
		t.Fatalf(
			"parse fixture: %v",
			err,
		)
	}

	return file
}
