package source_test

import (
	"go/ast"
	"go/parser"
	"go/token"
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

func TestNodes(
	t *testing.T,
) {
	syntax, err := parser.ParseFile(
		token.NewFileSet(),
		"p.go",
		"package p\n\nfunc a() {}\n\nfunc b() { _ = func() {} }\n",
		0,
	)
	if err != nil {
		t.Fatalf(
			"parse fixture: %v",
			err,
		)
	}

	declarations := source.Nodes[*ast.FuncDecl](syntax)
	literals := source.Nodes[*ast.FuncLit](syntax)
	countsMatch := len(declarations) == 2 && len(literals) == 1
	if !countsMatch {
		t.Errorf(
			"Nodes() found %d declarations and %d literals, want 2 and 1",
			len(declarations),
			len(literals),
		)
	}
}

func TestNodesKeepsSourceOrderAcrossNesting(
	t *testing.T,
) {
	syntax, err := parser.ParseFile(
		token.NewFileSet(),
		"p.go",
		"package p\n\nfunc f() {\n\tif true {\n\t\tif false {\n\t\t}\n\t}\n\tif true {\n\t}\n}\n",
		0,
	)
	if err != nil {
		t.Fatalf(
			"parse fixture: %v",
			err,
		)
	}

	positions := []token.Pos{}
	for _, statement := range source.Nodes[*ast.IfStmt](syntax) {
		positions = append(
			positions,
			statement.Pos(),
		)
	}

	isOrdered := len(positions) == 3 && slices.IsSorted(positions)
	if !isOrdered {
		t.Errorf(
			"Nodes() positions = %v, want 3 ascending positions",
			positions,
		)
	}
}

func TestDeclarationPackages(
	t *testing.T,
) {
	files := []source.File{
		parsedAt(
			t,
			"vocab/vocab.go",
			"package vocab\n\ntype Reservation struct{}\n\nconst Limit = 3\n\nvar (\n\tErrFull = errFull{}\n\t_       = Limit\n)\n\ntype errFull struct{}\n\nfunc (errFull) Error() string { return \"full\" }\n",
		),
		parsedAt(
			t,
			"vocab/more.go",
			"package vocab\n\ntype Seats int\n",
		),
		parsedAt(
			t,
			"engine/engine.go",
			"package engine\n\ntype Engine struct{}\n",
		),
		parsedAt(
			t,
			"engine/run.go",
			"package engine\n\nfunc Run() {}\n",
		),
		parsedAt(
			t,
			"state/state.go",
			"package state\n\nvar cache = map[int]int{}\n",
		),
		parsedAt(
			t,
			"methods/methods.go",
			"package methods\n\ntype T struct{}\n\nfunc (T) Run() {}\n",
		),
		parsedAt(
			t,
			"root.go",
			"package root\n\nimport \"fmt\"\n\nvar _ = fmt.Sprint\n",
		),
	}
	got := source.DeclarationPackages(files)
	want := []string{".", "vocab"}
	if !slices.Equal(
		got,
		want,
	) {
		t.Errorf(
			"DeclarationPackages() = %v, want %v",
			got,
			want,
		)
	}
}

func TestErrorMethods(
	t *testing.T,
) {
	if !slices.Equal(
		source.ErrorMethods(),
		[]string{
			"Error",
			"Unwrap",
			"Is",
			"As",
			"String",
		},
	) {
		t.Errorf(
			"ErrorMethods() = %v",
			source.ErrorMethods(),
		)
	}
}

func parsedAt(
	t *testing.T,
	rel string,
	code string,
) (file source.File) {
	t.Helper()
	file, err := source.Parse(
		"/repo",
		"/repo/"+rel,
		[]byte(code),
	)
	if err != nil {
		t.Fatalf(
			"parse %s: %v",
			rel,
			err,
		)
	}

	return file
}
