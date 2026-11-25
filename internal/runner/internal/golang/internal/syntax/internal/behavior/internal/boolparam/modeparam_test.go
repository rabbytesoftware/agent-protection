package boolparam_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/behavior/internal/boolparam"
)

func TestBoolparamModeParameter(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "string parameter only compared against one literal is reported",
			code: "package p\n\nfunc F(\n\tvalue int,\n\tsign string,\n) (scaled int) {\n\tif sign == \"-\" {\n\t\treturn -value\n\t}\n\n\treturn value\n}\n",
			want: []string{"p.go:3"},
		},
		{
			name: "int parameter compared against one constant in a condition chain is reported",
			code: "package p\n\nconst negative = 1\n\nfunc F(\n\tvalue int,\n\tmode int,\n) (scaled int) {\n\tif value > 0 && !(mode != negative) {\n\t\treturn -value\n\t}\n\n\tif mode == negative {\n\t\treturn 0\n\t}\n\n\treturn value\n}\n",
			want: []string{"p.go:5"},
		},
		{
			name: "switch on the parameter with one literal case is reported",
			code: "package p\n\nfunc F(\n\tvalue int,\n\tsign string,\n) (scaled int) {\n\tswitch sign {\n\tcase \"-\":\n\t\treturn -value\n\t}\n\n\treturn value\n}\n",
			want: []string{"p.go:3"},
		},
		{
			name: "tagless switch comparing the parameter to one literal is reported",
			code: "package p\n\nfunc F(\n\tvalue int,\n\tsign string,\n) (scaled int) {\n\tswitch {\n\tcase sign == \"-\" || value < 0:\n\t\treturn -value\n\t}\n\n\treturn value\n}\n",
			want: []string{"p.go:3"},
		},
		{
			name: "comparisons against two literals, other uses, typed enums and single-parameter classifiers pass",
			code: "package p\n\ntype Sign string\n\nfunc f(\n\tvalue int,\n\tsign string,\n) (scaled int) {\n\tif sign == \"-\" {\n\t\treturn -value\n\t}\n\n\tif sign == \"+\" {\n\t\treturn value\n\t}\n\n\treturn 0\n}\n\nfunc g(\n\tvalue int,\n\tsign string,\n) (scaled int) {\n\tif sign == \"-\" {\n\t\treturn -value\n\t}\n\n\treturn len(sign)\n}\n\nfunc h(\n\tvalue int,\n\tsign Sign,\n) (scaled int) {\n\tif sign == \"-\" {\n\t\treturn -value\n\t}\n\n\treturn value\n}\n\nfunc role(\n\tbase string,\n) (main bool) {\n\tif base == \"main.go\" {\n\t\treturn true\n\t}\n\n\treturn false\n}\n",
			want: nil,
		},
		{
			name: "a selector constant and a bodiless declaration pass",
			code: "package p\n\nfunc f(\n\tvalue int,\n\tsign string,\n) (scaled int) {\n\tif sign == pkg.Minus {\n\t\treturn -value\n\t}\n\n\treturn value\n}\n\nfunc g(\n\tvalue int,\n\tsign string,\n) (scaled int)\n",
			want: nil,
		},
		{
			name: "an unexported function a caller hands a literal mode is reported",
			code: "package p\n\nfunc G(\n\tvalue int,\n) (scaled int) {\n\treturn f(value, \"\")\n}\n\nfunc f(\n\tvalue int,\n\tsign string,\n) (scaled int) {\n\tif sign == \"-\" {\n\t\treturn -value\n\t}\n\n\treturn value\n}\n",
			want: []string{"p.go:9"},
		},
		{
			name: "an unexported filter whose callers pass data, never a literal, passes",
			code: "package p\n\nfunc G(\n\tvalue int,\n\tpath string,\n) (scaled int) {\n\treturn f(value, path)\n}\n\nfunc f(\n\tvalue int,\n\tpath string,\n) (scaled int) {\n\tif path != \"testing\" {\n\t\treturn value\n\t}\n\n\treturn 0\n}\n",
			want: nil,
		},
		{
			name: "a returned comparison and a comparison against another variable pass",
			code: "package p\n\nfunc f(\n\tvalue int,\n\tsign string,\n) (negative bool) {\n\treturn sign == \"-\"\n}\n\nfunc g(\n\twant string,\n\tsign string,\n) (scaled int) {\n\tif sign == want {\n\t\treturn 1\n\t}\n\n\treturn 0\n}\n",
			want: nil,
		},
	}
	for _, tc := range cases {
		assertLocations(
			t,
			tc.name,
			boolparam.Boolparam([]source.File{parsed(
				t,
				tc.code,
			)}),
			tc.want,
		)
	}
}
