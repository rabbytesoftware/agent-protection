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
			"money/price.go",
			"package money\n\ntype Price int\n\nfunc (p Price) String() string {\n\tif p > 10 {\n\t\treturn \"big\"\n\t}\n\n\treturn \"small\"\n}\n",
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

func TestIsPlainRendering(
	t *testing.T,
) {
	cases := []struct {
		body string
		want bool
	}{
		{body: "return r.message", want: true},
		{body: "return \"failed: \" + r.message", want: true},
		{body: "return string(r.code)", want: true},
		{body: "return r.cause.Error()", want: true},
		{body: "return fmt.Sprintf(\"%s\", r.message)", want: true},
		{body: "return errors.Unwrap(r.cause)", want: true},
		{body: "return", want: true},
		{body: "return r.message, r.place", want: false},
		{body: "r.count++\nreturn r.message", want: false},
		{body: "panic(r.message)", want: false},
		{body: "return strings.ToUpper(r.message)", want: false},
		{body: "return fmt.Sprintf(\"%s\", strings.ToUpper(r.message))", want: false},
		{body: "return r.count * 2", want: false},
		{body: "return string(r.count * 2)", want: false},
		{body: "return r.names[0]", want: false},
		{body: "return r.cause.Error(r.message)", want: false},
		{body: "return format(r.message)", want: false},
		{body: "return r.format().Text", want: false},
	}
	for _, tc := range cases {
		assertPlainRendering(
			t,
			tc.body,
			tc.want,
		)
	}

	if !source.IsPlainRendering(nil) {
		t.Error("IsPlainRendering(nil) = false, want true for a body-less declaration")
	}
}

func assertPlainRendering(
	t *testing.T,
	body string,
	want bool,
) {
	t.Helper()
	file := parsedAt(
		t,
		"r.go",
		"package r\n\nfunc (r R) String() string {\n\t"+body+"\n}\n",
	)
	function := source.Nodes[*ast.FuncDecl](file.Syntax)[0]
	got := source.IsPlainRendering(function.Body)
	if got != want {
		t.Errorf(
			"IsPlainRendering(%q) = %v, want %v",
			body,
			got,
			want,
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
