package source_test

import (
	"go/ast"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestBreaks(
	t *testing.T,
) {
	cases := []struct {
		name string
		call string
		want []string
	}{
		{name: "no items need no break", call: "g()", want: nil},
		{name: "one item per line needs no break", call: "g(\n\t1,\n\t2,\n)", want: nil},
		{name: "inline items break before each item and the closing bracket", call: "g(1, 2)", want: []string{"\n", "\n", ",\n"}},
		{name: "comma in a comment is no trailing comma", call: "g(\n\t1,\n\t2 /* , */)", want: []string{",\n"}},
		{name: "existing trailing comma is reused", call: "g(\n\t1,\n\t2, /* last */)", want: []string{"\n"}},
		{name: "spread before the closing bracket", call: "g(\n\t1,\n\ty...)", want: []string{",\n"}},
	}
	for _, tc := range cases {
		file := parsedFile(
			t,
			"package p\n\nvar x = "+tc.call+"\n",
		)
		call := source.Nodes[*ast.CallExpr](file.Syntax)[0]
		assertBreaks(
			t,
			tc.name,
			source.Breaks(
				file,
				call.Lparen,
				extents(call),
				call.Rparen,
			),
			tc.want,
		)
	}
}

func extents(
	call *ast.CallExpr,
) (items []source.Extent) {
	for _, argument := range call.Args {
		items = append(
			items,
			source.Extent{Start: argument.Pos(), End: argument.End()},
		)
	}

	return items
}

func assertBreaks(
	t *testing.T,
	name string,
	edits []source.Edit,
	want []string,
) {
	t.Helper()
	got := []string{}
	for _, edit := range edits {
		got = append(
			got,
			edit.Text,
		)
	}

	isMismatch := !slices.Equal(
		got,
		want,
	) && len(got)+len(want) > 0
	if isMismatch {
		t.Errorf(
			"%s: Breaks() texts = %q, want %q",
			name,
			got,
			want,
		)
	}
}
