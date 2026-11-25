package packagehygiene_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/shape/internal/packagehygiene"
)

func TestPackagehygieneNames(
	t *testing.T,
) {
	cases := []struct {
		name string
		rel  string
		code string
		want []string
	}{
		{
			name: "content-free names are reported",
			rel:  "internal/core/core.go",
			code: "package core\n",
			want: []string{"R22 internal/core/core.go:1"},
		},
		{
			name: "filler names are reported",
			rel:  "internal/orders/internal/extra/a.go",
			code: "package extra\n",
			want: []string{"R22 internal/orders/internal/extra/a.go:1"},
		},
		{
			name: "escape-hatch names are reported",
			rel:  "internal/orders/internal/legacy/a.go",
			code: "package legacy\n",
			want: []string{"R22 internal/orders/internal/legacy/a.go:1"},
		},
		{
			name: "data and glue packages are reported",
			rel:  "internal/glue/a.go",
			code: "package data\n",
			want: []string{"R22 internal/glue/a.go:1", "R21 internal/glue", "R22 internal/glue"},
		},
		{
			name: "util and helper affixes are reported",
			rel:  "internal/internalutil/a.go",
			code: "package internalutil\n",
			want: []string{"R22 internal/internalutil/a.go:1"},
		},
		{
			name: "legacy and compat affixes are reported",
			rel:  "internal/compatauth/a.go",
			code: "package compatauth\n",
			want: []string{"R22 internal/compatauth/a.go:1"},
		},
		{
			name: "names that only contain an escape word pass",
			rel:  "internal/manifold/a.go",
			code: "package manifold\n",
			want: nil,
		},
		{
			name: "template packages pass",
			rel:  "internal/tmpl/a.go",
			code: "package tmpl\n",
			want: nil,
		},
		{
			name: "numbered names are reported",
			rel:  "internal/orders/internal/part2/a.go",
			code: "package part2\n",
			want: []string{"R22 internal/orders/internal/part2/a.go:1"},
		},
		{
			name: "numbered real terms pass",
			rel:  "internal/oauth2/a.go",
			code: "package oauth2\n",
			want: nil,
		},
		{
			name: "plural names are reported",
			rel:  "internal/handlers/a.go",
			code: "package handlers\n",
			want: []string{"R21 internal/handlers/a.go:1"},
		},
		{
			name: "singular names ending in s pass",
			rel:  "internal/status/a.go",
			code: "package status\n",
			want: nil,
		},
		{
			name: "mass nouns ending in ics pass",
			rel:  "internal/metrics/a.go",
			code: "package metrics\n",
			want: nil,
		},
		{
			name: "short acronyms ending in s pass",
			rel:  "internal/tls/a.go",
			code: "package tls\n",
			want: nil,
		},
		{
			name: "listed singular names ending in s pass",
			rel:  "internal/postgres/a.go",
			code: "package postgres\n",
			want: nil,
		},
		{
			name: "plural vocabulary package names pass",
			rel:  "internal/bus/models/a.go",
			code: "package models\n",
			want: nil,
		},
		{
			name: "real nouns that share a filler stem pass",
			rel:  "internal/rest/a.go",
			code: "package rest\n",
			want: nil,
		},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			packagehygiene.Packagehygiene(parsedAll(
				t,
				map[string]string{tc.rel: tc.code},
			)),
			tc.want,
		)
	}
}
