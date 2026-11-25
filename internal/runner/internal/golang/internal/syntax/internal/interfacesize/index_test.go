package interfacesize_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/interfacesize"
)

func TestInterfacesizeIndex(
	t *testing.T,
) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{
			name:  "interfaces declared in test files are ignored",
			files: map[string]string{"internal/cube/reader.go": reader, "internal/cube/writer.go": writer, "internal/cube/squarer_test.go": squarer},
			want:  nil,
		},
		{
			name:  "interfaces declared in mocks are ignored",
			files: map[string]string{"internal/cube/mocks/reader.go": "package mocks\n\ntype Reader interface {\n\tLoad()\n\tKeys()\n}\n", "internal/cube/mocks/squarer.go": "package mocks\n\ntype Squarer interface {\n\tReader\n\tStore()\n\tDelete()\n}\n"},
			want:  nil,
		},
		{
			name:  "same-named interface in another package is not resolved",
			files: map[string]string{"internal/other/reader.go": "package other\n\ntype Reader interface {\n\tLoad()\n\tKeys()\n\tSquare()\n}\n", "internal/cube/squarer.go": "package cube\n\ntype Squarer interface {\n\tReader\n\tStore()\n}\n"},
			want:  nil,
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
