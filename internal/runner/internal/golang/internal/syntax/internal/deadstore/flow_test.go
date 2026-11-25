package deadstore_test

import "testing"

func TestDeadstoreScopes(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "shadowed error in a block does not hide the outer read",
			code: "package p\n\nfunc f() (err error) {\n\terr = a()\n\t{\n\t\terr := b()\n\t\tlog(err)\n\t}\n\n\treturn err\n}\n",
			want: nil,
		},
		{
			name: "unread shadowed error is reported",
			code: "package p\n\nfunc f() (err error) {\n\terr = a()\n\tif true {\n\t\terr := b()\n\t\tlog(1)\n\t}\n\n\treturn err\n}\n",
			want: []string{"p.go:6"},
		},
		{
			name: "deferred closure reading the result passes",
			code: "package p\n\nfunc f() (err error) {\n\tdefer func() {\n\t\tlog(err)\n\t}()\n\terr = a()\n\treturn nil\n}\n",
			want: nil,
		},
		{
			name: "address-taken error passes",
			code: "package p\n\nfunc f() (err error) {\n\tdefer finish(&err)\n\terr = a()\n\treturn nil\n}\n",
			want: nil,
		},
		{
			name: "naked return reads the named result",
			code: "package p\n\nfunc f() (err error) {\n\terr = a()\n\treturn\n}\n",
			want: nil,
		},
		{
			name: "Err suffixed names are tracked",
			code: "package p\n\nfunc f() (err error) {\n\tcloseErr := a()\n\treturn nil\n}\n",
			want: []string{"p.go:4"},
		},
		{
			name: "captured variable in a closure passes",
			code: "package p\n\nfunc f() (err error) {\n\trun := func() {\n\t\terr = a()\n\t}\n\trun()\n\treturn err\n}\n",
			want: nil,
		},
		{
			name: "parameter overwritten without read is reported",
			code: "package p\n\nfunc f(\n\terr error,\n) (out error) {\n\terr = a()\n\treturn nil\n}\n",
			want: []string{"p.go:6"},
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
