package namedreturn_test

import (
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/namedreturn"
)

func TestUnnamedResults(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "unnamed results of a constructor name the function and the fix",
			code: "package p\n\nfunc New(\n\tcfg Config,\n) (Queue, error) {\n\treturn Queue{}, nil\n}\n",
			want: []string{"p.go:5 func New returns unnamed results (Queue, error): name every result after what it holds, for example (queue Queue, err error)"},
		},
		{
			name: "unnamed func-typed result is reported on the outer function",
			code: "package p\n\nfunc Backoff() func(attempt int) time.Duration {\n\treturn nil\n}\n",
			want: []string{"p.go:3 func Backoff returns unnamed results (func(attempt int) time.Duration): name every result after what it holds, for example (<name> func(attempt int) time.Duration)"},
		},
		{
			name: "method on a pointer receiver names the type",
			code: "package p\n\nfunc (q *Queue) Jobs() []*pkg.Job {\n\treturn nil\n}\n",
			want: []string{"p.go:3 method Queue.Jobs returns unnamed results ([]*pkg.Job): name every result after what it holds, for example (jobs []*pkg.Job)"},
		},
		{
			name: "generic result types are named after the base type",
			code: "package p\n\nfunc (b Builder[T]) Build() (Bus[T], Pair[K, V]) {\n\treturn nil, nil\n}\n",
			want: []string{"p.go:3 method Builder[T].Build returns unnamed results (Bus[T], Pair[K, V]): name every result after what it holds, for example (bus Bus[T], pair Pair[K, V])"},
		},
		{
			name: "func literal results are reported",
			code: "package p\n\nvar less = func(a int, b int) bool {\n\treturn a < b\n}\n",
			want: []string{"p.go:3 func literal returns unnamed results (bool): name every result after what it holds, for example (<name> bool)"},
		},
		{
			name: "keyword-like and builtin type names fall back to a placeholder",
			code: "package p\n\nfunc f() (Func, map[string]int, []string) {\n\treturn nil, nil, nil\n}\n",
			want: []string{"p.go:3 func f returns unnamed results (Func, map[string]int, []string): name every result after what it holds, for example (<name> Func, <name> map[string]int, <name> []string)"},
		},
		{
			name: "named results, no results and test functions pass",
			code: "package p\n\nfunc f() (queue Queue, err error) {\n\treturn Queue{}, nil\n}\n\nfunc g() {\n}\n\nfunc TestX(t *testing.T) {\n}\n",
			want: nil,
		},
		{
			name: "interface methods, func-typed fields, params and bodiless declarations pass",
			code: "package p\n\ntype I interface {\n\tGet() int\n}\n\ntype S struct {\n\tf func() int\n}\n\nfunc h(\n\tg func() error,\n) {\n}\n\nfunc asm() int\n",
			want: nil,
		},
	}
	for _, tc := range cases {
		assertSnippets(
			t,
			tc.name,
			tc.code,
			tc.want,
		)
	}
}

func assertSnippets(
	t *testing.T,
	name string,
	code string,
	want []string,
) {
	t.Helper()
	got := []string{}
	for _, finding := range namedreturn.Namedreturn([]source.File{parsed(
		t,
		code,
	)}) {
		got = append(
			got,
			finding.Location+" "+finding.Snippet,
		)
	}

	isMismatch := !slices.Equal(
		got,
		want,
	) && len(got)+len(want) > 0
	if isMismatch {
		t.Errorf(
			"%s: findings = %q, want %q",
			name,
			got,
			want,
		)
	}
}
