package boolparam_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/behavior/internal/boolparam"
)

func TestBoolparamCallSite(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "a literal declared with var and passed in another file's call is reported",
			code: "package p\n\nfunc f(\n\tvalue int,\n) (scaled int) {\n\tvar request = operand{value: value}\n\treturn g(request)\n}\n",
			want: []string{"p/g.go:3"},
		},
		{
			name: "a literal passed at another position passes",
			code: "package p\n\nfunc f(\n\tvalue int,\n) (scaled int) {\n\treturn g(load(value), operand{})\n}\n",
			want: nil,
		},
	}
	for _, tc := range cases {
		assertLocations(
			t,
			tc.name,
			boolparam.Boolparam([]source.File{
				parsedAt(
					t,
					"p/operand.go",
					operandType,
				),
				parsedAt(
					t,
					"p/g.go",
					"package p\n\nfunc g(\n\trequest operand,\n\tother operand,\n) (scaled int) {\n\tif request.negate {\n\t\treturn -request.value\n\t}\n\n\treturn other.value\n}\n",
				),
				parsedAt(
					t,
					"p/f.go",
					tc.code,
				),
			}),
			tc.want,
		)
	}
}
