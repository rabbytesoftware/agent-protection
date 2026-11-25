package conditionperline_test

import (
	"go/format"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/style/internal/conditionperline"
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
			name: "single condition and comparison are left alone",
			code: "package p\n\nfunc f() {\n\tif a {\n\t}\n\n\tif a == b {\n\t}\n}\n",
			want: "package p\n\nfunc f() {\n\tif a {\n\t}\n\n\tif a == b {\n\t}\n}\n",
		},
		{
			name: "compliant condition is left alone",
			code: "package p\n\nfunc f() {\n\tif a &&\n\t\tb {\n\t}\n}\n",
			want: "package p\n\nfunc f() {\n\tif a &&\n\t\tb {\n\t}\n}\n",
		},
		{
			name: "joined conditions get a line each",
			code: "package p\n\nfunc f() {\n\tif a && b || c {\n\t}\n}\n",
			want: "package p\n\nfunc f() {\n\tif a &&\n\t\tb ||\n\t\tc {\n\t}\n}\n",
		},
		{
			name: "nested group is split too",
			code: "package p\n\nfunc f() {\n\tif a && (b || c) {\n\t}\n}\n",
			want: "package p\n\nfunc f() {\n\tif a &&\n\t\t(b ||\n\t\t\tc) {\n\t}\n}\n",
		},
		{
			name: "init statement and negation",
			code: "package p\n\nfunc f() {\n\tif ok := g(); !ok || h() {\n\t}\n}\n",
			want: "package p\n\nfunc f() {\n\tif ok := g(); !ok ||\n\t\th() {\n\t}\n}\n",
		},
		{
			name: "joins outside an if are left alone",
			code: "package p\n\nfunc f() {\n\tok := a && b\n\t_ = ok\n}\n",
			want: "package p\n\nfunc f() {\n\tok := a && b\n\t_ = ok\n}\n",
		},
		{
			name: "if nested in a condition closure is split once",
			code: "package p\n\nfunc f() {\n\tif g(func() {\n\t\tif a && b {\n\t\t}\n\t}) {\n\t}\n}\n",
			want: "package p\n\nfunc f() {\n\tif g(func() {\n\t\tif a &&\n\t\t\tb {\n\t\t}\n\t}) {\n\t}\n}\n",
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
		conditionperline.Conditionperline([]source.File{parsedContent(
			t,
			got,
		)}),
		nil,
	)
}

func fixed(
	t *testing.T,
	code string,
) (rewritten string) {
	t.Helper()
	file := parsedContent(
		t,
		code,
	)
	edits := conditionperline.Edits(file)
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
