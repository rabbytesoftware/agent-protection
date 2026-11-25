package droppederror_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/behavior/internal/droppederror"
)

func TestDroppederrorWrapping(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{name: "error formatted with %v is reported", code: "package p\n\nfunc f() error {\n\treturn fmt.Errorf(\"load %s: %v\", path, err)\n}\n", want: []string{"R26 p.go:4"}},
		{name: "error message formatted with %s is reported", code: "package p\n\nfunc f() error {\n\treturn fmt.Errorf(\"load: %s\", err.Error())\n}\n", want: []string{"R26 p.go:4"}},
		{name: "suffixed error formatted with %+v is reported", code: "package p\n\nfunc f() error {\n\treturn fmt.Errorf(\"close: %+v\", closeErr)\n}\n", want: []string{"R26 p.go:4"}},
		{name: "sentinel formatted with %v is reported", code: "package p\n\nfunc f() error {\n\treturn fmt.Errorf(\"%v: %s\", ErrMissing, name)\n}\n", want: []string{"R26 p.go:4"}},
		{name: "qualified sentinel formatted with %s is reported", code: "package p\n\nfunc f() error {\n\treturn fmt.Errorf(\"%s\", os.ErrNotExist)\n}\n", want: []string{"R26 p.go:4"}},
		{name: "error declared with the error type formatted with %v is reported", code: "package p\n\nfunc f(\n\tcause error,\n) error {\n\treturn fmt.Errorf(\"wrap: %v\", cause)\n}\n", want: []string{"R26 p.go:6"}},
		{name: "recovered panic value formatted with %v passes", code: "package p\n\nfunc f() (err error) {\n\tif recovered := recover(); recovered != nil {\n\t\terr = fmt.Errorf(\"panic: %v\", recovered)\n\t}\n\n\treturn err\n}\n", want: nil},
		{name: "error rebuilt from its message with errors.New is reported", code: "package p\n\nfunc f() error {\n\treturn errors.New(err.Error())\n}\n", want: []string{"R26 p.go:4"}},
		{name: "error rebuilt from a formatted message is reported once", code: "package p\n\nfunc f() error {\n\treturn errors.New(fmt.Sprintf(\"load: %v %s\", err, err.Error()))\n}\n", want: []string{"R26 p.go:4"}},
		{name: "error rebuilt by concatenating its message is reported", code: "package p\n\nfunc f(\n\tcause error,\n) error {\n\treturn errors.New(\"load: \" + cause.Error())\n}\n", want: []string{"R26 p.go:6"}},
		{name: "errors.New with a plain message passes", code: "package p\n\nfunc f() error {\n\treturn errors.New(\"load: \" + path)\n}\n", want: nil},
		{name: "error wrapped with %w passes", code: "package p\n\nfunc f() error {\n\treturn fmt.Errorf(\"load %s: %w\", path, err)\n}\n", want: nil},
		{name: "escaped percent before the wrap passes", code: "package p\n\nfunc f() error {\n\treturn fmt.Errorf(\"100%%: %w\", err)\n}\n", want: nil},
		{name: "non-error arguments with other verbs pass", code: "package p\n\nfunc f() error {\n\treturn fmt.Errorf(\"%w: %s %d\", err, stderr, ErrorCount)\n}\n", want: nil},
		{name: "explicit argument indexes are left alone", code: "package p\n\nfunc f() error {\n\treturn fmt.Errorf(\"%[1]v\", err)\n}\n", want: nil},
		{name: "non-literal format is left alone", code: "package p\n\nfunc f() error {\n\treturn fmt.Errorf(layout, err)\n}\n", want: nil},
		{name: "other formatting functions pass", code: "package p\n\nfunc f() string {\n\treturn fmt.Sprintf(\"%v\", err)\n}\n", want: nil},
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
