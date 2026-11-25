package siblingimport_test

import (
	"testing"
)

func TestSiblingimportsLiftsCousins(
	t *testing.T,
) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{
			name: "cousins behind a grouping directory without Go files are wired separately",
			files: map[string]string{
				"internal/checkout/checkout.go":    unit("checkout internal/store/order internal/store/payment"),
				"internal/store/order/order.go":    unit("order"),
				"internal/store/payment/pay.go":    unit("payment"),
				"internal/store/payment/refund.go": unit("payment"),
			},
			want: []string{"R47 internal/checkout"},
		},
		{
			name: "children of a package directory count as that one sibling",
			files: map[string]string{
				"internal/checkout/checkout.go":    unit("checkout internal/store/order internal/store/payment"),
				"internal/store/store.go":          unit("store"),
				"internal/store/order/order.go":    unit("order"),
				"internal/store/payment/pay.go":    unit("payment"),
				"internal/store/payment/refund.go": unit("payment"),
			},
			want: nil,
		},
		{
			name: "a cousin import is lifted to the sibling directory holding it",
			files: map[string]string{
				"internal/a/a.go":       unit("a internal/group/b"),
				"internal/group/b/b.go": unit("b internal/c"),
				"internal/c/c.go":       unit("c"),
			},
			want: []string{"R16 internal/a/a.go:4"},
		},
		{
			name: "layering inside a grouping directory is judged among its own children",
			files: map[string]string{
				"internal/a/a.go":       unit("a internal/group/b"),
				"internal/group/b/b.go": unit("b internal/group/c"),
				"internal/group/c/c.go": unit("c"),
			},
			want: nil,
		},
		{
			name: "mock packages do not count towards wiring",
			files: map[string]string{
				"internal/app/app.go":       unit("app internal/engine internal/mocks/store"),
				"internal/engine/e.go":      unit("engine"),
				"internal/mocks/store/s.go": unit("store internal/engine"),
			},
			want: nil,
		},
		{
			name: "a module child importing a package of its own module is not a sibling import",
			files: map[string]string{
				"svc/container.go":                  unit("svc svc/internal/app"),
				"svc/internal/app/container.go":     unit("app svc svc/internal/app/internal/x"),
				"svc/internal/app/internal/x/x.go":  unit("x svc/internal/app"),
				"svc/internal/app/internal/x/x2.go": unit("x svc/internal/other"),
				"svc/internal/other/o.go":           unit("other"),
			},
			want: nil,
		},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			checked(
				t,
				tc.files,
			),
			tc.want,
		)
	}
}

func TestSiblingimportsNamesCousinsByTheirPath(
	t *testing.T,
) {
	findings := checked(
		t,
		map[string]string{
			"internal/checkout/checkout.go": unit("checkout internal/store/order internal/store/payment"),
			"internal/store/order/order.go": unit("order"),
			"internal/store/payment/pay.go": unit("payment"),
		},
	)
	assertSnippets(
		t,
		findings,
		[]string{
			"imports siblings store/order, store/payment: wiring siblings together is the parent's job; move it into container.go and inject what this package needs through an interface it declares",
		},
	)
}
