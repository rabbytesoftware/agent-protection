package returnperline_test

import (
	"go/format"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/style/internal/returnperline"
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
			name: "bare and single-value returns are left alone",
			code: "package p\n\nfunc f() {\n\treturn\n}\n\nfunc g() int {\n\treturn 1\n}\n",
			want: "package p\n\nfunc f() {\n\treturn\n}\n\nfunc g() int {\n\treturn 1\n}\n",
		},
		{
			name: "compliant return is left alone",
			code: "package p\n\nfunc f() (int, error) {\n\treturn 1,\n\t\tnil\n}\n",
			want: "package p\n\nfunc f() (int, error) {\n\treturn 1,\n\t\tnil\n}\n",
		},
		{
			name: "two values get a line each",
			code: "package p\n\nfunc f() (int, error) {\n\treturn 1, nil\n}\n",
			want: "package p\n\nfunc f() (int, error) {\n\treturn 1,\n\t\tnil\n}\n",
		},
		{
			name: "three values get a line each",
			code: "package p\n\nfunc f() (int, int, error) {\n\treturn 1, 2, nil\n}\n",
			want: "package p\n\nfunc f() (int, int, error) {\n\treturn 1,\n\t\t2,\n\t\tnil\n}\n",
		},
		{
			name: "only the collapsed pair is split",
			code: "package p\n\nfunc f() (int, int, error) {\n\treturn 1,\n\t\t2, nil\n}\n",
			want: "package p\n\nfunc f() (int, int, error) {\n\treturn 1,\n\t\t2,\n\t\tnil\n}\n",
		},
		{
			name: "value on the closing line of a multi-line value",
			code: "package p\n\nfunc f() (error, int) {\n\treturn g(\n\t\t1,\n\t), 2\n}\n",
			want: "package p\n\nfunc f() (error, int) {\n\treturn g(\n\t\t\t1,\n\t\t),\n\t\t2\n}\n",
		},
		{
			name: "returns inside func literals",
			code: "package p\n\nvar f = func() (int, error) {\n\treturn 1, nil\n}\n",
			want: "package p\n\nvar f = func() (int, error) {\n\treturn 1,\n\t\tnil\n}\n",
		},
		{
			name: "comments between values are kept",
			code: "package p\n\nfunc f() (int, error) {\n\treturn 1 /* one */, nil\n}\n",
			want: "package p\n\nfunc f() (int, error) {\n\treturn 1, /* one */\n\t\tnil\n}\n",
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
		returnperline.Returnperline([]source.File{parsedContent(
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
	edits := returnperline.Edits(file)
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
