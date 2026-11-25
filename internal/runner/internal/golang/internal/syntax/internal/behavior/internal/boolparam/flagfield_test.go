package boolparam_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/behavior/internal/boolparam"
)

const operandType = "package p\n\ntype operand struct {\n\tvalue  int\n\tnegate bool\n}\n"

func TestBoolparamFlagField(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "branching on a bool field of a struct built at the call is reported",
			code: "package p\n\nfunc f(\n\tvalue int,\n) (scaled int) {\n\treturn g(operand{value: value})\n}\n\nfunc g(\n\trequest operand,\n) (scaled int) {\n\tif request.negate {\n\t\treturn -request.value\n\t}\n\n\treturn request.value\n}\n",
			want: []string{"p/square.go:9"},
		},
		{
			name: "a pointer struct assigned from a literal before the call is reported",
			code: "package p\n\nfunc f(\n\tvalue int,\n) (scaled int) {\n\trequest := &operand{value: value}\n\treturn s.g(\n\t\t1,\n\t\trequest,\n\t)\n}\n\nfunc (s *square) g(\n\tfactor int,\n\trequest *operand,\n) (scaled int) {\n\tswitch {\n\tcase request.negate && factor > 0:\n\t\treturn -request.value\n\t}\n\n\treturn request.value\n}\n",
			want: []string{"p/square.go:13"},
		},
		{
			name: "a struct produced elsewhere and carried around passes",
			code: "package p\n\nfunc f(\n\tvalue int,\n) (scaled int) {\n\treturn g(load(value))\n}\n\nfunc g(\n\trequest operand,\n) (scaled int) {\n\tif request.negate {\n\t\treturn -request.value\n\t}\n\n\treturn request.value\n}\n",
			want: nil,
		},
		{
			name: "a non-literal assignment and a bodiless declaration pass",
			code: "package p\n\nfunc f(\n\tvalue int,\n) (scaled int) {\n\trequest := load(value)\n\treturn g(request)\n}\n\nfunc g(\n\trequest operand,\n) (scaled int)\n",
			want: nil,
		},
		{
			name: "reading a bool field without branching on it passes",
			code: "package p\n\nfunc f(\n\tvalue int,\n) (negated bool) {\n\treturn g(operand{value: value})\n}\n\nfunc g(\n\trequest operand,\n) (negated bool) {\n\tif request.value > 0 {\n\t\treturn request.negate\n\t}\n\n\treturn false\n}\n",
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
					"p/square.go",
					tc.code,
				),
			}),
			tc.want,
		)
	}
}
