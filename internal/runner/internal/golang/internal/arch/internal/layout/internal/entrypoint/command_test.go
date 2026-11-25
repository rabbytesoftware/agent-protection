package entrypoint_test

import "testing"

const mainWithExit = "package main\n\nimport (\n\t\"fmt\"\n\t\"os\"\n\n\t\"example.com/m/cmd/app/internal/app\"\n)\n\nvar version = \"dev\"\n\nconst name = \"app\"\n\nfunc main() {\n\terr := app.Run(os.Args, version, name)\n\tif err != nil {\n\t\tfmt.Fprintln(os.Stderr, err)\n\t\tos.Exit(1)\n\t}\n}\n"

func TestEntrypointPackageMain(
	t *testing.T,
) {
	assertFixtures(
		t,
		[]fixture{
			{
				name:  "main.go that only delegates passes",
				files: map[string]string{"cmd/app/main.go": mainWithExit},
				want:  nil,
			},
			{
				name:  "main.go exiting through os.Exit passes",
				files: map[string]string{"cmd/app/main.go": "package main\n\nimport (\n\t\"os\"\n\n\t\"example.com/m/cmd/app/internal/app\"\n)\n\nfunc main() {\n\tos.Exit(app.Run(os.Args))\n}\n"},
				want:  nil,
			},
			{
				name: "other file in a package main directory is reported",
				files: map[string]string{
					"main.go":   mainWithExit,
					"ledger.go": "package main\n\nconst limit = 3\n",
				},
				want: []string{"R14 ledger.go:1"},
			},
			{
				name:  "helper function in main.go is reported",
				files: map[string]string{"main.go": "package main\n\nfunc main() {\n\trun()\n}\n\nfunc run() {}\n"},
				want:  []string{"R14 main.go:7"},
			},
			{
				name:  "type and method in main.go are reported",
				files: map[string]string{"main.go": "package main\n\ntype ledger struct{}\n\nfunc (l ledger) total() int {\n\treturn 0\n}\n\nfunc main() {}\n"},
				want:  []string{"R14 main.go:3", "R14 main.go:5"},
			},
			{
				name:  "init in main.go is reported",
				files: map[string]string{"main.go": "package main\n\nfunc init() {}\n\nfunc main() {}\n"},
				want:  []string{"R14 main.go:3"},
			},
			{
				name:  "printing to stdout in main is reported",
				files: map[string]string{"main.go": "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"hi\")\n}\n"},
				want:  []string{"R14 main.go:6"},
			},
			{
				name:  "printing to stdout through Fprintln is reported",
				files: map[string]string{"main.go": "package main\n\nimport (\n\t\"fmt\"\n\t\"os\"\n)\n\nfunc main() {\n\tfmt.Fprintln(os.Stdout, \"hi\")\n\tfmt.Fprintln()\n}\n"},
				want:  []string{"R14 main.go:9", "R14 main.go:10"},
			},
			{
				name:  "error guard that only prints and exits passes",
				files: map[string]string{"main.go": mainWithExit},
				want:  nil,
			},
			{
				name:  "loop in main is reported",
				files: map[string]string{"main.go": "package main\n\nfunc main() {\n\tfor {\n\t}\n}\n"},
				want:  []string{"R14 main.go:4"},
			},
			{
				name:  "assignment inside a guard in main is reported",
				files: map[string]string{"main.go": "package main\n\nfunc main() {\n\terr := error(nil)\n\tif err != nil {\n\t\terr = nil\n\t}\n}\n"},
				want:  []string{"R14 main.go:4", "R14 main.go:5"},
			},
			{
				name: "example commands pass",
				files: map[string]string{
					"example/main.go":   "package main\n\nfunc main() {\n\tfor {\n\t}\n}\n",
					"example/ledger.go": "package main\n\nfunc run() {}\n",
				},
				want: nil,
			},
		},
	)
}

func TestEntrypointPackageMainSnippets(
	t *testing.T,
) {
	findings := checked(
		t,
		map[string]string{
			"cmd/app/main.go":   "package main\n\nfunc main() {\n\trun()\n}\n\nfunc run() {}\n",
			"cmd/app/ledger.go": "package main\n",
		},
	)
	assertSnippet(
		t,
		findings,
		"cmd/app/ledger.go:1",
		"file ledger.go in package main directory cmd/app: only main.go holds code in a command; move the implementation under cmd/app/internal/ and call its entry function from func main",
	)
	assertSnippet(
		t,
		findings,
		"cmd/app/main.go:7",
		"func run in cmd/app/main.go: main.go holds only func main, which only delegates to an entry function; move it under cmd/app/internal/",
	)
}

func TestEntrypointImplementationInRootCommand(
	t *testing.T,
) {
	assertFindings(
		t,
		"package main at the module root holding the implementation",
		checked(
			t,
			map[string]string{
				"main.go":   "package main\n\nimport (\n\t\"fmt\"\n\t\"os\"\n)\n\nfunc main() {\n\tledger := newLedger()\n\terr := ledger.record(os.Args[1:])\n\tif err != nil {\n\t\tfmt.Println(err)\n\t\tos.Exit(1)\n\t}\n\n\tfmt.Println(ledger.balance())\n}\n",
				"ledger.go": "package main\n\nimport (\n\t\"fmt\"\n\t\"strconv\"\n)\n\ntype ledger struct {\n\tentries []int\n}\n\nfunc newLedger() (created *ledger) {\n\treturn &ledger{}\n}\n\nfunc (l *ledger) record(\n\tvalues []string,\n) (err error) {\n\tfor _, value := range values {\n\t\terr = l.append(value)\n\t}\n\n\treturn err\n}\n\nfunc (l *ledger) append(\n\tvalue string,\n) (err error) {\n\tamount, err := strconv.Atoi(value)\n\tif err != nil {\n\t\treturn fmt.Errorf(\n\t\t\t\"parse %s: %w\",\n\t\t\tvalue,\n\t\t\terr,\n\t\t)\n\t}\n\n\tl.entries = append(\n\t\tl.entries,\n\t\tamount,\n\t)\n\treturn nil\n}\n\nfunc (l *ledger) balance() (total int) {\n\tfor _, entry := range l.entries {\n\t\ttotal += entry\n\t}\n\n\treturn total\n}\n",
			},
		),
		[]string{"R14 ledger.go:1", "R14 main.go:12", "R14 main.go:16"},
	)
}
