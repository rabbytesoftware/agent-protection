package cqs_test

import "testing"

func TestCqsChainable(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "chainable builder method returning its receiver passes",
			code: "package p\n\ntype B struct {\n\tv int\n}\n\nfunc (b *B) WithV(\n\tv int,\n) *B {\n\tb.v = v\n\tlater := func() (n int) {\n\t\treturn 1\n\t}\n\tb.v += later()\n\treturn b\n}\n",
			want: nil,
		},
		{
			name: "builder method that returns its receiver on one path only is reported",
			code: "package p\n\ntype B struct {\n\tv int\n}\n\nfunc (b *B) WithV(\n\tv int,\n) *B {\n\tb.v = v\n\tif v == 0 {\n\t\treturn &B{}\n\t}\n\n\treturn b\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "returning the receiver together with another value is reported",
			code: "package p\n\ntype B struct {\n\tv int\n}\n\nfunc (b *B) WithV() (*B, int) {\n\tb.v = 1\n\treturn b, b.v\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "chainable builder method mutating through a helper passes",
			code: "package p\n\ntype B struct {\n\tv int\n}\n\nfunc (b *B) set() {\n\tb.v = 1\n}\n\nfunc (b *B) WithV() *B {\n\tb.set()\n\treturn b\n}\n",
			want: nil,
		},
		{
			name: "method with an unnamed receiver is no chainable builder",
			code: "package p\n\ntype B struct {\n\tv int\n}\n\nfunc (*B) f(\n\ttarget *int,\n) (n int) {\n\t*target = 1\n\treturn 1\n}\n",
			want: []string{"p.go:7"},
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
