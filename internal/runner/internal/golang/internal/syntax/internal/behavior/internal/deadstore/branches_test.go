package deadstore_test

import "testing"

func TestDeadstoreSwitchesAndSelects(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "switch case reading the error passes",
			code: "package p\n\nfunc f(\n\tx int,\n) (err error) {\n\terr = a()\n\tswitch x {\n\tcase 1:\n\t\treturn err\n\tdefault:\n\t\treturn nil\n\t}\n}\n",
			want: nil,
		},
		{
			name: "switch overwriting in every case is reported",
			code: "package p\n\nfunc f(\n\tx int,\n) (err error) {\n\terr = a()\n\tswitch x {\n\tcase 1:\n\t\terr = b()\n\tdefault:\n\t\terr = c()\n\t}\n\n\treturn err\n}\n",
			want: []string{"p.go:6"},
		},
		{
			name: "switch without default keeps the store live",
			code: "package p\n\nfunc f(\n\tx int,\n) (err error) {\n\terr = a()\n\tswitch x {\n\tcase 1:\n\t\terr = b()\n\t}\n\n\treturn err\n}\n",
			want: nil,
		},
		{
			name: "fallthrough carries the store",
			code: "package p\n\nfunc f(\n\tx int,\n) (err error) {\n\tswitch x {\n\tcase 1:\n\t\terr = a()\n\t\tfallthrough\n\tcase 2:\n\t\treturn err\n\t}\n\n\treturn nil\n}\n",
			want: nil,
		},
		{
			name: "type switch guard is its own variable",
			code: "package p\n\nfunc f(\n\tv any,\n) (err error) {\n\terr = a()\n\tswitch err := v.(type) {\n\tcase error:\n\t\terr = b()\n\t\tlog(err)\n\t}\n\n\treturn err\n}\n",
			want: nil,
		},
		{
			name: "select case reading the error passes",
			code: "package p\n\nfunc f(\n\tc chan error,\n) (err error) {\n\terr = a()\n\tselect {\n\tcase <-c:\n\t\treturn err\n\tdefault:\n\t}\n\n\treturn nil\n}\n",
			want: nil,
		},
		{
			name: "select receiving into the error is checked",
			code: "package p\n\nfunc f(\n\tc chan error,\n) (err error) {\n\tselect {\n\tcase err = <-c:\n\t\treturn err\n\t}\n}\n",
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
