package publicapi_test

import "testing"

const (
	internalStore   = "package store\n\ntype Store struct{}\n\nfunc Open() {}\n"
	vocabularyStore = "package vocab\n\ntype Reservation struct{ Seats int }\n\nconst Limit = 3\n\nvar ErrFull = errFull{}\n\ntype errFull struct{}\n\nfunc (errFull) Error() string { return \"full\" }\n"
)

func TestPublicapiAliases(
	t *testing.T,
) {
	assertFixtures(
		t,
		[]fixture{
			{
				name: "alias to a type from an internal package is reported",
				files: map[string]string{
					"ledger.go":               "package ledger\n\nimport \"example.com/m/internal/store\"\n\ntype Ledger = store.Store\n",
					"internal/store/store.go": internalStore,
				},
				want: []string{"R14 ledger.go:5"},
			},
			{
				name: "pointer alias through a renamed internal import is reported",
				files: map[string]string{
					"ledger.go":               "package ledger\n\nimport impl \"example.com/m/internal/store\"\n\ntype Ledger = *impl.Store\n",
					"internal/store/store.go": internalStore,
				},
				want: []string{"R14 ledger.go:5"},
			},
			{
				name: "alias to a generic type from an internal package is reported",
				files: map[string]string{
					"ledger.go":               "package ledger\n\nimport (\n\t\"time\"\n\n\t\"example.com/m/internal/store\"\n)\n\nvar _ time.Duration\n\ntype Ledger = store.Store[int]\n",
					"internal/store/store.go": internalStore,
				},
				want: []string{"R14 ledger.go:11"},
			},
			{
				name:  "alias to a struct literal is reported",
				files: map[string]string{"ledger.go": "package ledger\n\ntype Ledger = struct{ Owner string }\n"},
				want:  []string{"R14 ledger.go:3"},
			},
			{
				name: "alias to a struct of the same package is reported",
				files: map[string]string{
					"entry.go":  "package ledger\n\ntype Entry struct{}\n",
					"ledger.go": "package ledger\n\ntype Ledger = Entry\n",
				},
				want: []string{"R14 ledger.go:3"},
			},
			{
				name:  "alias to a public package type passes",
				files: map[string]string{"ledger.go": "package ledger\n\nimport \"time\"\n\ntype Duration = time.Duration\n"},
				want:  nil,
			},
			{
				name:  "alias to a generic public type passes",
				files: map[string]string{"ledger.go": "package ledger\n\nimport \"example.com/m/kv\"\n\ntype Map = kv.Map[string, int]\n"},
				want:  nil,
			},
			{
				name:  "alias to an interface of the same package passes",
				files: map[string]string{"ledger.go": "package ledger\n\ntype Reader interface{}\n\ntype Source = Reader\n"},
				want:  nil,
			},
			{
				name:  "alias qualified by a name that is not imported passes",
				files: map[string]string{"ledger.go": "package ledger\n\ntype Source = a.B\n"},
				want:  nil,
			},
			{
				name: "alias to a declaration-only internal package passes",
				files: map[string]string{
					"ledger.go":               "package ledger\n\nimport \"example.com/m/internal/vocab\"\n\ntype Reservation = vocab.Reservation\n\nconst Limit = vocab.Limit\n",
					"internal/vocab/vocab.go": vocabularyStore,
				},
				want: nil,
			},
			{
				name: "alias to an internal package holding package state is reported",
				files: map[string]string{
					"ledger.go":               "package ledger\n\nimport \"example.com/m/internal/vocab\"\n\ntype Reservation = vocab.Reservation\n",
					"internal/vocab/vocab.go": "package vocab\n\ntype Reservation struct{}\n\nvar cache = map[int]int{}\n",
				},
				want: []string{"R14 ledger.go:5"},
			},
			{
				name: "struct alias inside an internal package that is not a module passes",
				files: map[string]string{
					"internal/ledger/ledger.go": "package ledger\n\ntype Ledger = struct{}\n",
				},
				want: nil,
			},
		},
	)
}

func TestPublicapiAliasSnippet(
	t *testing.T,
) {
	assertSnippet(
		t,
		checked(
			t,
			map[string]string{
				"ledger.go":               "package ledger\n\nimport \"example.com/m/internal/store\"\n\ntype Ledger = store.Store\n",
				"internal/store/store.go": internalStore,
			},
		),
		"ledger.go:5",
		"type alias Ledger exposes store.Store: declare an interface for what callers need",
	)
}
