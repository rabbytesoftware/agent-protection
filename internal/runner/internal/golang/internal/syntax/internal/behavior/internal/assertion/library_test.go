package assertion_test

import "testing"

func TestAssertionLibraries(
	t *testing.T,
) {
	assertCases(
		t,
		[]fixture{
			{
				name: "testify assert and require calls receiving t pass",
				files: map[string]string{
					"p/a_test.go": "package p\n\nimport (\n\t\"testing\"\n\n\t\"github.com/stretchr/testify/assert\"\n\tmust \"github.com/stretchr/testify/require\"\n)\n\nfunc TestSquare(\n\tt *testing.T,\n) {\n\tassert.Equal(t, 9, square(3))\n}\n\nfunc TestCube(\n\tt *testing.T,\n) {\n\tmust.NoError(t, cube(3))\n}\n",
				},
			},
			{
				name: "a versioned assertion library path passes",
				files: map[string]string{
					"p/a_test.go": "package p\n\nimport (\n\t\"testing\"\n\n\t\"gotest.tools/v3/assert\"\n\t_ \"embed\"\n\t. \"strings\"\n)\n\nfunc TestSquare(\n\tt *testing.T,\n) {\n\tassert.Equal(t, 9, square(3))\n}\n",
				},
			},
			{
				name: "an assertion library call that does not receive t does not assert",
				files: map[string]string{
					"p/a_test.go": "package p\n\nimport (\n\t\"testing\"\n\n\t\"github.com/stretchr/testify/assert\"\n)\n\nfunc TestSquare(\n\tt *testing.T,\n) {\n\tassert.Equal(&testing.T{}, 9, square(3))\n\tassert.Equal(nil, 9, square(3))\n}\n",
				},
				want: []string{
					"R42 p/a_test.go:9",
				},
			},
			{
				name: "an unknown package receiving t does not assert",
				files: map[string]string{
					"p/a_test.go": "package p\n\nimport (\n\t\"testing\"\n\n\t\"example.com/vendor/checks\"\n\t\"example.com/vendor/v2\"\n)\n\nfunc TestSquare(\n\tt *testing.T,\n) {\n\tchecks.Square(t)\n\tvendor.Square(t)\n\tsquare(3)\n}\n",
				},
				want: []string{
					"R42 p/a_test.go:10",
				},
			},
		},
	)
}
