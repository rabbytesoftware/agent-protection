package paramperline_test

import (
	"go/format"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/style/internal/paramperline"
)

func TestEdits(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want string
	}{
		{
			name: "compliant signature is left alone",
			code: "package p\n\nfunc f(\n\ta int,\n) {\n}\n",
			want: "package p\n\nfunc f(\n\ta int,\n) {\n}\n",
		},
		{
			name: "empty parameter list is left alone",
			code: "package p\n\nfunc f() {\n}\n",
			want: "package p\n\nfunc f() {\n}\n",
		},
		{
			name: "inline parameters get a line each and a trailing comma",
			code: "package p\n\nfunc f(a int, b string) {\n}\n",
			want: "package p\n\nfunc f(\n\ta int,\n\tb string,\n) {\n}\n",
		},
		{
			name: "names sharing a type are split but keep sharing it",
			code: "package p\n\nfunc f(a, b int) {\n}\n",
			want: "package p\n\nfunc f(\n\ta,\n\tb int,\n) {\n}\n",
		},
		{
			name: "existing trailing comma is kept",
			code: "package p\n\nfunc f(\n\ta int,\n\tb int,) {\n}\n",
			want: "package p\n\nfunc f(\n\ta int,\n\tb int,\n) {\n}\n",
		},
		{
			name: "type parameters and results stay inline",
			code: "package p\n\nfunc f[K comparable, V any](key K, value V) (k K, v V) {\n\treturn key, value\n}\n",
			want: "package p\n\nfunc f[K comparable, V any](\n\tkey K,\n\tvalue V,\n) (k K, v V) {\n\treturn key, value\n}\n",
		},
		{
			name: "variadic parameter",
			code: "package p\n\nfunc f(format string, args ...any) {\n}\n",
			want: "package p\n\nfunc f(\n\tformat string,\n\targs ...any,\n) {\n}\n",
		},
		{
			name: "func literal, interface method and func type",
			code: "package p\n\nvar f = func(a int) {\n}\n\ntype i interface {\n\tm(a int) error\n}\n\ntype h func(int, string)\n",
			want: "package p\n\nvar f = func(\n\ta int,\n) {\n}\n\ntype i interface {\n\tm(\n\t\ta int,\n\t) error\n}\n\ntype h func(\n\tint,\n\tstring,\n)\n",
		},
		{
			name: "func typed parameter keeps its own inline parameters",
			code: "package p\n\nfunc f(callback func(x int, y int) error, b int) {\n}\n",
			want: "package p\n\nfunc f(\n\tcallback func(x int, y int) error,\n\tb int,\n) {\n}\n",
		},
		{
			name: "parameter sharing the line a multi-line parameter ends on",
			code: "package p\n\nfunc f(\n\tcallback func(\n\t\tx int,\n\t) error, b int,\n) {\n}\n",
			want: "package p\n\nfunc f(\n\tcallback func(\n\t\tx int,\n\t) error,\n\tb int,\n) {\n}\n",
		},
		{
			name: "comments inside the list are kept",
			code: "package p\n\nfunc f(a int /* first, really */, b int /* last */) {\n}\n",
			want: "package p\n\nfunc f(\n\ta int, /* first, really */\n\tb int, /* last */\n) {\n}\n",
		},
	}
	for _, tc := range cases {
		assertFixed(
			t,
			tc.name,
			tc.code,
			tc.want,
		)
	}
}

func assertFixed(
	t *testing.T,
	name string,
	code string,
	want string,
) {
	t.Helper()
	got := fixed(
		t,
		code,
	)
	if got != want {
		t.Errorf(
			"%s: fixed =\n%s\nwant\n%s",
			name,
			got,
			want,
		)
	}

	again := fixed(
		t,
		got,
	)
	if again != got {
		t.Errorf(
			"%s: second fix changed the code again:\n%s",
			name,
			again,
		)
	}

	assertFindings(
		t,
		name,
		paramperline.Paramperline([]source.File{parsedContent(
			t,
			got,
		)}),
		nil,
	)
}

func fixed(
	t *testing.T,
	code string,
) (result string) {
	t.Helper()
	file := parsedContent(
		t,
		code,
	)
	edits := paramperline.Edits(file)
	if len(edits) == 0 {
		return code
	}

	formatted, err := format.Source(source.Splice(
		file.Content,
		edits,
	))
	if err != nil {
		t.Fatalf(
			"format fixed code: %v",
			err,
		)
	}

	return string(formatted)
}

func parsedContent(
	t *testing.T,
	code string,
) (file source.File) {
	t.Helper()
	file, err := source.Parse(
		"",
		"p.go",
		[]byte(code),
	)
	if err != nil {
		t.Fatalf(
			"parse fixture: %v",
			err,
		)
	}

	return file
}
