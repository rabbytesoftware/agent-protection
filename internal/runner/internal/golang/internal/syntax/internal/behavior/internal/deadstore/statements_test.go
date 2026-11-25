package deadstore_test

import "testing"

func TestDeadstoreAssignments(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "defined error never read is reported",
			code: "package p\n\nfunc f(\n\ts string,\n) (n int) {\n\tn, err := strconv.Atoi(s)\n\treturn n\n}\n",
			want: []string{"p.go:6"},
		},
		{
			name: "checked error passes",
			code: "package p\n\nfunc f(\n\ts string,\n) (n int, err error) {\n\tn, err = strconv.Atoi(s)\n\tif err != nil {\n\t\treturn 0, err\n\t}\n\n\treturn n, nil\n}\n",
			want: nil,
		},
		{
			name: "overwritten error is reported",
			code: "package p\n\nfunc f() (err error) {\n\terr = a()\n\terr = b()\n\treturn err\n}\n",
			want: []string{"p.go:4"},
		},
		{
			name: "nil reset passes",
			code: "package p\n\nfunc f() (err error) {\n\terr = nil\n\terr = a()\n\treturn err\n}\n",
			want: nil,
		},
		{
			name: "store not from a call passes",
			code: "package p\n\nfunc f() (err error) {\n\terr = errFixed\n\treturn nil\n}\n",
			want: nil,
		},
		{
			name: "var declaration with a call value is reported when unread",
			code: "package p\n\nfunc f() (n int) {\n\tvar err = a()\n\treturn 0\n}\n",
			want: []string{"p.go:4"},
		},
		{
			name: "var declaration read later passes",
			code: "package p\n\nfunc f() (err error) {\n\tvar runErr error = a()\n\treturn runErr\n}\n",
			want: nil,
		},
		{
			name: "field stores are ignored",
			code: "package p\n\nfunc f(\n\ts *state,\n) (n int) {\n\ts.err = a()\n\treturn 0\n}\n",
			want: nil,
		},
		{
			name: "selector named err is not a read",
			code: "package p\n\nfunc f() (err error) {\n\terr = a()\n\tlog(s.err)\n\treturn nil\n}\n",
			want: []string{"p.go:4"},
		},
		{
			name: "compound assignment reads",
			code: "package p\n\nfunc f() (n int) {\n\tcountErr := a()\n\tcountErr += 1\n\treturn countErr\n}\n",
			want: nil,
		},
		{
			name: "multiple values from separate calls",
			code: "package p\n\nfunc f() (err error) {\n\tfirstErr, secondErr := a(), b()\n\treturn firstErr\n}\n",
			want: []string{"p.go:4"},
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
