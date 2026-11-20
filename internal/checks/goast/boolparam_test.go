package goast

import "testing"

func parseOne(t *testing.T, dir, name, src string) []ParsedFile {
	t.Helper()
	writeFile(t, dir, name, src)
	files, err := WalkGoFiles(dir)
	if err != nil {
		t.Fatalf("WalkGoFiles: %v", err)
	}
	return files
}

func TestBoolParam(t *testing.T) {
	testCases := []struct {
		name string
		src  string
		want int
	}{
		{
			name: "function with bool param is flagged",
			src:  "package x\nfunc F(a int, validate bool) {}\n",
			want: 1,
		},
		{
			name: "function without bool param is clean",
			src:  "package x\nfunc F(a int, b string) {}\n",
			want: 0,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			files := parseOne(t, dir, "f.go", tc.src)
			got := BoolParam(dir, files)
			if len(got) != tc.want {
				t.Errorf("BoolParam() = %d findings, want %d: %v", len(got), tc.want, got)
			}
		})
	}
}
