package ctxfirstparam_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/behavior/internal/ctxfirstparam"
)

func TestCtxfirstparamStored(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{name: "context parameter in a composite literal field is reported", code: "package p\n\nimport \"context\"\n\nfunc f(\n\tctx context.Context,\n) (s *state) {\n\treturn &state{\n\t\tparent: ctx,\n\t}\n}\n", want: []string{"p.go:9"}},
		{name: "context parameter in an unkeyed literal element is reported", code: "package p\n\nimport \"context\"\n\nfunc f(\n\tctx context.Context,\n) (values []any) {\n\treturn []any{ctx}\n}\n", want: []string{"p.go:8"}},
		{name: "context parameter assigned to a field is reported", code: "package p\n\nimport \"context\"\n\nfunc (s *state) bind(\n\tctx context.Context,\n) {\n\ts.holder = ctx\n}\n", want: []string{"p.go:8"}},
		{name: "context parameter sent on a channel is reported", code: "package p\n\nimport \"context\"\n\nfunc f(\n\tctx context.Context,\n\tout chan any,\n) {\n\tout <- ctx\n}\n", want: []string{"p.go:9"}},
		{name: "derived context stored in a literal is reported", code: "package p\n\nimport \"context\"\n\nfunc f() (s state) {\n\tctx, cancel := context.WithCancel(context.Background())\n\tdefer cancel()\n\treturn state{holder: ctx}\n}\n", want: []string{"p.go:8"}},
		{name: "method value of a context parameter is reported", code: "package p\n\nimport \"context\"\n\nfunc f(\n\tctx context.Context,\n) (s state) {\n\treturn state{canceled: ctx.Err}\n}\n", want: []string{"p.go:8"}},
		{name: "called context methods pass", code: "package p\n\nimport \"context\"\n\nfunc f(\n\tctx context.Context,\n) (err error) {\n\t<-ctx.Done()\n\treturn ctx.Err()\n}\n", want: nil},
		{name: "context passed as an argument passes", code: "package p\n\nimport \"context\"\n\nfunc f(\n\tctx context.Context,\n) (err error) {\n\tlocal := ctx\n\treturn run(local)\n}\n", want: nil},
		{name: "non-context parameter stored in a literal passes", code: "package p\n\nfunc f(\n\tname string,\n) (s state) {\n\treturn state{name: name}\n}\n", want: nil},
		{name: "context derived by another package passes", code: "package p\n\nfunc f() (s state) {\n\tctx := other.Context()\n\treturn state{holder: ctx}\n}\n", want: nil},
		{name: "unresolved identifier named ctx passes", code: "package p\n\nfunc f() (s state) {\n\treturn state{holder: ctx}\n}\n", want: nil},
		{name: "mismatched multi-assignment passes", code: "package p\n\nimport \"context\"\n\nfunc f(\n\tctx context.Context,\n) {\n\ts.a, s.b = pair(ctx)\n}\n", want: nil},
		{name: "derived context by a non-call passes", code: "package p\n\nfunc f() (s state) {\n\tctx := parent\n\treturn state{holder: ctx}\n}\n", want: nil},
		{name: "derived context from an indexed call passes", code: "package p\n\nfunc f() (s state) {\n\tctx := makers[0]()\n\treturn state{holder: ctx}\n}\n", want: nil},
		{name: "second result of a context call passes", code: "package p\n\nimport \"context\"\n\nfunc f() (s state) {\n\tparent, ctx := context.WithCancel(context.Background())\n\treturn state{holder: ctx, parent: parent}\n}\n", want: []string{"p.go:7"}},
	}
	for _, tc := range cases {
		assertLocations(
			t,
			tc.name,
			ctxfirstparam.Ctxfirstparam([]source.File{parsed(
				t,
				tc.code,
			)}),
			tc.want,
		)
	}
}
