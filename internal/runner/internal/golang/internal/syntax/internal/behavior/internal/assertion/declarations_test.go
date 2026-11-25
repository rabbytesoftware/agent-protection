package assertion_test

import "testing"

func TestAssertionDeclarations(
	t *testing.T,
) {
	assertCases(
		t,
		[]fixture{
			{
				name: "a package-level func literal helper asserts",
				files: map[string]string{
					"p/a_test.go": "package p\n\nimport \"testing\"\n\nvar check = func(\n\tt *testing.T,\n) {\n\tt.Fatal(\"x\")\n}\n\nvar count, label = 1, \"x\"\n\nfunc external()\n\nfunc TestSquare(\n\tt *testing.T,\n) {\n\tcheck(t)\n}\n",
				},
			},
			{
				name: "a scenario table dispatched through a function value asserts",
				files: map[string]string{
					"p/a_test.go": "package p\n\nimport \"testing\"\n\ntype scenario func(\n\tt testing.TB,\n)\n\nfunc unknown(\n\tt testing.TB,\n) {\n\tt.Fatal(\"x\")\n}\n\nfunc scenarios() map[string]scenario {\n\treturn map[string]scenario{\"unknown\": unknown}\n}\n\nfunc run(\n\tt *testing.T,\n\tfn scenario,\n) {\n\tfn(t)\n}\n\nfunc TestScenarios(\n\tt *testing.T,\n) {\n\tfor _, fn := range scenarios() {\n\t\trun(\n\t\t\tt,\n\t\t\tfn,\n\t\t)\n\t}\n}\n",
				},
			},
			{
				name: "scenario builders holding asserting literals are dispatch targets",
				files: map[string]string{
					"p/a_test.go": "package p\n\nimport \"testing\"\n\nfunc scenarios() []func(*testing.T) {\n\treturn []func(*testing.T){func(t *testing.T) {\n\t\tt.Fatal(\"x\")\n\t}}\n}\n\nfunc TestScenarios(\n\tt *testing.T,\n) {\n\tfor _, fn := range scenarios() {\n\t\tfn(t)\n\t}\n}\n",
				},
			},
			{
				name: "a method value carrying t asserts",
				files: map[string]string{
					"p/a_test.go": "package p\n\nimport \"testing\"\n\ntype forbidden struct {\n\tt *testing.T\n}\n\nfunc (f forbidden) fail() {\n\tf.t.Fatal(\"ran\")\n}\n\nfunc TestHandler(\n\tt *testing.T,\n) {\n\tsubscribe(forbidden{t: t}.fail)\n}\n",
				},
			},
			{
				name: "a dynamic call that does not receive t reaches nothing",
				files: map[string]string{
					"p/a_test.go": "package p\n\nimport \"testing\"\n\nfunc helper(\n\tt *testing.T,\n) {\n\tt.Fatal(\"x\")\n}\n\nfunc TestOther(\n\tt *testing.T,\n) {\n\tt.Run(\"x\", func(t *testing.T) {\n\t\tt.Fatal(\"x\")\n\t})\n}\n\nfunc TestSquare(\n\tt *testing.T,\n) {\n\tvar fn func()\n\tfn()\n\t_ = helper\n}\n",
				},
				want: []string{
					"R42 p/a_test.go:19",
				},
			},
		},
	)
}
