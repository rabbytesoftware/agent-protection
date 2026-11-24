package cqs_test

import "testing"

func TestCqsClosures(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "returned closure that writes the receiver passes",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc (t *T) later() (run func()) {\n\treturn func() {\n\t\tt.v = 1\n\t}\n}\n",
			want: nil,
		},
		{
			name: "closure assigned to a variable and handed back passes",
			code: "package p\n\nimport \"sync\"\n\ntype T struct {\n\tv int\n}\n\nfunc (t *T) release() {\n\tt.v--\n}\n\nfunc (t *T) unlocker() (unlock func()) {\n\tvar once sync.Once\n\tfree := func() {\n\t\tt.release()\n\t}\n\treturn func() {\n\t\tonce.Do(free)\n\t}\n}\n",
			want: nil,
		},
		{
			name: "closure stored in a local field passes",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\ntype H struct {\n\trun func()\n}\n\nfunc (t *T) hook() (h *H) {\n\th = &H{}\n\th.run = func() {\n\t\tt.v++\n\t}\n\treturn h\n}\n",
			want: nil,
		},
		{
			name: "closure sent on a channel passes",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc (t *T) queue(\n\tjobs chan func(),\n) (n int) {\n\tjobs <- func() {\n\t\tt.v = 1\n\t}\n\treturn 1\n}\n",
			want: nil,
		},
		{
			name: "immediately invoked closure that writes the receiver is reported",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc (t *T) f() (n int) {\n\tfunc() {\n\t\tt.v = 1\n\t}()\n\treturn 1\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "parenthesized immediately invoked closure is reported",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc (t *T) f() (n int) {\n\t(func() {\n\t\tt.v = 1\n\t})()\n\treturn 1\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "deferred closure that writes the receiver is reported",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc (t *T) f() (n int) {\n\tdefer func() {\n\t\tt.v++\n\t}()\n\treturn 1\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "goroutine closure that writes the receiver is reported",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc (t *T) f() (n int) {\n\tgo func() {\n\t\tt.v++\n\t}()\n\treturn 1\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "closure passed as an argument that writes the receiver is reported",
			code: "package p\n\nimport \"sync\"\n\ntype T struct {\n\tonce sync.Once\n\tv\tint\n}\n\nfunc (t *T) f() (n int) {\n\tt.once.Do(func() {\n\t\tt.v = 1\n\t})\n\treturn t.v\n}\n",
			want: []string{"p.go:10"},
		},
		{
			name: "closure passed as an argument that calls a mutating helper is reported",
			code: "package p\n\nimport \"slices\"\n\ntype T struct {\n\ts []int\n\tn int\n}\n\nfunc (t *T) count() {\n\tt.n++\n}\n\nfunc (t *T) sorted() (s []int) {\n\tslices.SortFunc(t.s, func(a int, b int) (c int) {\n\t\tt.count()\n\t\treturn a - b\n\t})\n\treturn t.s\n}\n",
			want: []string{"p.go:14"},
		},
		{
			name: "invoked closure nested inside a returned closure passes",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc (t *T) later() (run func()) {\n\treturn func() {\n\t\tfunc() {\n\t\t\tt.v = 1\n\t\t}()\n\t}\n}\n",
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
