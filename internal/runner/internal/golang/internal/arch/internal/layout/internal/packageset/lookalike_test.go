package packageset_test

import (
	"fmt"
	"path"
	"testing"
)

const (
	lookalikeSnippet = "R58 internal/%s %s and its sibling %s split one concept: merge them or name each for what it provides"
	split            = "split"
)

func TestLookalikeSiblings(
	t *testing.T,
) {
	cases := []struct {
		name    string
		first   string
		later   string
		verdict string
	}{
		{name: "noun and its -age form split one concept", first: "storage", later: "store", verdict: split},
		{name: "verb and its -ing form split one concept", first: "parse", later: "parsing", verdict: split},
		{name: "names sharing most of the shorter one split one concept", first: "parsed", later: "parser", verdict: split},
		{name: "plural and -ing forms split one concept", first: "versioning", later: "versions", verdict: split},
		{name: "old copy of a sibling splits one concept", first: "oldstore", later: "store", verdict: split},
		{name: "versioned copy of a sibling splits one concept", first: "ledger", later: "ledgerv2", verdict: split},
		{name: "related but distinct concepts pass", first: "server", later: "service"},
		{name: "words sharing a short prefix pass", first: "reader", later: "replayer"},
		{name: "abbreviation beside the full word passes", first: "dependencies", later: "deps"},
		{name: "version directories pass", first: "v1", later: "v2"},
		{name: "name that prefixes its sibling is reported", first: "io", later: "ion", verdict: split},
		{name: "unrelated short names pass", first: "auth", later: "audit"},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			checked(
				t,
				map[string]string{
					fixturePath(tc.first): "package " + tc.first + "\n",
					fixturePath(tc.later): "package " + tc.later + "\n",
				},
			),
			lookalikeWant(
				tc.first,
				tc.later,
				tc.verdict,
			),
		)
	}
}

func fixturePath(
	name string,
) (rel string) {
	return path.Join(
		"internal",
		name,
		"a.go",
	)
}

func lookalikeWant(
	first string,
	later string,
	verdict string,
) (want []string) {
	if verdict != split {
		return []string{}
	}

	return []string{fmt.Sprintf(
		lookalikeSnippet,
		later,
		later,
		first,
	)}
}
