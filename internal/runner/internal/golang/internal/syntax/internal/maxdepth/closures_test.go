package maxdepth_test

import (
	"strings"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/maxdepth"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestMaxdepthMeasuresLocalClosuresWhereTheyAreUsed(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "closure hoisted out of a multi-line call argument is reported",
			code: `package p

func f() {
	cond := func() bool {
		return true
	}
	waitFor(
		"message",
		cond,
	)
}
`,
			want: []string{"R02 p.go:5"},
		},
		{
			name: "closure hoisted out of a composite literal field is reported",
			code: `package p

func f() {
	match := func() bool {
		return true
	}
	run(sub{
		Match: match,
	})
}
`,
			want: []string{"R02 p.go:5"},
		},
		{
			name: "closure called at the function's top level stays at level two",
			code: `package p

func f() {
	cleanup := func() {
		work()
	}
	defer cleanup()
	go cleanup()
	cleanup()
}
`,
			want: nil,
		},
		{
			name: "closure passed to a one-line call at the top level stays at level two",
			code: `package p

func f() {
	cleanup := func() {
		work()
	}
	register(cleanup)
	if err := check(cleanup); err != nil {
		return
	}
}
`,
			want: nil,
		},
		{
			name: "closure called inside a loop is reported",
			code: `package p

func f(
	xs []int,
) {
	step := func() {
		work()
	}
	for range xs {
		step()
	}
}
`,
			want: []string{"R02 p.go:7"},
		},
		{
			name: "closure started as a goroutine inside a loop is reported",
			code: `package p

func f(
	xs []int,
) {
	worker := func() {
		work()
	}
	for range xs {
		go worker()
	}
}
`,
			want: []string{"R02 p.go:7"},
		},
		{
			name: "closure deferred inside an if is reported",
			code: `package p

func f(
	ok bool,
) {
	cleanup := func() {
		work()
	}
	if ok {
		defer cleanup()
	}
}
`,
			want: []string{"R02 p.go:7"},
		},
		{
			name: "sort.Slice less function hoisted into a variable is reported",
			code: `package p

func f(
	xs []int,
) {
	less := func(
		i int,
		j int,
	) bool {
		return xs[i] < xs[j]
	}
	sort.Slice(
		xs,
		less,
	)
}
`,
			want: []string{"R02 p.go:10"},
		},
		{
			name: "t.Run body hoisted out of a table loop is reported at its loop depth",
			code: `package p

func f(
	t *testing.T,
) {
	check := func(
		t *testing.T,
	) {
		t.Helper()
	}
	for _, name := range names {
		t.Run(
			name,
			check,
		)
	}
}
`,
			want: []string{"R02 p.go:9"},
		},
		{
			name: "slices.DeleteFunc predicate declared with var is reported",
			code: `package p

func f(
	xs []int,
) {
	var isOdd = func(
		x int,
	) bool {
		return x%2 == 1
	}
	xs = slices.DeleteFunc(
		xs,
		isOdd,
	)
}
`,
			want: []string{"R02 p.go:9"},
		},
		{
			name: "if inside a deeply used closure is reported once per block",
			code: `package p

func f(
	xs []int,
) {
	step := func() {
		if len(xs) > 0 {
			return
		}
	}
	for range xs {
		step()
	}
}
`,
			want: []string{"R02 p.go:7", "R02 p.go:8"},
		},
		{
			name: "returned closure stays where it is written",
			code: `package p

func f() (
	run func(),
) {
	run = func() {
		work()
	}

	return run
}
`,
			want: nil,
		},
		{
			name: "package-level closure is a named symbol, not a local one",
			code: `package p

var step = func() {
	work()
}

func f(
	xs []int,
) {
	for range xs {
		step()
	}
}
`,
			want: nil,
		},
		{
			name: "single-line closure used deep adds no indentation level",
			code: `package p

func f(
	xs []int,
) {
	step := func() { work() }
	for range xs {
		step()
	}
}
`,
			want: nil,
		},
		{
			name: "empty closure used deep has no block to report",
			code: `package p

func f(
	xs []int,
) {
	step := func() {
	}
	for range xs {
		step()
	}
}
`,
			want: nil,
		},
		{
			name: "recursive closure is measured where it is written",
			code: `package p

func f() {
	var walk func(int)
	walk = func(
		n int,
	) {
		walk(n - 1)
	}
	walk(3)
}
`,
			want: nil,
		},
		{
			name: "shallow wrapper closure carries its use depth to the closure it calls",
			code: `package p

func f(
	xs []int,
) {
	inner := func() {
		work()
	}
	outer := func() { inner() }
	for range xs {
		outer()
	}
}
`,
			want: []string{"R02 p.go:7"},
		},
		{
			name: "mutually recursive closures settle instead of looping forever",
			code: `package p

func f() {
	var ping, pong func()
	ping = func() {
		pong()
	}
	pong = func() {
		ping()
	}
	ping()
}
`,
			want: []string{"R02 p.go:6", "R02 p.go:9"},
		},
		{
			name: "shadowing variable in a nested scope is a different closure",
			code: `package p

func f(
	xs []int,
) {
	step := func() {
		work()
	}
	step()
	for range xs {
		step := work
		step()
	}
}
`,
			want: nil,
		},
		{
			name: "closure stored in a struct field is measured where it is written",
			code: `package p

func f() {
	cfg.OnError = func() {
		work()
	}
	run(
		ctx,
		cfg,
	)
}
`,
			want: nil,
		},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			maxdepth.Maxdepth([]source.File{parsed(
				t,
				tc.code,
			)}),
			tc.want,
		)
	}
}

func TestMaxdepthNamesTheLocalClosureInItsSnippet(
	t *testing.T,
) {
	findings := maxdepth.Maxdepth([]source.File{parsed(
		t,
		`package p

func f(
	xs []int,
) {
	step := func() {
		work()
	}
	for range xs {
		step()
	}
}
`,
	)})
	isClosureSnippet := len(findings) == 1 && strings.Contains(
		findings[0].Snippet,
		"f: block nested 3 levels deep where its local closure is used",
	)
	if !isClosureSnippet {
		t.Errorf(
			"findings = %+v, want one snippet naming the local closure",
			findings,
		)
	}
}
