package genericname_test

import (
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/genericname"
)

func TestWords(
	t *testing.T,
) {
	cases := []struct {
		name string
		text string
		want []string
	}{
		{name: "camelCase splits at humps", text: "runHandler", want: []string{"run", "handler"}},
		{name: "acronym ends before the next word", text: "HTTPServer", want: []string{"http", "server"}},
		{name: "trailing acronym stays whole", text: "AggregateID", want: []string{"aggregate", "id"}},
		{name: "digits are their own word", text: "handler2", want: []string{"handler", "2"}},
		{name: "prose splits at spaces and punctuation", text: "Get returns the fakeItem's value.", want: []string{"get", "returns", "the", "fake", "item", "s", "value"}},
		{name: "empty text has no words", text: "", want: nil},
	}
	for _, tc := range cases {
		assertWords(
			t,
			tc.name,
			tc.text,
			tc.want,
		)
	}
}

func TestStem(
	t *testing.T,
) {
	cases := []struct {
		name string
		word string
		want string
	}{
		{name: "plural drops its s", word: "reads", want: "read"},
		{name: "agent noun drops er", word: "reader", want: "read"},
		{name: "gerund drops ing", word: "reading", want: "read"},
		{name: "past tense and silent e meet", word: "created", want: "creat"},
		{name: "third person and silent e meet", word: "creates", want: "creat"},
		{name: "ies plural becomes y", word: "deliveries", want: "delivery"},
		{name: "doubled consonant is undone", word: "shipped", want: "ship"},
		{name: "or suffix drops like er", word: "error", want: "err"},
		{name: "double s is not a plural", word: "process", want: "process"},
		{name: "short words stay whole", word: "bus", want: "bus"},
		{name: "root never shrinks below three letters", word: "red", want: "red"},
	}
	for _, tc := range cases {
		assertStem(
			t,
			tc.name,
			tc.word,
			tc.want,
		)
	}
}

func assertWords(
	t *testing.T,
	name string,
	text string,
	want []string,
) {
	t.Helper()
	got := genericname.Words(text)
	if !slices.Equal(
		got,
		want,
	) {
		t.Errorf(
			"%s: Words(%q) = %v, want %v",
			name,
			text,
			got,
			want,
		)
	}
}

func assertStem(
	t *testing.T,
	name string,
	word string,
	want string,
) {
	t.Helper()
	got := genericname.Stem(word)
	if got != want {
		t.Errorf(
			"%s: Stem(%q) = %q, want %q",
			name,
			word,
			got,
			want,
		)
	}
}
