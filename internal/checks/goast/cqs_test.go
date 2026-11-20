package goast

import "testing"

func TestCQSMismatch(t *testing.T) {
	testCases := []struct {
		name string
		src  string
		want int
	}{
		{
			name: "returns value and mutates receiver is flagged",
			src: `package x
type T struct{ v int }
func (t *T) F() int {
	t.v = 5
	return t.v
}
`,
			want: 1,
		},
		{
			name: "pure query with no mutation is clean",
			src: `package x
type T struct{ v int }
func (t *T) F() int {
	return t.v
}
`,
			want: 0,
		},
		{
			name: "mutation with only error return is clean",
			src: `package x
type T struct{ v int }
func (t *T) F() error {
	t.v = 5
	return nil
}
`,
			want: 0,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			files := parseOne(t, dir, "f.go", tc.src)
			got := CQSMismatch(dir, files)
			if len(got) != tc.want {
				t.Errorf("CQSMismatch() = %d findings, want %d: %v", len(got), tc.want, got)
			}
		})
	}
}
