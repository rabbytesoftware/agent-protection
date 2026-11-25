package stringcomment_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/style/internal/stringcomment"
)

func TestStringcommentProse(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{name: "parenthesized and concatenated text is joined", code: "package p\n\nfunc TestF(t *testing.T) {\n\tt.Log((\"one two three \") + \"four five six\")\n}\n", want: []string{"R07 p_test.go:4"}},
		{name: "quoted and punctuated words count", code: "package p\n\nfunc TestF(t *testing.T) {\n\tt.Log(\"(one) 'two' three, four; five! six?\")\n}\n", want: []string{"R07 p_test.go:4"}},
		{name: "identifiers, numbers and symbols do not count as words", code: "package p\n\nfunc TestF(t *testing.T) {\n\tt.Log(\"got %v want %v at 3 ==\")\n}\n", want: nil},
		{name: "subtraction of strings, runes and malformed literals are not text", code: "package p\n\nfunc TestF(t *testing.T) {\n\tt.Log(a - \"b\", 'c', \"one two three\" - \"four five six\")\n}\n", want: nil},
		{name: "five words pass", code: "package p\n\nfunc TestF(t *testing.T) {\n\tt.Log(\"one two three four five\")\n}\n", want: nil},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			stringcomment.Stringcomment([]source.File{parsed(
				t,
				"p_test.go",
				tc.code,
			)}),
			tc.want,
		)
	}
}
