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
			code: "package p\n\ntype T struct {\n\ts []int\n}\n\nfunc (t *T) f() (n int) {\n\tto := make([]int, len(t.s))\n\treturn copy(to, t.s)\n}\n",
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

func TestCqsSharedWrites(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "writing a value receiver map field is reported",
			code: "package p\n\ntype T struct {\n\tm map[string]int\n}\n\nfunc (t T) f() (n int) {\n\tt.m[\"k\"]++\n\treturn len(t.m)\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "incrementing a map parameter entry is reported",
			code: "package p\n\nfunc f(\n\ttally map[string]int,\n) (n int) {\n\ttally[\"k\"]++\n\treturn len(tally)\n}\n",
			want: []string{"p.go:3"},
		},
		{
			name: "sending on a channel parameter is reported",
			code: "package p\n\nfunc f(\n\tevents chan<- string,\n) (n int) {\n\tevents <- \"a\"\n\treturn len(events)\n}\n",
			want: []string{"p.go:3"},
		},
		{
			name: "deleting from a map parameter is reported",
			code: "package p\n\nfunc f(\n\tm map[string]int,\n) (n int) {\n\tdelete(m, \"k\")\n\treturn len(m)\n}\n",
			want: []string{"p.go:3"},
		},
		{
			name: "clearing a slice parameter is reported",
			code: "package p\n\nfunc f(\n\tvalues []int,\n) (n int) {\n\tclear(values)\n\treturn len(values)\n}\n",
			want: []string{"p.go:3"},
		},
		{
			name: "writing a slice parameter element is reported",
			code: "package p\n\nfunc f(\n\tvalues []int,\n) (n int) {\n\tvalues[0] = 1\n\treturn len(values)\n}\n",
			want: []string{"p.go:3"},
		},
		{
			name: "appending to a value receiver field is reported",
			code: "package p\n\ntype T struct {\n\ts []int\n}\n\nfunc (t T) f() (n int) {\n\tt.s = append(t.s, 1)\n\treturn len(t.s)\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "sending on a value receiver channel field is reported",
			code: "package p\n\ntype T struct {\n\tc chan int\n}\n\nfunc (t T) f() (n int) {\n\tt.c <- 1\n\treturn len(t.c)\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "sending on a pointer receiver channel field is reported",
			code: "package p\n\ntype T struct {\n\tc chan int\n}\n\nfunc (t *T) f() (n int) {\n\tt.c <- 1\n\treturn len(t.c)\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "passing a map parameter to a helper that writes it is reported",
			code: "package p\n\nfunc put(\n\tm map[string]int,\n) {\n\tm[\"k\"] = 1\n}\n\nfunc f(\n\tm map[string]int,\n) (n int) {\n\tput(m)\n\treturn len(m)\n}\n",
			want: []string{"p.go:9"},
		},
		{
			name: "writing an array parameter copy passes",
			code: "package p\n\nfunc f(\n\tvalues [3]int,\n) (n int) {\n\tvalues[0] = 1\n\treturn values[0]\n}\n",
			want: nil,
		},
		{
			name: "rebinding a slice parameter with append passes",
			code: "package p\n\nfunc f(\n\tvalues []int,\n) (n int) {\n\tvalues = append(values, 1)\n\treturn len(values)\n}\n",
			want: nil,
		},
		{
			name: "reading a map parameter passes",
			code: "package p\n\nfunc f(\n\tm map[string]int,\n) (n int) {\n\treturn m[\"k\"]\n}\n",
			want: nil,
		},
		{
			name: "storing and returning a created value in a value receiver map passes",
			code: "package p\n\ntype Item struct {\n\tname string\n}\n\ntype T struct {\n\tm map[string]Item\n}\n\nfunc (t T) Add(\n\tname string,\n) (item Item) {\n\titem = Item{name: name}\n\tt.m[name] = item\n\treturn item\n}\n",
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
