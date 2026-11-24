package maxdepth_test

import (
	"go/parser"
	"go/token"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/report"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/astrules/internal/maxdepth"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestMaxdepth(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "if inside a loop body stays at level two",
			code: `package p

func f(
	xs []int,
) {
	for _, x := range xs {
		work(x)
	}
	if len(xs) > 0 {
		return
	}
}

func work(
	x int,
) {
}
`,
			want: nil,
		},
		{
			name: "if inside a loop is reported at its first level three statement",
			code: `package p

func f(
	xs []int,
) {
	for _, x := range xs {
		if x > 0 {
			return
		}
	}
}
`,
			want: []string{"R02 p.go:8"},
		},
		{
			name: "every level three block is reported once, not once per statement",
			code: `package p

func f(
	xs []int,
) {
	for _, x := range xs {
		if x > 0 {
			xs[0] = x
			xs[1] = x
		}
		if x < 0 {
			return
		}
	}
}
`,
			want: []string{"R02 p.go:8", "R02 p.go:12"},
		},
		{
			name: "deeper blocks are each reported",
			code: `package p

func f(
	xs []int,
) {
	for range xs {
		for range xs {
			if len(xs) > 0 {
				return
			}
		}
	}
}
`,
			want: []string{"R02 p.go:8", "R02 p.go:9"},
		},
		{
			name: "if inside an if is reported",
			code: `package p

func f(
	x int,
) {
	if x > 0 {
		if x > 1 {
			return
		}
	}
}
`,
			want: []string{"R02 p.go:8"},
		},
		{
			name: "else branch body counts like the if body",
			code: `package p

func f(
	x int,
) {
	if x > 0 {
		return
	} else if x < 0 {
		return
	}
}
`,
			want: nil,
		},
		{
			name: "switch case bodies are level two",
			code: `package p

func f(
	x int,
) {
	switch x {
	case 1:
		return
	default:
		return
	}
}
`,
			want: nil,
		},
		{
			name: "switch inside an if puts case bodies at level three",
			code: `package p

func f(
	x int,
) {
	if x > 0 {
		switch x {
		case 1:
			return
		}
	}
}
`,
			want: []string{"R02 p.go:9"},
		},
		{
			name: "select case bodies are level two",
			code: `package p

func f(
	done chan struct{},
) {
	select {
	case <-done:
		return
	}
}
`,
			want: nil,
		},
		{
			name: "closure body directly in the function is level two",
			code: `package p

func f() {
	run := func() {
		work()
	}
	run()
}

func work() {
}
`,
			want: nil,
		},
		{
			name: "if inside a closure is reported",
			code: `package p

func f() {
	run := func() {
		if true {
			return
		}
	}
	run()
}
`,
			want: []string{"R02 p.go:6"},
		},
		{
			name: "closure inside a loop is reported",
			code: `package p

func f(
	xs []int,
) {
	for range xs {
		go func() {
			work()
		}()
	}
}

func work() {
}
`,
			want: []string{"R02 p.go:8"},
		},
		{
			name: "closure passed as a multi-line call argument is reported",
			code: `package p

func f() {
	run(
		func() {
			work()
		},
	)
}

func run(
	callback func(),
) {
	callback()
}

func work() {
}
`,
			want: []string{"R02 p.go:6"},
		},
		{
			name: "closure buried in nested call arguments is reported",
			code: `package p

func f() {
	run(
		wrap(
			func() {
				work()
			},
		),
	)
}
`,
			want: []string{"R02 p.go:7"},
		},
		{
			name: "single-line closure adds no indentation level",
			code: `package p

func f(
	xs []int,
) {
	for range xs {
		defer func() { work() }()
	}
}

func work() {
}
`,
			want: nil,
		},
		{
			name: "composite literal elements and call arguments are not statements",
			code: `package p

func f() (
	m map[string][]int,
) {
	work(
		1,
		2,
	)
	return map[string][]int{
		"a": {
			1,
		},
	}
}

func work(
	a int,
	b int,
) {
}
`,
			want: nil,
		},
		{
			name: "if header init statements are not nesting",
			code: `package p

func f() (
	err error,
) {
	if err := work(); err != nil {
		return err
	}
	for i := 0; i < 3; i++ {
		work()
	}
	return nil
}

func work() (
	err error,
) {
	return nil
}
`,
			want: nil,
		},
		{
			name: "labeled loop is measured at its statement",
			code: `package p

func f() {
outer:
	for {
		break outer
	}
}
`,
			want: nil,
		},
		{
			name: "method bodies are measured like functions",
			code: `package p

type t struct{}

func (r t) m(
	xs []int,
) {
	for range xs {
		if len(xs) > 0 {
			return
		}
	}
}
`,
			want: []string{"R02 p.go:10"},
		},
		{
			name: "package-level func literal is measured from its declaration",
			code: `package p

var run = func() {
	if true {
		return
	}
}
`,
			want: nil,
		},
		{
			name: "too-deep package-level func literal is reported",
			code: `package p

var run = func() {
	for {
		if true {
			return
		}
	}
}
`,
			want: []string{"R02 p.go:6"},
		},
		{
			name: "func literal in a grouped var is measured from its spec",
			code: `package p

var (
	run = func() {
		if true {
			return
		}
	}
	jump = func() {
		for {
			if true {
				return
			}
		}
	}
)
`,
			want: []string{"R02 p.go:12"},
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

func parsed(
	t *testing.T,
	code string,
) (file source.File) {
	t.Helper()
	fileSet := token.NewFileSet()
	syntax, err := parser.ParseFile(
		fileSet,
		"p.go",
		code,
		parser.ParseComments,
	)
	if err != nil {
		t.Fatalf(
			"parse fixture: %v",
			err,
		)
	}

	return source.File{Path: "p.go", Rel: "p.go", FileSet: fileSet, Syntax: syntax}
}

func assertFindings(
	t *testing.T,
	name string,
	findings []report.Finding,
	want []string,
) {
	t.Helper()
	got := []string{}
	for _, finding := range findings {
		got = append(
			got,
			finding.Rule+" "+finding.Location,
		)
	}

	isMismatch := !slices.Equal(
		got,
		want,
	) && len(got)+len(want) > 0
	if isMismatch {
		t.Errorf(
			"%s: findings = %v, want %v",
			name,
			got,
			want,
		)
	}
}
