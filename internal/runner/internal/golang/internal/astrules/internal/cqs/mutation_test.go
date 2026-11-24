package cqs_test

import "testing"

func TestCqsDirectWrites(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "incrementing a receiver field is reported",
			code: "package p\n\ntype T struct {\n\tn int\n}\n\nfunc (t *T) f() (n int) {\n\tt.n++\n\treturn t.n\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "writing into a receiver map field is reported",
			code: "package p\n\ntype T struct {\n\tm map[string]int\n}\n\nfunc (t *T) f() (n int) {\n\tt.m[\"k\"] = 1\n\treturn 1\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "assigning a nested receiver field is reported",
			code: "package p\n\ntype T struct {\n\tinner struct {\n\t\tv int\n\t}\n}\n\nfunc (t *T) f() (n int) {\n\tt.inner.v = 1\n\treturn 1\n}\n",
			want: []string{"p.go:9"},
		},
		{
			name: "assigning through a parenthesized dereference is reported",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc (t *T) f() (n int) {\n\t(*t).v = 1\n\treturn 1\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "deleting from a receiver map field is reported",
			code: "package p\n\ntype T struct {\n\tm map[string]int\n}\n\nfunc (t *T) f() (n int) {\n\tdelete(t.m, \"k\")\n\treturn 1\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "clearing a receiver field is reported",
			code: "package p\n\ntype T struct {\n\tm map[string]int\n}\n\nfunc (t *T) f() (n int) {\n\tclear(t.m)\n\treturn 1\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "copying into a receiver slice field is reported",
			code: "package p\n\ntype T struct {\n\ts []int\n}\n\nfunc (t *T) f(\n\tfrom []int,\n) (n int) {\n\treturn copy(t.s, from)\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "copying out of a receiver slice field passes",
			code: "package p\n\ntype T struct {\n\ts []int\n}\n\nfunc (t *T) f(\n\tto []int,\n) (n int) {\n\treturn copy(to, t.s)\n}\n",
			want: nil,
		},
		{
			name: "rebinding the pointer parameter itself passes",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc f(\n\tt *T,\n) (n int) {\n\tt = &T{}\n\t(t) = nil\n\treturn 1\n}\n",
			want: nil,
		},
		{
			name: "incrementing a local counter passes",
			code: "package p\n\nfunc f() (n int) {\n\tn++\n\treturn n\n}\n",
			want: nil,
		},
		{
			name: "writing a value receiver map field passes",
			code: "package p\n\ntype T struct {\n\tm map[string]int\n}\n\nfunc (t T) f() (n int) {\n\tt.m[\"k\"] = 1\n\treturn 1\n}\n",
			want: nil,
		},
	}
	for _, tc := range cases {
		assertCase(
			t,
			tc.name,
			tc.code,
			tc.want,
		)
	}
}
