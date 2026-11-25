package publicapi_test

import "testing"

func TestPublicapiScopeLabels(
	t *testing.T,
) {
	findings := checked(
		t,
		map[string]string{
			"jobq.go":                      "package jobq\n\nfunc helper() {}\n",
			"internal/svc/svc.go":          "package svc\n\nfunc helper() {}\n",
			"internal/svc/internal/a/a.go": "package a\n",
			"pkg/cache/cache.go":           "package cache\n\nfunc helper() {}\n",
		},
	)
	assertSnippet(
		t,
		findings,
		"jobq.go:3",
		"func helper in module root jobq: move it under internal/",
	)
	assertSnippet(
		t,
		findings,
		"internal/svc/svc.go:3",
		"func helper in module root svc: move it under internal/svc/internal/",
	)
	assertSnippet(
		t,
		findings,
		"pkg/cache/cache.go:3",
		"func helper in public package cache: move it under pkg/cache/internal/",
	)
}
