package publicapi_test

import "testing"

func TestPublicapiErrorMethodBodies(
	t *testing.T,
) {
	cases := []struct {
		name string
		body string
		want []string
	}{
		{name: "field return passes", body: "return r.message", want: nil},
		{name: "literal return passes", body: "return \"failed\"", want: nil},
		{name: "concatenation passes", body: "return \"failed: \" + r.message + \" at \" + r.place", want: nil},
		{name: "conversion of the receiver passes", body: "return string(r.message)", want: nil},
		{name: "wrapped error rendering passes", body: "return r.message + \": \" + r.cause.Error()", want: nil},
		{name: "one fmt.Sprintf call passes", body: "return fmt.Sprintf(\"%s at %s\", r.message, r.place)", want: nil},
		{name: "one errors call passes", body: "return errors.Unwrap(r.cause)", want: nil},
		{name: "loop is reported", body: "for range r.message {\n}\nreturn \"\"", want: []string{"R14 receipt.go:9"}},
		{name: "branch is reported", body: "if r.message == \"\" {\nreturn \"empty\"\n}\nreturn r.message", want: []string{"R14 receipt.go:9"}},
		{name: "arithmetic is reported", body: "return strconv.Itoa(len(r.message) * 2)", want: []string{"R14 receipt.go:9"}},
		{name: "nested call in Sprintf is reported", body: "return fmt.Sprintf(\"%s\", strings.ToUpper(r.message))", want: []string{"R14 receipt.go:9"}},
		{name: "comparison is reported", body: "return r.message == r.place", want: []string{"R14 receipt.go:9"}},
		{name: "two returns are reported", body: "return r.message, r.place", want: []string{"R14 receipt.go:9"}},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			checked(
				t,
				map[string]string{"receipt.go": receipt(tc.body)},
			),
			tc.want,
		)
	}
}

func TestPublicapiErrorMethodSnippet(
	t *testing.T,
) {
	assertSnippet(
		t,
		checked(
			t,
			map[string]string{"receipt.go": receipt("for range r.message {\n}\nreturn \"\"")},
		),
		"receipt.go:9",
		"error method Receipt.String in public package arith: Error, Unwrap, Is, As and String hold a single return of a field, a literal, a concatenation or one fmt.Sprintf, fmt.Errorf or errors call; move the logic under internal/",
	)
}

func TestPublicapiErrorMethodInVocabularyPackage(
	t *testing.T,
) {
	assertFindings(
		t,
		"error method with logic in a vocabulary package",
		checked(
			t,
			map[string]string{"internal/models/receipt.go": "package models\n\ntype Receipt struct{ message string }\n\nfunc (r Receipt) Error() string {\n\tif r.message == \"\" {\n\t\treturn \"empty\"\n\t}\n\treturn r.message\n}\n"},
		),
		[]string{"R57 internal/models/receipt.go:5"},
	)
}

func TestPublicapiRenderingLogicAtTheRoot(
	t *testing.T,
) {
	assertFindings(
		t,
		"String method computing a receipt at the module root",
		checked(
			t,
			map[string]string{
				"receipt.go":      "package arith\n\nimport (\n\t\"strconv\"\n\t\"strings\"\n)\n\ntype Receipt []int\n\nconst (\n\tTaxPercent = 21\n\tPercent    = 100\n)\n\nfunc (r Receipt) String() (rendered string) {\n\tlines := []string{}\n\ttotal := 0\n\tfor _, amount := range r {\n\t\ttaxed := amount + amount*TaxPercent/Percent\n\t\ttotal += taxed\n\t\tlines = append(\n\t\t\tlines,\n\t\t\tstrconv.Itoa(taxed),\n\t\t)\n\t}\n\n\treturn strings.Join(\n\t\tappend(\n\t\t\tlines,\n\t\t\t\"total \"+strconv.Itoa(total),\n\t\t),\n\t\t\"\\n\",\n\t)\n}\n",
				"internal/a/a.go": "package a\n",
				"internal/b/b.go": "package b\n",
			},
		),
		[]string{"R14 receipt.go:15"},
	)
}

func receipt(
	body string,
) (code string) {
	return "package arith\n\nimport \"fmt\"\n\nvar _ = fmt.Sprint\n\ntype Receipt struct{ message, place string; cause error }\n\nfunc (r Receipt) String() string {\n" + body + "\n}\n"
}
