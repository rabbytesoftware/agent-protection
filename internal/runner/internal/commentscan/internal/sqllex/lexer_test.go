package sqllex_test

import (
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/sqllex"
	"testing"
)

func TestSQLLexer(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "line comment is reported", code: "-- TODO fix\nSELECT 1;\n", want: []int{1}},
		{name: "trailing comment is reported", code: "SELECT 1; -- why\n", want: []int{1}},
		{name: "block comment is reported", code: "/* why */ SELECT 1;\n", want: []int{1}},
		{name: "dashes in a string pass", code: "SELECT '--not';\n", want: nil},
		{name: "doubled quote escape", code: "SELECT 'it''s -- not';\n", want: nil},
		{name: "quoted identifier", code: "SELECT \"a--b\" FROM t;\n", want: nil},
		{name: "backtick identifier", code: "SELECT `a--b` FROM t;\n", want: nil},
		{name: "dollar quoted body", code: "CREATE FUNCTION f() AS $$\n-- not\n$$;\n-- real\n", want: []int{4}},
		{name: "tagged dollar quote", code: "DO $body$ -- not $body$; -- real\n", want: []int{1}},
		{name: "positional parameter is not a dollar quote", code: "SELECT $1; -- real\n", want: []int{1}},
		{name: "dollar inside an identifier", code: "SELECT a$b$c -- real\n", want: []int{1}},
		{name: "dollar tag without a closer", code: "SELECT $abc -- real\n", want: []int{1}},
		{name: "dollar tag with a space is not a tag", code: "SELECT $a b$ -- real\n", want: []int{1}},
		{name: "unterminated dollar quote runs to the end", code: "SELECT $$\n-- inside\n", want: nil},
		{name: "unterminated tagged dollar quote runs to the end", code: "SELECT $t$\n-- inside\n", want: nil},
		{name: "unterminated block comment is reported", code: "/* open\nSELECT 1;\n", want: []int{1}},
		{name: "minus operator passes", code: "SELECT 3 - 1;\n", want: nil},
		{name: "dollar at end of file is safe", code: "SELECT $", want: nil},
	}
	for _, tc := range cases {
		assertLines(
			t,
			tc.name,
			sqllex.Sqllex(),
			tc.code,
			tc.want,
		)
	}
}
