package publicapi_test

import "testing"

func TestPublicapiVocabularyPackages(
	t *testing.T,
) {
	assertFixtures(
		t,
		[]fixture{
			{
				name:  "declarations in an internal models package pass",
				files: map[string]string{"internal/models/models.go": apiSurface},
				want:  nil,
			},
			{
				name:  "function in an internal models package is reported",
				files: map[string]string{"internal/models/models.go": "package models\n\nfunc positiveOr() {}\n"},
				want:  []string{"R57 internal/models/models.go:3"},
			},
			{
				name:  "method and unexported type in an internal options package are reported",
				files: map[string]string{"internal/x/options/options.go": "package options\n\ntype Config struct{}\n\ntype state int\n\nfunc (c Config) Validate() {}\n"},
				want:  []string{"R57 internal/x/options/options.go:5", "R57 internal/x/options/options.go:7"},
			},
			{
				name:  "public vocabulary package is reported by R14 only",
				files: map[string]string{"models/models.go": "package models\n\nfunc positiveOr() {}\n"},
				want:  []string{"R14 models/models.go:3"},
			},
			{
				name:  "clean container.go entry function in a vocabulary module passes",
				files: map[string]string{"api/container.go": "package api\n\nfunc New() int {\n\treturn build()\n}\n"},
				want:  nil,
			},
			{
				name:  "mock files in a vocabulary package pass",
				files: map[string]string{"internal/models/mock_models.go": "package models\n\nfunc fake() {}\n"},
				want:  nil,
			},
			{
				name: "String with logic in a package the root re-exports is reported",
				files: map[string]string{
					"price.go":                "package arith\n\nimport \"example.com/m/internal/money\"\n\ntype Price = money.Price\n",
					"internal/money/price.go": "package money\n\ntype Price int\n\nfunc (p Price) String() string {\n\tif p > 10 {\n\t\treturn \"big\"\n\t}\n\n\treturn \"small\"\n}\n\nfunc (p Price) Is(target error) bool {\n\treturn false\n}\n",
				},
				want: []string{"R14 price.go:5", "R57 internal/money/price.go:5"},
			},
			{
				name: "re-exported package with other behavior is left to R14 and R48",
				files: map[string]string{
					"price.go":                "package arith\n\nimport \"example.com/m/internal/money\"\n\ntype Price = money.Price\n",
					"internal/money/price.go": "package money\n\ntype Price int\n\nfunc (p Price) String() string {\n\tif p > 10 {\n\t\treturn \"big\"\n\t}\n\n\treturn \"small\"\n}\n",
					"internal/money/tax.go":   "package money\n\nfunc Tax() {}\n",
				},
				want: []string{"R14 price.go:5"},
			},
			{
				name: "String with logic in a package nothing re-exports passes",
				files: map[string]string{
					"price.go":                "package arith\n\nimport \"fmt\"\n\ntype Price int\n\ntype Seats = int\n\ntype External = fmt.Stringer\n",
					"internal/money/price.go": "package money\n\ntype Price int\n\nfunc (p Price) String() string {\n\tif p > 10 {\n\t\treturn \"big\"\n\t}\n\n\treturn \"small\"\n}\n",
				},
				want: nil,
			},
			{
				name:  "package with a non-vocabulary name passes",
				files: map[string]string{"internal/modelsx/m.go": "package modelsx\n\nfunc helper() {}\n"},
				want:  nil,
			},
		},
	)
}

func TestPublicapiReExportedRenderingSnippet(
	t *testing.T,
) {
	findings := checked(
		t,
		map[string]string{
			"price.go":                "package arith\n\nimport \"example.com/m/internal/money\"\n\ntype Price = money.Price\n",
			"internal/money/price.go": "package money\n\ntype Price int\n\nfunc (p Price) String() string {\n\treturn p.format()\n}\n",
		},
	)
	assertSnippet(
		t,
		findings,
		"internal/money/price.go:5",
		"error method Price.String in package money, which an API file re-exports: a re-exported package holds declarations only, and Error, Unwrap, Is, As and String hold a single return of a field, a literal, a concatenation or one fmt.Sprintf, fmt.Errorf or errors call; move the logic into the package that uses it",
	)
}

func TestPublicapiVocabularySnippets(
	t *testing.T,
) {
	findings := checked(
		t,
		map[string]string{"internal/models/models.go": "package models\n\nfunc positiveOr() {}\n\nvar cache = 1\n"},
	)
	assertSnippet(
		t,
		findings,
		"internal/models/models.go:3",
		"func positiveOr in vocabulary package models: vocabulary packages hold declarations only; move behavior into the package that uses it",
	)
	assertSnippet(
		t,
		findings,
		"internal/models/models.go:5",
		"var cache in vocabulary package models: vocabulary packages hold only exported types, constants and Err sentinels; move it into the package that uses it",
	)
}
