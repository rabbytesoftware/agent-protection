package deadstore_test

import "testing"

func TestDeadstoreLoops(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "error overwritten in a loop is reported",
			code: "package p\n\nfunc f(\n\tpaths []string,\n) (err error) {\n\tfor _, path := range paths {\n\t\terr = a(path)\n\t}\n\n\treturn err\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "error checked inside the loop passes",
			code: "package p\n\nfunc f(\n\tpaths []string,\n) (err error) {\n\tfor _, path := range paths {\n\t\terr = a(path)\n\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n\t}\n\n\treturn nil\n}\n",
			want: nil,
		},
		{
			name: "errors.Join accumulation passes",
			code: "package p\n\nfunc f(\n\tpaths []string,\n) (err error) {\n\tfor _, path := range paths {\n\t\terr = errors.Join(\n\t\t\terr,\n\t\t\ta(path),\n\t\t)\n\t}\n\n\treturn err\n}\n",
			want: nil,
		},
		{
			name: "loop condition reading the error passes",
			code: "package p\n\nfunc f() (err error) {\n\tfor err == nil {\n\t\terr = a()\n\t}\n\n\treturn err\n}\n",
			want: nil,
		},
		{
			name: "loop that breaks after the store passes",
			code: "package p\n\nfunc f() (err error) {\n\tfor {\n\t\terr = a()\n\t\tbreak\n\t}\n\n\treturn err\n}\n",
			want: nil,
		},
		{
			name: "read after the loop through a break passes",
			code: "package p\n\nfunc f() (err error) {\n\tfor i := 0; i < 3; i++ {\n\t\terr = a()\n\t\tif ok() {\n\t\t\tbreak\n\t\t}\n\t}\n\n\treturn err\n}\n",
			want: []string{"p.go:5"},
		},
		{
			name: "continue before the read keeps the loop lossy",
			code: "package p\n\nfunc f() (err error) {\n\tfor i := 0; i < 3; i++ {\n\t\terr = a()\n\t\tif ok() {\n\t\t\tcontinue\n\t\t}\n\n\t\tlog(err)\n\t}\n\n\treturn nil\n}\n",
			want: nil,
		},
		{
			name: "read at the top of the next iteration passes",
			code: "package p\n\nfunc f() (err error) {\n\tfor range 3 {\n\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n\n\t\terr = a()\n\t}\n\n\treturn err\n}\n",
			want: nil,
		},
		{
			name: "range variable error is tracked",
			code: "package p\n\nfunc f(\n\terrs []error,\n) (n int) {\n\tfor _, err := range errs {\n\t\tlog(err)\n\t}\n\n\treturn 0\n}\n",
			want: nil,
		},
		{
			name: "infinite loop with return keeps the store",
			code: "package p\n\nfunc f() (err error) {\n\tfor {\n\t\terr = a()\n\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n\t}\n}\n",
			want: nil,
		},
		{
			name: "nested loop store is checked against the inner loop",
			code: "package p\n\nfunc f() (err error) {\n\tfor range 2 {\n\t\tfor range 2 {\n\t\t\terr = a()\n\t\t}\n\n\t\tif err != nil {\n\t\t\treturn err\n\t\t}\n\t}\n\n\treturn nil\n}\n",
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
