package source_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

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
