package commentpolicy_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/commentpolicy"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestCommentpolicyReportsRestatingDocs(
	t *testing.T,
) {
	cases := []struct {
		name string
		path string
		code string
		want []string
	}{
		{
			name: "type doc that inflects its own name is reported",
			path: "p.go",
			code: "package p\n\n// CreateOrderCmd creates a new order.\ntype CreateOrderCmd struct{}\n",
			want: []string{"R07 p.go:3"},
		},
		{
			name: "agent noun and verb share a stem",
			path: "p.go",
			code: "package p\n\n// Reader reads.\ntype Reader interface{}\n",
			want: []string{"R07 p.go:3"},
		},
		{
			name: "method doc restating name and receiver is reported",
			path: "p.go",
			code: "package p\n\ntype shardPool struct{}\n\n// Shards returns the pool shards.\nfunc (p *shardPool) Shards() {\n}\n",
			want: []string{"R07 p.go:5"},
		},
		{
			name: "receiver words do not count for a plain function",
			path: "p.go",
			code: "package p\n\n// Shards returns the pool shards.\nfunc Shards() {\n}\n",
			want: nil,
		},
		{
			name: "templated fake doc is reported",
			path: "p_test.go",
			code: "package p\n\ntype fakeBus struct{}\n\n// Publish satisfies the interface that fakeBus fakes.\nfunc (b fakeBus) Publish() {\n}\n",
			want: []string{"R07 p_test.go:5"},
		},
		{
			name: "test double boilerplate is reported",
			path: "p_test.go",
			code: "package p\n\ntype blockingBus struct{}\n\n// Close is the test double implementation for the code under test.\nfunc (b blockingBus) Close() {\n}\n",
			want: []string{"R07 p_test.go:5"},
		},
		{
			name: "short doc that adds a behavior passes",
			path: "p.go",
			code: "package p\n\ntype nopBus struct{}\n\n// Close does nothing and succeeds.\nfunc (b nopBus) Close() {\n}\n",
			want: nil,
		},
		{
			name: "short doc that adds an object passes",
			path: "p.go",
			code: "package p\n\n// AggregateReader reads aggregate state.\ntype AggregateReader interface{}\n",
			want: nil,
		},
		{
			name: "sole spec documented on its declaration is checked",
			path: "p.go",
			code: "package p\n\n// Limit is the limit.\nconst Limit = 3\n",
			want: []string{"R07 p.go:3"},
		},
		{
			name: "spec doc inside a group is checked",
			path: "p.go",
			code: "package p\n\nconst (\n\t// Limit is the limit.\n\tLimit = 3\n\t// Floor is the lowest value accepted.\n\tFloor = 1\n)\n",
			want: []string{"R07 p.go:4"},
		},
		{
			name: "group doc over several specs is not a spec doc",
			path: "p.go",
			code: "package p\n\n// Limits.\nconst (\n\t// Limit caps retries.\n\tLimit = 3\n\t// Floor is the lowest value accepted.\n\tFloor = 1\n)\n",
			want: nil,
		},
		{
			name: "deprecation note carries information",
			path: "p.go",
			code: "package p\n\n// Old returns old.\n//\n// Deprecated: use Fresh.\nfunc Old() {\n}\n",
			want: nil,
		},
		{
			name: "unexported declarations are reported once, as commented",
			path: "p.go",
			code: "package p\n\n// limit is the limit.\nconst limit = 3\n",
			want: []string{"R07 p.go:3"},
		},
		{
			name: "test function doc is reported once, as commented",
			path: "p_test.go",
			code: "package p\n\nimport \"testing\"\n\n// TestF tests F.\nfunc TestF(\n\tt *testing.T,\n) {\n}\n",
			want: []string{"R07 p_test.go:5"},
		},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			commentpolicy.Commentpolicy([]source.File{parsed(
				t,
				tc.path,
				tc.code,
			)}),
			tc.want,
		)
	}
}
