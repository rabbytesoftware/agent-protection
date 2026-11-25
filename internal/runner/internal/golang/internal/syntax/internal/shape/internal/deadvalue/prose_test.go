package deadvalue_test

import "testing"

func TestDeadvalueProse(
	t *testing.T,
) {
	assertFixtures(
		t,
		[]fixture{
			{
				name:  "unread constant with a TODO marker is a comment",
				files: map[string]string{"internal/square/square.go": "package square\n\nconst todo = \"TODO: replace this hack\"\n"},
				want:  []string{"R07 internal/square/square.go:3 const todo holds prose nothing reads (\"TODO: replace this hack\"): a sentence parked in a string is a comment; delete it and make the code say it, or pass it to where a user sees it"},
			},
			{
				name:  "unread sentence is a comment and its excerpt is shortened",
				files: map[string]string{"internal/square/square.go": "package square\n\nconst rationale = \"factor stays at one because the legacy billing service \" + (\"double counts, remove after Q3 migration\")\n"},
				want:  []string{"R07 internal/square/square.go:3 const rationale holds prose nothing reads (\"factor stays at one because the legacy billing s...\"): a sentence parked in a string is a comment; delete it and make the code say it, or pass it to where a user sees it"},
			},
			{
				name:  "exported prose in an internal package nobody reads is a comment",
				files: map[string]string{"internal/square/square.go": "package square\n\nvar Why = `FIXME later`\n"},
				want:  []string{"R07 internal/square/square.go:3 var Why holds prose nothing reads (\"FIXME later\"): a sentence parked in a string is a comment; delete it and make the code say it, or pass it to where a user sees it"},
			},
			{
				name:  "a sentence passed to a user is a message",
				files: map[string]string{"internal/square/square.go": "package square\n\nimport \"fmt\"\n\nconst usage = \"usage: square takes one value and prints its square.\"\n\nfunc Print() {\n\tfmt.Println(usage)\n}\n"},
				want:  nil,
			},
			{
				name:  "short words, identifiers and non-string values are not prose",
				files: map[string]string{"internal/square/square.go": "package square\n\nconst (\n\tpattern = \"%s/%s: %d of %d (%v)\"\n\tlabel = \"square root\"\n\tsize = 3 + 4\n\tmixed = \"a\" + 1\n)\n"},
				want: []string{
					"R69 internal/square/square.go:4 const pattern is declared and never read: delete it; a value nothing reads is dead code, and what it was meant to say belongs in names, structure or the commit message",
					"R69 internal/square/square.go:5 const label is declared and never read: delete it; a value nothing reads is dead code, and what it was meant to say belongs in names, structure or the commit message",
					"R69 internal/square/square.go:6 const size is declared and never read: delete it; a value nothing reads is dead code, and what it was meant to say belongs in names, structure or the commit message",
					"R69 internal/square/square.go:7 const mixed is declared and never read: delete it; a value nothing reads is dead code, and what it was meant to say belongs in names, structure or the commit message",
				},
			},
		},
	)
}
