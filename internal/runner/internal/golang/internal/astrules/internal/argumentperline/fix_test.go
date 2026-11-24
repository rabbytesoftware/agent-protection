package argumentperline_test

import (
	"go/format"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/argumentperline"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
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
			name: "zero and one argument stay inline",
			code: "package p\n\nfunc f() {\n\tg()\n\th(1)\n}\n",
			want: "package p\n\nfunc f() {\n\tg()\n\th(1)\n}\n",
		},
		{
			name: "compliant call is left alone",
			code: "package p\n\nfunc f() {\n\tg(\n\t\t1,\n\t\t2,\n\t)\n}\n",
			want: "package p\n\nfunc f() {\n\tg(\n\t\t1,\n\t\t2,\n\t)\n}\n",
		},
		{
			name: "inline arguments get a line each and a trailing comma",
			code: "package p\n\nfunc f() {\n\tg(1, 2)\n}\n",
			want: "package p\n\nfunc f() {\n\tg(\n\t\t1,\n\t\t2,\n\t)\n}\n",
		},
		{
			name: "nested calls are both broken",
			code: "package p\n\nfunc f() {\n\tg(h(1, 2), 3)\n}\n",
			want: "package p\n\nfunc f() {\n\tg(\n\t\th(\n\t\t\t1,\n\t\t\t2,\n\t\t),\n\t\t3,\n\t)\n}\n",
		},
		{
			name: "nested single-argument call stays inline",
			code: "package p\n\nfunc f() {\n\tg(h(1), 3)\n}\n",
			want: "package p\n\nfunc f() {\n\tg(\n\t\th(1),\n\t\t3,\n\t)\n}\n",
		},
		{
			name: "variadic spread keeps its ellipsis before the trailing comma",
			code: "package p\n\nfunc f() {\n\tx = append(x, y...)\n}\n",
			want: "package p\n\nfunc f() {\n\tx = append(\n\t\tx,\n\t\ty...,\n\t)\n}\n",
		},
		{
			name: "func literal argument",
			code: "package p\n\nfunc f() {\n\tg(func() {\n\t\th()\n\t}, 2)\n}\n",
			want: "package p\n\nfunc f() {\n\tg(\n\t\tfunc() {\n\t\t\th()\n\t\t},\n\t\t2,\n\t)\n}\n",
		},
		{
			name: "generic call keeps its type arguments inline",
			code: "package p\n\nfunc f() {\n\tg[int, string](1, \"a\")\n}\n",
			want: "package p\n\nfunc f() {\n\tg[int, string](\n\t\t1,\n\t\t\"a\",\n\t)\n}\n",
		},
		{
			name: "multi-line raw string is never split",
			code: "package p\n\nfunc f() {\n\tg(`one\ntwo`, 2)\n}\n",
			want: "package p\n\nfunc f() {\n\tg(\n\t\t`one\ntwo`,\n\t\t2,\n\t)\n}\n",
		},
		{
			name: "comments inside the list are kept",
			code: "package p\n\nfunc f() {\n\tg(1 /* one, really */, 2 /* two */)\n}\n",
			want: "package p\n\nfunc f() {\n\tg(\n\t\t1, /* one, really */\n\t\t2, /* two */\n\t)\n}\n",
		},
		{
			name: "existing trailing comma is kept",
			code: "package p\n\nfunc f() {\n\tg(\n\t\t1,\n\t\t2,)\n}\n",
			want: "package p\n\nfunc f() {\n\tg(\n\t\t1,\n\t\t2,\n\t)\n}\n",
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

	assertLocations(
		t,
		name,
		argumentperline.Argumentperline([]source.File{parsedContent(
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
	edits := argumentperline.Edits(file)
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
