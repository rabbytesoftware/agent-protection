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
				name:  "package with a non-vocabulary name passes",
				files: map[string]string{"internal/modelsx/m.go": "package modelsx\n\nfunc helper() {}\n"},
				want:  nil,
			},
		},
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
