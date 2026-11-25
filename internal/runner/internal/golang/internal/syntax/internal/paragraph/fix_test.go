package paragraph_test

import (
	"go/format"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/paragraph"
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
			name: "blank line goes after a closed block",
			code: "package p\n\nfunc f() {\n\tif a {\n\t\tg()\n\t}\n\tg()\n}\n",
			want: "package p\n\nfunc f() {\n\tif a {\n\t\tg()\n\t}\n\n\tg()\n}\n",
		},
		{
			name: "blank line goes after the trailing comment of a closed block",
			code: "package p\n\nfunc f() {\n\tfor range a {\n\t} // done\n\t// next\n\tg()\n}\n",
			want: "package p\n\nfunc f() {\n\tfor range a {\n\t} // done\n\n\t// next\n\tg()\n}\n",
		},
		{
			name: "block closing its enclosing block needs no blank line",
			code: "package p\n\nfunc f() {\n\tg()\n\tif a {\n\t}\n}\n",
			want: "package p\n\nfunc f() {\n\tg()\n\tif a {\n\t}\n}\n",
		},
		{
			name: "case and comm clause bodies are fixed",
			code: "package p\n\nfunc f() {\n\tswitch {\n\tcase a:\n\t\tif b {\n\t\t}\n\t\tg()\n\t}\n\n\tselect {\n\tcase <-c:\n\t\tswitch {\n\t\t}\n\t\tg()\n\t}\n}\n",
			want: "package p\n\nfunc f() {\n\tswitch {\n\tcase a:\n\t\tif b {\n\t\t}\n\n\t\tg()\n\t}\n\n\tselect {\n\tcase <-c:\n\t\tswitch {\n\t\t}\n\n\t\tg()\n\t}\n}\n",
		},
		{
			name: "check is reattached to its assignment",
			code: "package p\n\nfunc f() {\n\terr := g()\n\n\n\tif err != nil {\n\t}\n}\n",
			want: "package p\n\nfunc f() {\n\terr := g()\n\tif err != nil {\n\t}\n}\n",
		},
		{
			name: "comment between an assignment and its check stays",
			code: "package p\n\nfunc f() {\n\terr := g()\n\n\t// why\n\n\tif err != nil {\n\t}\n}\n",
			want: "package p\n\nfunc f() {\n\terr := g()\n\t// why\n\tif err != nil {\n\t}\n}\n",
		},
		{
			name: "blank before an unrelated if stays",
			code: "package p\n\nfunc f() {\n\tx := g()\n\n\tif a {\n\t}\n\n\t_ = x\n}\n",
			want: "package p\n\nfunc f() {\n\tx := g()\n\n\tif a {\n\t}\n\n\t_ = x\n}\n",
		},
		{
			name: "leading and trailing blank lines are removed",
			code: "package p\n\nfunc f() {\n\n\tg()\n\n}\n",
			want: "package p\n\nfunc f() {\n\tg()\n}\n",
		},
		{
			name: "blank lines inside block comments stay",
			code: "package p\n\nfunc f() {\n\n\t/* one\n\n\ttwo */\n\tg()\n}\n",
			want: "package p\n\nfunc f() {\n\t/* one\n\n\ttwo */\n\tg()\n}\n",
		},
		{
			name: "blank line inside a multi-line string stays",
			code: "package p\n\nfunc f() {\n\tif a {\n\t}\n\tg(`one\n\ntwo`)\n}\n",
			want: "package p\n\nfunc f() {\n\tif a {\n\t}\n\n\tg(`one\n\ntwo`)\n}\n",
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
		paragraph.Paragraph([]source.File{parsedContent(
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
	edits := paragraph.Edits(file)
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
