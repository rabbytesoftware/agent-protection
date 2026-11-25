package interfacesize_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/shape/internal/interfacesize"
)

func TestInterfacesizeModulePath(
	t *testing.T,
) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{
			name: "embeds imported from another package of the module are resolved",
			files: map[string]string{
				"internal/cube/reader.go": reader,
				"internal/cube/writer.go": writer,
				"internal/store/store.go": "package store\n\nimport \"example.com/m/internal/cube\"\n\ntype Store interface {\n\tcube.Reader\n\tcube.Writer\n}\n",
			},
			want: []string{"R28 internal/store/store.go:5 interface Store has 6 methods once embedded cube.Reader, cube.Writer are flattened (max 3): split it along what each consumer calls and have consumers depend on the small interfaces"},
		},
		{
			name: "explicit import alias is resolved",
			files: map[string]string{
				"internal/cube/reader.go": reader,
				"internal/store/store.go": "package store\n\nimport c \"example.com/m/internal/cube\"\n\ntype Store interface {\n\tc.Reader\n\tPut()\n}\n",
			},
			want: []string{"R28 internal/store/store.go:5 interface Store has 4 methods once embedded c.Reader are flattened (max 3): split it along what each consumer calls and have consumers depend on the small interfaces"},
		},
		{
			name: "declared package name of a differently named directory is resolved",
			files: map[string]string{
				"internal/v2/reader.go":   reader,
				"internal/store/store.go": "package store\n\nimport \"example.com/m/internal/v2\"\n\ntype Store interface {\n\tcube.Reader\n\tPut()\n}\n",
			},
			want: []string{"R28 internal/store/store.go:5 interface Store has 4 methods once embedded cube.Reader are flattened (max 3): split it along what each consumer calls and have consumers depend on the small interfaces"},
		},
		{
			name: "import of the module root is resolved",
			files: map[string]string{
				"root.go":                 "package m\n\ntype Reader interface {\n\tLoad()\n\tKeys()\n\tSquare()\n}\n",
				"internal/cube/x.go":      "package cube\n",
				"internal/store/store.go": "package store\n\nimport (\n\t\"example.com/m\"\n\t\"example.com/m/internal/cube\"\n)\n\ntype Store interface {\n\tm.Reader\n\tPut()\n}\n\nvar _ = cube.X\n",
			},
			want: []string{"R28 internal/store/store.go:8 interface Store has 4 methods once embedded m.Reader are flattened (max 3): split it along what each consumer calls and have consumers depend on the small interfaces"},
		},
		{
			name: "external import whose path ends in a local directory is outvoted",
			files: map[string]string{
				"internal/cube/reader.go":  reader,
				"internal/store/store.go":  "package store\n\nimport \"example.com/m/internal/cube\"\n\nvar _ cube.Reader = nil\n",
				"internal/store/other.go":  "package store\n\nimport \"example.com/m/internal/cube\"\n\nvar _ cube.Reader = nil\n",
				"internal/store/vendor.go": "package store\n\nimport cube \"github.com/x/y/internal/cube\"\n\ntype Vendor interface {\n\tcube.Reader\n\tPut()\n}\n",
			},
			want: nil,
		},
		{
			name: "blank, dot and stdlib imports are not resolved",
			files: map[string]string{
				"internal/cube/reader.go": reader,
				"internal/store/store.go": "package store\n\nimport (\n\t_ \"example.com/m/internal/cube\"\n\t. \"example.com/m/internal/cube\"\n\t\"io\"\n)\n\ntype Store interface {\n\tio.Reader\n\tPut()\n}\n",
			},
			want: nil,
		},
		{
			name: "files without local imports resolve only same-package embeds",
			files: map[string]string{
				"internal/cube/reader.go": reader,
				"internal/store/store.go": "package store\n\nimport \"fmt\"\n\ntype Store interface {\n\tfmt.Stringer\n\tPut()\n}\n",
			},
			want: nil,
		},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			interfacesize.Interfacesize(parsedFiles(
				t,
				tc.files,
			)),
			tc.want,
		)
	}
}
