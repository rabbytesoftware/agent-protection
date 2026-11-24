package source_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestSplice(
	t *testing.T,
) {
	cases := []struct {
		name  string
		edits []source.Edit
		want  string
	}{
		{name: "no edits keep the content", edits: nil, want: "abcdef"},
		{name: "insertion", edits: []source.Edit{{Start: 2, End: 2, Text: "X"}}, want: "abXcdef"},
		{name: "deletion", edits: []source.Edit{{Start: 1, End: 3}}, want: "adef"},
		{name: "edits apply in offset order", edits: []source.Edit{{Start: 5, End: 6, Text: "F"}, {Start: 0, End: 1, Text: "A"}}, want: "AbcdeF"},
		{name: "insertions at one offset keep their order", edits: []source.Edit{{Start: 3, End: 3, Text: "1"}, {Start: 3, End: 3, Text: "2"}}, want: "abc12def"},
		{name: "insertion before a deletion at the same offset", edits: []source.Edit{{Start: 2, End: 4}, {Start: 2, End: 2, Text: "X"}}, want: "abXef"},
		{name: "overlapping edit is dropped", edits: []source.Edit{{Start: 1, End: 4}, {Start: 2, End: 5, Text: "Y"}}, want: "aef"},
	}
	for _, tc := range cases {
		assertSplice(
			t,
			tc.name,
			tc.edits,
			tc.want,
		)
	}
}

func assertSplice(
	t *testing.T,
	name string,
	edits []source.Edit,
	want string,
) {
	t.Helper()
	got := string(source.Splice(
		[]byte("abcdef"),
		edits,
	))
	if got != want {
		t.Errorf(
			"%s: Splice() = %q, want %q",
			name,
			got,
			want,
		)
	}
}
