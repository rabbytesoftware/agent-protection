package goast

import "testing"

func TestHiddenSideEffect(t *testing.T) {
	testCases := []struct {
		name string
		src  string
		want int
	}{
		{
			name: "pure-sounding name that writes is flagged",
			src: `package x
import "os"
func GetConfig() {
	os.WriteFile("x", nil, 0o644)
}
`,
			want: 1,
		},
		{
			name: "pure-sounding name that only reads is clean",
			src: `package x
func GetValue() int {
	return 5
}
`,
			want: 0,
		},
		{
			name: "non-pure-sounding name that writes is clean (name already honest)",
			src: `package x
import "os"
func SaveConfig() {
	os.WriteFile("x", nil, 0o644)
}
`,
			want: 0,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			files := parseOne(t, dir, "f.go", tc.src)
			got := HiddenSideEffect(dir, files)
			if len(got) != tc.want {
				t.Errorf("HiddenSideEffect() = %d findings, want %d: %v", len(got), tc.want, got)
			}
		})
	}
}
