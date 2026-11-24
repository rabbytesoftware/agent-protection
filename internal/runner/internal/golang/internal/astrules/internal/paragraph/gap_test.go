package paragraph_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/paragraph"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestParagraphCommentLinesAreNotBlank(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{name: "comment between a block and a statement is not a blank line", code: "package p\n\nfunc f() {\n\tif a {\n\t}\n\t// Deprecated: g.\n\tg()\n}\n", want: []string{"R55 p.go:7"}},
		{name: "comment plus a blank line passes", code: "package p\n\nfunc f() {\n\tif a {\n\t}\n\n\t// Deprecated: g.\n\tg()\n}\n", want: nil},
		{name: "trailing comment on the closing brace is not a blank line", code: "package p\n\nfunc f() {\n\tif a {\n\t} // Deprecated: g.\n\tg()\n}\n", want: []string{"R55 p.go:6"}},
		{name: "comment between an assignment and its check keeps it attached", code: "package p\n\nfunc f() {\n\terr := g()\n\t// Deprecated: g.\n\tif err != nil {\n\t}\n}\n", want: nil},
		{name: "multi-line comment group fills the gap", code: "package p\n\nfunc f() {\n\tif a {\n\t}\n\t/* Deprecated:\n\tg. */\n\tg()\n}\n", want: []string{"R55 p.go:8"}},
	}
	for _, tc := range cases {
		assertLocations(
			t,
			tc.name,
			paragraph.Paragraph([]source.File{parsed(
				t,
				tc.code,
			)}),
			tc.want,
		)
	}
}
