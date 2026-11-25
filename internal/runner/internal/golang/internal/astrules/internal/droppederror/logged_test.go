package droppederror_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/droppederror"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestDroppederrorLogged(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{name: "printing an error and carrying on is reported", code: "package p\n\nfunc f() {\n\terr := os.Remove(path)\n\tfmt.Println(err)\n}\n", want: []string{"R26 p.go:5"}},
		{name: "logging inside a nil check that falls through is reported", code: "package p\n\nfunc f() {\n\tif err := run(); err != nil {\n\t\tlog.Printf(\"run: %v\", err)\n\t}\n\n\tnext()\n}\n", want: []string{"R26 p.go:5"}},
		{name: "structured logging of an error message is reported", code: "package p\n\nfunc f() {\n\tif err := run(); err != nil {\n\t\tslog.Error(\"run\", \"err\", err.Error())\n\t}\n}\n", want: []string{"R26 p.go:5"}},
		{name: "logging nested in a structured attribute is reported", code: "package p\n\nfunc f() {\n\tif err := run(); err != nil {\n\t\tslog.Warn(\"run\", slog.Any(\"err\", err))\n\t}\n}\n", want: []string{"R26 p.go:5"}},
		{name: "builtin println of an error inside a loop that continues is reported", code: "package p\n\nfunc f() {\n\tfor _, item := range items {\n\t\tif err := handle(item); err != nil {\n\t\t\tprintln(err)\n\t\t\tcontinue\n\t\t}\n\t}\n}\n", want: []string{"R26 p.go:6"}},
		{name: "logging a recovered panic passes", code: "package p\n\nfunc f() {\n\tdefer func() {\n\t\tvar err = recover()\n\t\tif err != nil {\n\t\t\tlog.Println(err)\n\t\t}\n\t}()\n}\n", want: nil},
		{name: "logging then returning the error passes", code: "package p\n\nfunc f() error {\n\tif err := run(); err != nil {\n\t\tlog.Println(err)\n\t\treturn err\n\t}\n\n\treturn nil\n}\n", want: nil},
		{name: "logging then exiting passes", code: "package p\n\nfunc main() {\n\tif err := run(); err != nil {\n\t\tfmt.Println(err)\n\t\tos.Exit(1)\n\t}\n}\n", want: nil},
		{name: "fatal logging passes", code: "package p\n\nfunc main() {\n\tif err := run(); err != nil {\n\t\tlog.Fatal(err)\n\t}\n}\n", want: nil},
		{name: "logging then panicking passes", code: "package p\n\nfunc f() {\n\tif err := run(); err != nil {\n\t\tlog.Println(err)\n\t\tpanic(err)\n\t}\n}\n", want: nil},
		{name: "logging an error that is returned after the block passes", code: "package p\n\nfunc f() (err error) {\n\terr = run()\n\tif err != nil {\n\t\tlog.Println(err)\n\t}\n\n\treturn err\n}\n", want: nil},
		{name: "a later use after rebinding does not carry the logged error", code: "package p\n\nfunc f() (err error) {\n\terr = run()\n\tif err != nil {\n\t\tlog.Println(err)\n\t}\n\n\terr = next()\n\treturn err\n}\n", want: []string{"R26 p.go:6"}},
		{name: "logging a value that is not an error passes", code: "package p\n\nfunc f() {\n\tfmt.Println(count)\n\tlog.Printf(\"%d\", total)\n}\n", want: nil},
		{name: "writing an error to a stream is left alone", code: "package p\n\nfunc f() {\n\tfmt.Fprintln(os.Stderr, err)\n}\n", want: nil},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			droppederror.Droppederror([]source.File{parsed(
				t,
				tc.code,
			)}),
			tc.want,
		)
	}
}
