package deadstore_test

import "testing"

func TestDeadstoreBranchMerges(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "error read on one branch passes",
			code: "package p\n\nfunc f(\n\tx bool,\n) (err error) {\n\terr = a()\n\tif x {\n\t\treturn err\n\t}\n\n\treturn nil\n}\n",
			want: nil,
		},
		{
			name: "fallback overwrite on one branch passes",
			code: "package p\n\nfunc f(\n\tx bool,\n) (err error) {\n\terr = a()\n\tif x {\n\t\terr = b()\n\t}\n\n\treturn err\n}\n",
			want: nil,
		},
		{
			name: "overwrite on both branches is reported",
			code: "package p\n\nfunc f(\n\tx bool,\n) (err error) {\n\terr = a()\n\tif x {\n\t\terr = b()\n\t\treturn err\n\t}\n\n\terr = c()\n\treturn err\n}\n",
			want: []string{"p.go:6"},
		},
		{
			name: "else branch is followed",
			code: "package p\n\nfunc f(\n\tx bool,\n) (err error) {\n\tif x {\n\t\terr = a()\n\t} else {\n\t\terr = b()\n\t}\n\n\treturn err\n}\n",
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
