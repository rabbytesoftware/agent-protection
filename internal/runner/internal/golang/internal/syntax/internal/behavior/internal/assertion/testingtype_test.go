package assertion_test

import "testing"

func TestAssertionTestingTypes(
	t *testing.T,
) {
	assertCases(
		t,
		[]fixture{
			{
				name: "a helper taking testing.TB asserts for the test",
				files: map[string]string{
					"p/a_test.go": "package p\n\nimport \"testing\"\n\nfunc check(\n\ttb testing.TB,\n) {\n\ttb.Error(\"x\")\n}\n\nfunc TestSquare(\n\tt *testing.T,\n) {\n\tcheck(t)\n}\n",
				},
			},
			{
				name: "a helper taking a local interface with Fatalf asserts for the test",
				files: map[string]string{
					"p/a_test.go": "package p\n\nimport \"testing\"\n\nfunc check(\n\tf failer,\n) {\n\tf.Fatalf(\"x\")\n}\n\nfunc TestSquare(\n\tt *testing.T,\n) {\n\tcheck(t)\n}\n",
					"p/fail.go":   "package p\n\ntype failer interface {\n\tHelper()\n\tFatalf(format string, args ...any)\n}\n\ntype plain interface {\n\tName() string\n}\n\ntype number int\n",
				},
			},
			{
				name: "a helper reached through a qualified interface type asserts for the test",
				files: map[string]string{
					"p/a_test.go": "package p\n\nimport \"testing\"\n\nfunc check(\n\tf p.Failer,\n) {\n\tf.Fatalf(\"x\")\n}\n\nfunc TestSquare(\n\tt *testing.T,\n) {\n\tcheck(t)\n}\n",
					"p/fail.go":   "package p\n\ntype Failer interface {\n\tFatalf(format string, args ...any)\n}\n",
				},
			},
			{
				name: "a harness storing t in a field asserts through its methods",
				files: map[string]string{
					"p/a_test.go": "package p\n\nimport \"testing\"\n\ntype harness struct {\n\tt *testing.T\n\tname string\n}\n\nfunc (h harness) verify() {\n\th.t.Fatal(\"x\")\n}\n\nfunc TestSquare(\n\tt *testing.T,\n) {\n\th := harness{t: t}\n\th.verify()\n}\n",
				},
			},
			{
				name: "a suite failing through its T() asserts",
				files: map[string]string{
					"p/a_test.go": "package p\n\nimport \"testing\"\n\nfunc TestSquare(\n\tt *testing.T,\n) {\n\tholder := suite{t: t}\n\tholder.T().Errorf(\"x\")\n}\n",
				},
			},
			{
				name: "Error on a value that is not a testing value does not assert",
				files: map[string]string{
					"p/a_test.go": "package p\n\nimport \"testing\"\n\nfunc TestSquare(\n\tt *testing.T,\n) {\n\terr := run()\n\t_ = err.Error()\n\tholder.T(1).Fatal()\n\trun().Fatal()\n\tt.Run(\"x\", nothing)\n}\n",
				},
				want: []string{
					"R42 p/a_test.go:5",
				},
			},
		},
	)
}
