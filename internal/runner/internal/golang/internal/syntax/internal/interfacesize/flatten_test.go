package interfacesize_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/interfacesize"
)

func TestInterfacesizeFlatten(
	t *testing.T,
) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{
			name:  "overlapping embedded methods are counted once",
			files: map[string]string{"internal/cube/both.go": "package cube\n\ntype Both interface {\n\tLeft\n\tRight\n}\n\ntype Left interface {\n\tA()\n\tB()\n}\n\ntype Right interface {\n\tA()\n\tB()\n}\n"},
			want:  nil,
		},
		{
			name:  "nested embedding is flattened transitively",
			files: map[string]string{"internal/cube/deep.go": "package cube\n\ntype Deep interface {\n\tMiddle\n}\n\ntype Middle interface {\n\tInner\n\tC()\n\tD()\n}\n\ntype Inner interface {\n\tA()\n\tB()\n}\n"},
			want:  []string{"R28 internal/cube/deep.go:3 interface Deep has 4 methods once embedded Middle are flattened (max 3): split it along what each consumer calls and have consumers depend on the small interfaces", "R28 internal/cube/deep.go:7 interface Middle has 4 methods once embedded Inner are flattened (max 3): split it along what each consumer calls and have consumers depend on the small interfaces"},
		},
		{
			name:  "cyclic embedding terminates",
			files: map[string]string{"internal/cube/cycle.go": "package cube\n\ntype Ping interface {\n\tPong\n\tA()\n}\n\ntype Pong interface {\n\tPing\n\tB()\n}\n"},
			want:  nil,
		},
		{
			name:  "generic embedded interface is resolved",
			files: map[string]string{"internal/cube/sink.go": "package cube\n\ntype Sink[T any] interface {\n\tPut(\n\t\tv T,\n\t)\n\tFlush()\n}\n\ntype Pipe interface {\n\tSink[int]\n\tOpen()\n\tClose()\n}\n"},
			want:  []string{"R28 internal/cube/sink.go:10 interface Pipe has 4 methods once embedded Sink[int] are flattened (max 3): split it along what each consumer calls and have consumers depend on the small interfaces"},
		},
		{
			name:  "union constraint and any embeds pass",
			files: map[string]string{"internal/cube/num.go": "package cube\n\ntype Number interface {\n\t~int | ~float64\n\tany\n}\n"},
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
