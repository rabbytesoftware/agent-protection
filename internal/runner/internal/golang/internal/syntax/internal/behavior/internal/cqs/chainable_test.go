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

func TestCqsStandardMethods(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{name: "stateful io.Writer Write passes", code: "package p\n\ntype W struct {\n\tpending []byte\n}\n\nfunc (w *W) Write(\n\tpayload []byte,\n) (n int, err error) {\n\tw.pending = append(w.pending, payload...)\n\treturn len(payload), nil\n}\n", want: nil},
		{name: "stateful io.Reader Read passes", code: "package p\n\ntype R struct {\n\toffset int\n}\n\nfunc (r *R) Read(\n\tbuffer []byte,\n) (int, error) {\n\tr.offset++\n\treturn 1, nil\n}\n", want: nil},
		{name: "io.Seeker Seek passes", code: "package p\n\ntype R struct {\n\toffset int64\n}\n\nfunc (r *R) Seek(\n\toffset int64,\n\twhence int,\n) (int64, error) {\n\tr.offset = offset\n\treturn offset, nil\n}\n", want: nil},
		{name: "io.ReaderFrom ReadFrom passes", code: "package p\n\nimport \"io\"\n\ntype B struct {\n\tn int64\n}\n\nfunc (b *B) ReadFrom(\n\tsource io.Reader,\n) (int64, error) {\n\tb.n = 1\n\treturn b.n, nil\n}\n", want: nil},
		{name: "io.RuneReader ReadRune passes", code: "package p\n\ntype R struct {\n\toffset int\n}\n\nfunc (r *R) ReadRune() (rune, int, error) {\n\tr.offset++\n\treturn 'a', 1, nil\n}\n", want: nil},
		{name: "heap.Interface Pop passes", code: "package p\n\ntype H struct {\n\titems []int\n}\n\nfunc (h *H) Pop() any {\n\tlast := h.items[len(h.items)-1]\n\th.items = h.items[:len(h.items)-1]\n\treturn last\n}\n", want: nil},
		{name: "scanner-style Next iterator passes", code: "package p\n\ntype I struct {\n\toffset int\n}\n\nfunc (i *I) Next() (advanced bool) {\n\ti.offset++\n\treturn true\n}\n", want: nil},
		{name: "Read with a non-standard signature is reported", code: "package p\n\ntype R struct {\n\toffset int\n}\n\nfunc (r *R) Read() (n int) {\n\tr.offset++\n\treturn r.offset\n}\n", want: []string{"p.go:7"}},
		{name: "Next returning a value besides the bool is reported", code: "package p\n\ntype I struct {\n\toffset int\n}\n\nfunc (i *I) Next() (int, bool) {\n\ti.offset++\n\treturn i.offset, true\n}\n", want: []string{"p.go:7"}},
		{name: "standard method name as a plain function is reported", code: "package p\n\nfunc Write(\n\ttarget *int,\n) (n int, err error) {\n\t*target = 1\n\treturn 1, nil\n}\n", want: []string{"p.go:3"}},
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
