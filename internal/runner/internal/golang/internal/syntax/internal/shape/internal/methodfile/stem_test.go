package methodfile_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/shape/internal/methodfile"
)

func TestMethodfileStems(
	t *testing.T,
) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{
			name:  "leading acronym receiver matches its snake case file",
			files: map[string]string{"web/http_server.go": "package web\n\nfunc (s *HTTPServer) Serve() {}\n"},
			want:  nil,
		},
		{
			name:  "multi parameter generic receiver matches its file",
			files: map[string]string{"pair/pair.go": "package pair\n\nfunc (p Pair[K, V]) Key() (k K) {\n\treturn p.k\n}\n"},
			want:  nil,
		},
		{
			name:  "parenthesized receiver matches its file",
			files: map[string]string{"cube/cube.go": "package cube\n\nfunc (c (*cube)) Size() {}\n"},
			want:  nil,
		},
		{
			name:  "receiver named after a word of a longer file stem is reported",
			files: map[string]string{"cube/cube_helpers.go": "package cube\n\nfunc (c *cube) Size() {}\n"},
			want:  []string{"cube/cube_helpers.go:3"},
		},
		{
			name:  "filler word with digits is reported",
			files: map[string]string{"cube/chunk3.go": plainCode},
			want:  []string{"cube/chunk3.go"},
		},
		{
			name:  "platform name with digits passes",
			files: map[string]string{"cube/plan9.go": plainCode, "cube/cube_arm64.go": plainCode},
			want:  nil,
		},
	}
	for _, tc := range cases {
		assertLocations(
			t,
			tc.name,
			methodfile.Methodfile(parsedAll(
				t,
				tc.files,
			)),
			tc.want,
		)
	}
}
