package assertion_test

import "testing"

func TestAssertionProof(
	t *testing.T,
) {
	assertCases(
		t,
		[]fixture{
			{
				name: "an assertion reached through a chain of helpers passes",
				files: map[string]string{
					"p/a_test.go": "package p\n\nimport \"testing\"\n\nfunc first(\n\tt *testing.T,\n) {\n\tsecond(t)\n}\n\nfunc second(\n\tt *testing.T,\n) {\n\tthird(t)\n}\n\nfunc third(\n\tt *testing.T,\n) {\n\tt.Fatal(\"x\")\n}\n\nfunc TestChain(\n\tt *testing.T,\n) {\n\tfirst(t)\n}\n",
				},
			},
			{
				name: "mutually recursive helpers that never assert are reported",
				files: map[string]string{
					"p/a_test.go": "package p\n\nimport \"testing\"\n\nfunc ping(\n\tt *testing.T,\n) {\n\tpong(t)\n}\n\nfunc pong(\n\tt *testing.T,\n) {\n\tping(t)\n}\n\nfunc TestLoop(\n\tt *testing.T,\n) {\n\tping(t)\n}\n",
				},
				want: []string{
					"R42 p/a_test.go:17",
				},
			},
			{
				name: "t.Run with a named subtest that asserts passes",
				files: map[string]string{
					"p/a_test.go": "package p\n\nimport \"testing\"\n\nfunc subtest(\n\tt *testing.T,\n) {\n\tt.Fatal(\"x\")\n}\n\nfunc TestRun(\n\tt *testing.T,\n) {\n\tt.Run(\n\t\t\"x\",\n\t\tsubtest,\n\t)\n}\n",
				},
			},
			{
				name: "a helper handed a fresh testing value instead of t is reported",
				files: map[string]string{
					"p/a_test.go": "package p\n\nimport \"testing\"\n\nfunc check(\n\tt *testing.T,\n) {\n\tt.Fatal(\"x\")\n}\n\nfunc TestFresh(\n\tt *testing.T,\n) {\n\tcheck(&testing.T{})\n\tcheck((nil))\n}\n",
				},
				want: []string{
					"R42 p/a_test.go:11",
				},
			},
			{
				name: "a function value handed to a call that does not carry t is not reached",
				files: map[string]string{
					"p/a_test.go": "package p\n\nimport \"testing\"\n\nfunc subtest(\n\tt *testing.T,\n) {\n\tt.Fatal(\"x\")\n}\n\nfunc TestRun(\n\tt *testing.T,\n) {\n\tschedule(subtest)\n}\n",
				},
				want: []string{
					"R42 p/a_test.go:11",
				},
			},
		},
	)
}
