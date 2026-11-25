package deadvalue_test

import "testing"

func TestDeadvalueUsage(
	t *testing.T,
) {
	assertFixtures(
		t,
		[]fixture{
			{
				name:  "reads through composite literals, calls and returns pass",
				files: map[string]string{"internal/cube/cube.go": "package cube\n\nimport \"errors\"\n\nconst (\n\tlimit = 3\n\tname = \"cube\"\n\tmessage = \"cube is closed\"\n)\n\nvar errClosed = errors.New(message)\n\nfunc Run() (n int, label string, err error) {\n\tpoint := struct{ N int }{N: limit}\n\treturn point.N, name, errClosed\n}\n"},
				want:  nil,
			},
			{
				name:  "a selector of the same name does not read an unexported const",
				files: map[string]string{"internal/cube/cube.go": "package cube\n\nconst limit = 3\n\nfunc Run(p other.Point) (n int) {\n\treturn p.limit\n}\n"},
				want:  []string{"R69 internal/cube/cube.go:3 " + deadSnippet},
			},
			{
				name:  "a selector bound to blank is not a read of an exported const",
				files: map[string]string{"internal/cube/cube.go": "package cube\n\nconst Limit = 3\n", "internal/square/square.go": "package square\n\nimport \"example.com/m/internal/cube\"\n\nvar _ = cube.Limit\n"},
				want:  []string{"R69 internal/cube/cube.go:3 const Limit is declared and never read: delete it; a value nothing reads is dead code, and what it was meant to say belongs in names, structure or the commit message"},
			},
			{
				name:  "a mixed assignment with a named target is a read",
				files: map[string]string{"internal/cube/cube.go": "package cube\n\nconst limit = 3\n\nfunc Run() (n int) {\n\t_, n = 1, limit\n\treturn n\n}\n"},
				want:  nil,
			},
		},
	)
}
