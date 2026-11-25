package entrypoint_test

import "testing"

func TestEntrypointCalls(
	t *testing.T,
) {
	assertBodies(
		t,
		[]bodyCase{
			{name: "own internal package call passes", body: "return impl.New(size), nil", want: nil},
			{name: "own declaration call passes", body: "return build()", want: nil},
			{name: "other module constructor passes", body: "impl.Use(lib.New())\nreturn nil, nil", want: nil},
			{name: "method call on a local passes", body: "q, err := build()\nq.Start()\nreturn q, err", want: nil},
			{name: "func parameter call passes", body: "impl.Use(size)\nreturn nil, nil", want: nil},
			{name: "error constructors pass", body: "return nil, errors.Join(errors.New(\"a\"), fmt.Errorf(\"b: %w\", impl.Err))", want: nil},
			{name: "slices.Concat passes", body: "impl.Use(slices.Concat(impl.A(), impl.B()))\nreturn nil, nil", want: nil},
			{name: "generic constructor passes", body: "impl.Use(impl.New[int]())\nreturn nil, nil", want: nil},
			{name: "environment lookup is reported", body: "impl.Use(os.Getenv(\"MODE\"))\nreturn nil, nil", want: []string{"R14 container.go:23"}},
			{name: "strings call is reported", body: "impl.Use(strings.ToLower(\"A\"))\nreturn nil, nil", want: []string{"R14 container.go:23"}},
			{name: "os.Exit is reported outside main", body: "os.Exit(1)\nreturn nil, nil", want: []string{"R14 container.go:23"}},
			{name: "builtin call is reported", body: "impl.Use(len(\"a\"))\nreturn nil, nil", want: []string{"R14 container.go:23"}},
			{name: "conversion to a composite type is reported", body: "impl.Use([]byte(\"a\"))\nreturn nil, nil", want: []string{"R14 container.go:23"}},
			{name: "shadowed qualifier is a method call", body: "os := impl.Open()\nos.Getenv()\nreturn nil, nil", want: nil},
			{name: "other module result assigned and passed on passes", body: "client, err := lib.New()\nreturn impl.New(client), err", want: nil},
			{name: "other module result returned passes", body: "return lib.New(), nil", want: nil},
			{name: "other module call whose result is discarded is reported", body: "lib.Init()\nreturn nil, nil", want: []string{"R14 container.go:23"}},
			{name: "other module result bound to blank is reported", body: "_ = lib.New()\nreturn nil, nil", want: []string{"R14 container.go:23"}},
			{name: "other module result used in an expression is reported", body: "impl.Use(*lib.Flag())\nreturn nil, nil", want: []string{"R14 container.go:23"}},
		},
	)
}

func TestEntrypointCallsInModulesWithoutDomains(
	t *testing.T,
) {
	assertFindings(
		t,
		"own packages of a module path without a dot are not standard library",
		checkedAt(
			t,
			"ledger",
			map[string]string{"container.go": "package ledger\n\nimport \"ledger/internal/book\"\n\nfunc New() int {\n\treturn book.New()\n}\n"},
		),
		nil,
	)
}

func TestEntrypointConfigurationReads(
	t *testing.T,
) {
	cases := []struct {
		name     string
		imported string
		call     string
	}{
		{name: "pflag lookup", imported: "github.com/spf13/pflag", call: "pflag.String(\"mode\", \"\", \"\")"},
		{name: "viper lookup", imported: "github.com/spf13/viper", call: "viper.GetString(\"mode\")"},
		{name: "versioned env parser", imported: "github.com/caarlos0/env/v11", call: "env.Parse(nil)"},
		{name: "envconfig", imported: "github.com/kelseyhightower/envconfig", call: "envconfig.Process(\"app\", nil)"},
		{name: "unix environment", imported: "golang.org/x/sys/unix", call: "unix.Getenv(\"MODE\")"},
		{name: "go-prefixed envconfig", imported: "github.com/sethvargo/go-envconfig", call: "envconfig.MustProcess(nil, nil)"},
		{name: "lo helpers", imported: "github.com/samber/lo", call: "lo.Ternary(true, 1, 2)"},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name+" assigned and passed on is still reported",
			checked(
				t,
				map[string]string{"container.go": "package jobq\n\nimport (\n\t\"example.com/m/internal/impl\"\n\t\"" + tc.imported + "\"\n)\n\nfunc New() (q int) {\n\tvalue := " + tc.call + "\n\treturn impl.New(value)\n}\n"},
			),
			[]string{"R14 container.go:9"},
		)
	}
}

func TestEntrypointOtherUnixCallsAreWiring(
	t *testing.T,
) {
	assertFindings(
		t,
		"a unix constructor whose result is passed on passes",
		checked(
			t,
			map[string]string{"container.go": "package jobq\n\nimport (\n\t\"example.com/m/internal/impl\"\n\t\"golang.org/x/sys/unix\"\n)\n\nfunc New() (q int) {\n\tvalue := unix.Getpagesize()\n\treturn impl.New(value)\n}\n"},
		),
		nil,
	)
}

func TestEntrypointGopkgImportsResolve(
	t *testing.T,
) {
	assertFindings(
		t,
		"a gopkg.in call whose result is discarded is reported",
		checked(
			t,
			map[string]string{"container.go": "package jobq\n\nimport \"gopkg.in/yaml.v3\"\n\nfunc New() (q int) {\n\tyaml.Unmarshal(nil, nil)\n\treturn 1\n}\n"},
		),
		[]string{"R14 container.go:6"},
	)
}
