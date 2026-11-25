package assertion_test

import "testing"

func TestAssertionCatalog(
	t *testing.T,
) {
	assertCases(
		t,
		[]fixture{
			{
				name: "a helper in a test-support package resolved by import path asserts",
				files: map[string]string{
					"store/store_test.go":    "package store\n\nimport (\n\t\"testing\"\n\n\t\"example.com/m/store/storetest\"\n)\n\nfunc TestStore(\n\tt *testing.T,\n) {\n\tstoretest.Run(t)\n}\n",
					"store/storetest/run.go": "package storetest\n\nimport \"testing\"\n\nfunc Run(\n\tt *testing.T,\n) {\n\tt.Fatal(\"x\")\n}\n",
				},
			},
			{
				name: "a helper in another package that never asserts is reported",
				files: map[string]string{
					"store/store_test.go":    "package store\n\nimport (\n\t\"testing\"\n\n\t\"example.com/m/store/storetest\"\n)\n\nfunc TestStore(\n\tt *testing.T,\n) {\n\tstoretest.Run(t)\n}\n",
					"store/storetest/run.go": "package storetest\n\nimport \"testing\"\n\nfunc Run(\n\tt *testing.T,\n) {\n\tt.Log(\"x\")\n}\n",
				},
				want: []string{
					"R42 store/store_test.go:9",
				},
			},
			{
				name: "a method helper and a generic helper receiving t assert",
				files: map[string]string{
					"p/a_test.go": "package p\n\nimport \"testing\"\n\ntype fixture struct{}\n\nfunc (fixture) verify(\n\tt *testing.T,\n) {\n\tt.Fatal(\"x\")\n}\n\nfunc check[V any](\n\tt *testing.T,\n\tvalue V,\n) {\n\tt.Fatal(value)\n}\n\nfunc TestMethod(\n\tt *testing.T,\n) {\n\tfixture{}.verify(t)\n}\n\nfunc TestGeneric(\n\tt *testing.T,\n) {\n\tcheck[int](\n\t\tt,\n\t\t3,\n\t)\n}\n",
				},
			},
			{
				name: "a helper in a sibling directory of the same name is not reached",
				files: map[string]string{
					"p/a_test.go": "package p\n\nimport \"testing\"\n\nfunc TestSquare(\n\tt *testing.T,\n) {\n\tcheck(t)\n}\n",
					"q/check.go":  "package q\n\nimport \"testing\"\n\nfunc check(\n\tt *testing.T,\n) {\n\tt.Fatal(\"x\")\n}\n",
				},
				want: []string{
					"R42 p/a_test.go:5",
				},
			},
			{
				name: "a module root package import resolves to nothing",
				files: map[string]string{
					"p/a_test.go": "package p\n\nimport (\n\t\"testing\"\n\n\t\"example.com/m\"\n)\n\nfunc TestSquare(\n\tt *testing.T,\n) {\n\tm.Check(t)\n}\n",
					"root.go":     "package m\n\nimport \"testing\"\n\nfunc Check(\n\tt *testing.T,\n) {\n\tt.Fatal(\"x\")\n}\n",
				},
				want: []string{
					"R42 p/a_test.go:9",
				},
			},
		},
	)
}
