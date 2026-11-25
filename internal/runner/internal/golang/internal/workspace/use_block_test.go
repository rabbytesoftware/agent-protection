package workspace_test

import "testing"

func TestWorkspaceReadsGoWorkUseBlocks(
	t *testing.T,
) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{name: "a use block with quoted and trailing-slash paths lists every member", files: map[string]string{"a/go.mod": "module a\n", "b/go.mod": "module b\n", "c/go.mod": "module c\n", "go.work": "go 1.24\n\n// workspace\nuse (\n\t\"./a\"\n\t`./b/`\n)\n\nreplace x => ./c\n"}, want: []string{"X a.go:1", "X a/a.go:1", "X a", "X b/a.go:1", "X b", "R70 c/go.mod"}},
		{name: "single-line uses list members and skip absolute paths and bare keywords", files: map[string]string{"tools/go.mod": "module example.com/tools\n", "go.work": "go 1.24\n\nuse .\nuse \"./tools\"\nuse /abs/elsewhere\nuse\n"}, want: []string{"X a.go:1", "X tools/a.go:1", "X tools"}},
	}
	for _, tc := range cases {
		assertWorkspace(
			t,
			tc.name,
			moduleWith(
				t,
				tc.files,
			),
			tc.want,
		)
	}
}
