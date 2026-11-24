package cqs_test

import "testing"

func TestCqsResolvesCallees(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "calling a mutating method on a generic receiver is reported",
			code: "package p\n\ntype S[K comparable] struct {\n\tm map[K]int\n}\n\nfunc (s *S[K]) put(\n\tk K,\n) {\n\ts.m[k] = 1\n}\n\nfunc (s *S[K]) Put(\n\tk K,\n) (n int) {\n\ts.put(k)\n\treturn len(s.m)\n}\n",
			want: []string{"p.go:13"},
		},
		{
			name: "calling a mutating method on a two-parameter generic receiver is reported",
			code: "package p\n\ntype S[K comparable, V any] struct {\n\tm map[K]V\n}\n\nfunc (s *S[K, V]) drop(\n\tk K,\n) {\n\tdelete(s.m, k)\n}\n\nfunc (s *S[K, V]) Drop(\n\tk K,\n) (n int) {\n\ts.drop(k)\n\treturn len(s.m)\n}\n",
			want: []string{"p.go:13"},
		},
		{
			name: "calling a value receiver method that writes its copy passes",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc (t T) bump() {\n\tt.v++\n}\n\nfunc (t *T) Get() (v int) {\n\tt.bump()\n\treturn t.v\n}\n",
			want: nil,
		},
		{
			name: "mutated parameter position follows blank and grouped parameters",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc set(\n\t_ int,\n\ta, b *T,\n) {\n\tb.v = 1\n}\n\nfunc (t *T) First() (v int) {\n\tset(0, t, nil)\n\treturn 1\n}\n\nfunc (t *T) Second() (v int) {\n\tset(0, nil, t)\n\treturn 1\n}\n",
			want: []string{"p.go:19"},
		},
		{
			name: "method of a type from another package resolves to nothing",
			code: "package p\n\nimport \"strings\"\n\nfunc (b *T) grow() {\n\tb.v++\n}\n\ntype T struct {\n\tv int\n}\n\nfunc f(\n\tb *strings.Builder,\n) (n int) {\n\tb.grow()\n\treturn 1\n}\n",
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
