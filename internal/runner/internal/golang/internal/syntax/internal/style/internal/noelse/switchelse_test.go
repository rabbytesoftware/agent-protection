package noelse_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/style/internal/noelse"
)

func TestSwitchElse(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "tagless switch with one case and a default is reported",
			code: "package p\n\nfunc f(\n\tv int,\n) (sign int) {\n\tswitch {\n\tcase v < 0:\n\t\treturn -1\n\tdefault:\n\t\treturn 1\n\t}\n}\n",
			want: []string{"p.go:6"},
		},
		{
			name: "switch true with one case and a default is reported",
			code: "package p\n\nfunc f(\n\tv int,\n) (sign int) {\n\tswitch true {\n\tcase v < 0, v > 9:\n\t\treturn -1\n\tdefault:\n\t\treturn 1\n\t}\n}\n",
			want: []string{"p.go:6"},
		},
		{
			name: "switch on a bool against true is reported",
			code: "package p\n\nfunc f(\n\tv int,\n) (sign int) {\n\tswitch v < 0 {\n\tcase true:\n\t\treturn -1\n\t}\n\treturn 1\n}\n",
			want: []string{"p.go:6"},
		},
		{
			name: "switch on a bool against true and false is reported once",
			code: "package p\n\nfunc f(\n\tnegative bool,\n) (sign int) {\n\tswitch negative {\n\tcase true:\n\t\treturn -1\n\tcase false:\n\t\treturn 1\n\t}\n\treturn 0\n}\n",
			want: []string{"p.go:6"},
		},
		{
			name: "tagless switch with two cases and a default passes",
			code: "package p\n\nfunc f(\n\tv int,\n) (sign int) {\n\tswitch {\n\tcase v < 0:\n\t\treturn -1\n\tcase v > 0:\n\t\treturn 1\n\tdefault:\n\t\treturn 0\n\t}\n}\n",
			want: nil,
		},
		{
			name: "tagless switch with one case and no default passes",
			code: "package p\n\nfunc f(\n\tv int,\n) (sign int) {\n\tswitch {\n\tcase v < 0:\n\t\treturn -1\n\t}\n\treturn 1\n}\n",
			want: nil,
		},
		{
			name: "tagged switch with one case and a default passes",
			code: "package p\n\nfunc f(\n\tv int,\n) (sign int) {\n\tswitch v {\n\tcase 0:\n\t\treturn 0\n\tdefault:\n\t\treturn 1\n\t}\n}\n",
			want: nil,
		},
		{
			name: "type switch and a default-only switch pass",
			code: "package p\n\nfunc f(\n\tv any,\n) (sign int) {\n\tswitch v.(type) {\n\tcase int:\n\t\treturn 0\n\tdefault:\n\t\treturn 1\n\t}\n}\n\nfunc g(\n\tv bool,\n) {\n\tswitch v {\n\tdefault:\n\t}\n}\n",
			want: nil,
		},
	}
	for _, tc := range cases {
		assertLocations(
			t,
			tc.name,
			noelse.Noelse([]source.File{parsed(
				t,
				tc.code,
			)}),
			tc.want,
		)
	}
}
