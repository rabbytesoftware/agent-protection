package tsast

import "testing"

func TestExtractJSON(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{name: "pure JSON array", input: `[{"a":1}]`, want: `[{"a":1}]`},
		{name: "banner text before JSON", input: "npm notice\nsome warning\n[{\"a\":1}]", want: `[{"a":1}]`},
		{name: "no JSON present falls back to original", input: "no json here", want: "no json here"},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := string(extractJSON([]byte(tc.input)))
			if got != tc.want {
				t.Errorf("extractJSON(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestRelPath(t *testing.T) {
	testCases := []struct {
		name string
		root string
		path string
		want string
	}{
		{name: "nested file", root: "/a/b", path: "/a/b/src/x.ts", want: "src/x.ts"},
		{name: "same dir", root: "/a/b", path: "/a/b/x.ts", want: "x.ts"},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := RelPath(tc.root, tc.path)
			if got != tc.want {
				t.Errorf("RelPath(%q, %q) = %q, want %q", tc.root, tc.path, got, tc.want)
			}
		})
	}
}

func TestBunBinary(t *testing.T) {
	got := bunBinary()
	if got == "" {
		t.Error("bunBinary() returned empty string")
	}
}
