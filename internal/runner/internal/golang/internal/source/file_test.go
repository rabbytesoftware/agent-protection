package source_test

import (
	"go/ast"
	"maps"
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

func TestMockFiles(
	t *testing.T,
) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
		users map[string]string
	}{
		{
			name: "mock referenced only from tests keeps its exemption",
			files: map[string]string{
				"store/mock_store.go":  "package store\n\ntype mockStore struct{}\n",
				"store/store.go":       "package store\n\ntype Store struct{}\n",
				"store/store_test.go":  "package store\n\nvar _ = mockStore{}\n",
				"store/mocks/bus.go":   "package mocks\n\ntype Bus struct{}\n",
				"store/support/sup.go": "package support\n\nimport (\n\t\"testing\"\n\n\t\"example.com/m/store/mocks\"\n)\n\nvar _ = mocks.Bus{}\n\nfunc Helper(t *testing.T) {}\n",
			},
			want:  []string{"store/mock_store.go", "store/mocks/bus.go"},
			users: map[string]string{},
		},
		{
			name: "mock called from a production file of its package loses its exemption",
			files: map[string]string{
				"fake.go":      "package arith\n\ntype fake struct{}\n\nfunc newFake() (f *fake) {\n\treturn &fake{}\n}\n",
				"container.go": "package arith\n\nfunc Arith() (f *fake) {\n\treturn newFake()\n}\n",
			},
			want:  nil,
			users: map[string]string{"fake.go": "container.go"},
		},
		{
			name: "mock imported by a production package loses its exemption, also through a dot import",
			files: map[string]string{
				"internal/bus/mocks/bus.go": "package mocks\n\ntype Bus struct{}\n",
				"internal/bus/fake_bus.go":  "package bus\n\nfunc Fake() {}\n",
				"internal/app/app.go":       "package app\n\nimport (\n\tm \"example.com/m/internal/bus/mocks\"\n\t. \"example.com/m/internal/bus\"\n)\n\nvar _ = m.Bus{}\n\nfunc Run() {\n\tFake()\n}\n",
			},
			want:  nil,
			users: map[string]string{"internal/bus/fake_bus.go": "internal/app/app.go", "internal/bus/mocks/bus.go": "internal/app/app.go"},
		},
		{
			name: "mock used by another mock that production uses loses its exemption",
			files: map[string]string{
				"mock_clock.go": "package app\n\ntype clock struct{}\n",
				"mock_timer.go": "package app\n\ntype timer struct{ c clock }\n",
				"app.go":        "package app\n\nvar _ = timer{}\n",
			},
			want:  nil,
			users: map[string]string{"mock_clock.go": "mock_timer.go", "mock_timer.go": "app.go"},
		},
		{
			name: "method a mock file adds to a production type counts once production calls it",
			files: map[string]string{
				"service.go":      "package app\n\ntype Service struct{}\n\nfunc Run() {\n\tService{}.wire()\n}\n",
				"mock_service.go": "package app\n\nfunc (Service) wire() {}\n",
			},
			want:  nil,
			users: map[string]string{"mock_service.go": "service.go"},
		},
		{
			name: "same names in other packages, selectors on locals, blank imports and unexported names do not count",
			files: map[string]string{
				"internal/bus/fake_bus.go": "package bus\n\ntype Bus struct{}\n\nfunc fake() {}\n",
				"internal/app/app.go":      "package app\n\nimport (\n\t_ \"example.com/m/internal/bus\"\n\tother \"example.com/m/internal/other\"\n)\n\ntype Bus struct{}\n\nfunc fake() {\n\tbus := Bus{}\n\t_ = bus.Bus\n\t_ = other.Bus\n}\n",
				"internal/web/web.go":      "package web\n\nimport \"example.com/m/internal/bus\"\n\nvar _ = bus.fake\n",
				"mock_root.go":             "package root\n\ntype Root struct{}\n",
				"internal/use/use.go":      "package use\n\nimport \"example.com/m\"\n\nvar _ = m.Root{}\n",
			},
			want:  []string{"internal/bus/fake_bus.go", "mock_root.go"},
			users: map[string]string{},
		},
	}
	for _, tc := range cases {
		files := parsedFiles(
			t,
			tc.files,
		)
		assertMocks(
			t,
			tc.name,
			files,
			tc.want,
			tc.users,
		)
	}
}

func assertMocks(
	t *testing.T,
	name string,
	files []source.File,
	want []string,
	users map[string]string,
) {
	t.Helper()
	got := source.MockFiles(files)
	slices.Sort(got)
	if !slices.Equal(
		got,
		want,
	) {
		t.Errorf(
			"%s: MockFiles() = %v, want %v",
			name,
			got,
			want,
		)
	}

	gotUsers := source.MockUsers(files)
	if !maps.Equal(
		gotUsers,
		users,
	) {
		t.Errorf(
			"%s: MockUsers() = %v, want %v",
			name,
			gotUsers,
			users,
		)
	}

	kept := len(source.WithoutMocks(files))
	if kept != len(files)-len(want) {
		t.Errorf(
			"%s: WithoutMocks() kept %d files, want %d",
			name,
			kept,
			len(files)-len(want),
		)
	}
}

func parsedFiles(
	t *testing.T,
	sources map[string]string,
) (files []source.File) {
	t.Helper()
	for _, rel := range slices.Sorted(maps.Keys(sources)) {
		files = append(
			files,
			parsedAt(
				t,
				rel,
				sources[rel],
			),
		)
	}

	return files
}
