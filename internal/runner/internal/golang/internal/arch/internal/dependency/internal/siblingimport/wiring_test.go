package siblingimport_test

import (
	"testing"
)

func TestSiblingimportsWiresSubtrees(
	t *testing.T,
) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{
			name: "a module whose children import two siblings is reported once at the module",
			files: map[string]string{
				"internal/invoice/container.go":            unit("invoice internal/invoice/internal/line internal/invoice/internal/total"),
				"internal/invoice/internal/line/line.go":   unit("line internal/tax"),
				"internal/invoice/internal/total/total.go": unit("total internal/fee"),
				"internal/tax/tax.go":                      unit("tax"),
				"internal/fee/fee.go":                      unit("fee"),
			},
			want: []string{"R47 internal/invoice"},
		},
		{
			name: "many children importing one shared foundation is one sibling",
			files: map[string]string{
				"internal/arch/container.go":        unit("arch internal/arch/internal/x internal/arch/internal/y"),
				"internal/arch/internal/x/x.go":     unit("x internal/source"),
				"internal/arch/internal/y/y.go":     unit("y internal/source internal/arch/internal/x"),
				"internal/source/source.go":         unit("source"),
				"internal/source/internal/w/w.go":   unit("w"),
				"internal/source/internal/w/w2.go":  unit("w"),
				"internal/arch/internal/y/other.go": unit("y"),
			},
			want: nil,
		},
		{
			name: "declaration-only siblings reached from a subtree do not count towards wiring",
			files: map[string]string{
				"internal/app/app.go":            unit("app internal/engine"),
				"internal/app/internal/x/x.go":   unit("x internal/contracts internal/failures"),
				"internal/engine/e.go":           unit("engine"),
				"internal/contracts/contract.go": declarationOnly,
				"internal/failures/failures.go":  errorsOnly,
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

func TestSiblingimportsNamesTheSubtreeImporter(
	t *testing.T,
) {
	findings := checked(
		t,
		map[string]string{
			"internal/invoice/container.go":            unit("invoice internal/invoice/internal/line internal/invoice/internal/total internal/tax"),
			"internal/invoice/internal/line/line.go":   unit("line internal/tax"),
			"internal/invoice/internal/total/total.go": unit("total internal/fee"),
			"internal/tax/tax.go":                      unit("tax"),
			"internal/fee/fee.go":                      unit("fee"),
		},
	)
	assertSnippets(
		t,
		findings,
		[]string{
			"imports siblings fee (from internal/invoice/internal/total), tax: wiring siblings together is the parent's job; move it into container.go and inject what this package needs through an interface it declares",
		},
	)
}
