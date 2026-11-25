package compositionroot_test

import "testing"

const (
	vocabularyPackage = "package vocab\n\ntype Reservation struct{ Seats int }\n\nconst Limit = 3\n"
	sentinelPackage   = "package vocab\n\nimport \"errors\"\n\nconst Limit = 3\n\ntype Box struct{ ErrShut error }\n\nvar (\n\tErrFull = errors.New(\"full\")\n\tErrShut = errors.New(\"shut\")\n)\n"
	reservationAlias  = "package ratelimit\n\nimport \"example.com/m/internal/vocab\"\n\ntype Reservation = vocab.Reservation\n\nconst Limit = vocab.Limit\n"
)

func TestCompositionrootReExports(
	t *testing.T,
) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{
			name: "API file aliasing a declaration-only internal package passes",
			files: map[string]string{
				"container.go":            unit("ratelimit internal/limiter"),
				"reservation.go":          reservationAlias,
				"internal/vocab/vocab.go": vocabularyPackage,
				"internal/limiter/l.go":   unit("limiter"),
			},
			want: nil,
		},
		{
			name: "renamed vocabulary import used only in aliases passes",
			files: map[string]string{
				"container.go":            unit("ratelimit internal/limiter"),
				"reservation.go":          "package ratelimit\n\nimport words \"example.com/m/internal/vocab\"\n\ntype (\n\tReservation = words.Reservation\n\tSeats       = int\n)\n",
				"internal/vocab/vocab.go": vocabularyPackage,
				"internal/limiter/l.go":   unit("limiter"),
			},
			want: nil,
		},
		{
			name: "vocabulary used outside an alias is reported",
			files: map[string]string{
				"container.go":            unit("ratelimit internal/limiter"),
				"reservation.go":          "package ratelimit\n\nimport \"example.com/m/internal/vocab\"\n\ntype Reservation = vocab.Reservation\n\ntype Ledger struct{ Last vocab.Reservation }\n",
				"internal/vocab/vocab.go": vocabularyPackage,
				"internal/limiter/l.go":   unit("limiter"),
			},
			want: []string{"R48 reservation.go:3"},
		},
		{
			name: "vocabulary in a var is reported",
			files: map[string]string{
				"container.go":            unit("ratelimit internal/limiter"),
				"reservation.go":          "package ratelimit\n\nimport \"example.com/m/internal/vocab\"\n\nvar Limit = vocab.Limit\n",
				"internal/vocab/vocab.go": vocabularyPackage,
				"internal/limiter/l.go":   unit("limiter"),
			},
			want: []string{"R48 reservation.go:3"},
		},
		{
			name: "Err sentinels re-exported from a vocabulary package pass",
			files: map[string]string{
				"container.go":            unit("ratelimit internal/limiter"),
				"errors.go":               "package ratelimit\n\nimport \"example.com/m/internal/vocab\"\n\nvar (\n\tErrFull = vocab.ErrFull\n\tErrShut = vocab.ErrShut\n)\n",
				"internal/vocab/vocab.go": sentinelPackage,
				"internal/limiter/l.go":   unit("limiter"),
			},
			want: nil,
		},
		{
			name: "var not named Err re-exporting a sentinel is reported",
			files: map[string]string{
				"container.go":            unit("ratelimit internal/limiter"),
				"errors.go":               "package ratelimit\n\nimport \"example.com/m/internal/vocab\"\n\nvar Full = vocab.ErrFull\n",
				"internal/vocab/vocab.go": sentinelPackage,
				"internal/limiter/l.go":   unit("limiter"),
			},
			want: []string{"R48 errors.go:3"},
		},
		{
			name: "Err var wrapping a vocabulary sentinel is reported",
			files: map[string]string{
				"container.go":            unit("ratelimit internal/limiter"),
				"errors.go":               "package ratelimit\n\nimport (\n\t\"fmt\"\n\n\t\"example.com/m/internal/vocab\"\n)\n\nvar ErrFull = fmt.Errorf(\"full: %w\", vocab.ErrFull)\n",
				"internal/vocab/vocab.go": sentinelPackage,
				"internal/limiter/l.go":   unit("limiter"),
			},
			want: []string{"R48 errors.go:6"},
		},
		{
			name: "Err var re-exporting a non-sentinel value is reported",
			files: map[string]string{
				"container.go":            unit("ratelimit internal/limiter"),
				"errors.go":               "package ratelimit\n\nimport \"example.com/m/internal/vocab\"\n\nvar ErrLimit = vocab.Limit\n",
				"internal/vocab/vocab.go": sentinelPackage,
				"internal/limiter/l.go":   unit("limiter"),
			},
			want: []string{"R48 errors.go:3"},
		},
		{
			name: "Err var without a value or with a nested selector is reported",
			files: map[string]string{
				"container.go":            unit("ratelimit internal/limiter"),
				"errors.go":               "package ratelimit\n\nimport \"example.com/m/internal/vocab\"\n\nvar ErrNone, ErrBox = vocab.ErrFull, vocab.Box.ErrShut\n",
				"internal/vocab/vocab.go": sentinelPackage,
				"internal/limiter/l.go":   unit("limiter"),
			},
			want: []string{"R48 errors.go:3"},
		},
		{
			name: "alias to an internal package with behavior is reported",
			files: map[string]string{
				"container.go":          unit("ratelimit"),
				"reservation.go":        "package ratelimit\n\nimport \"example.com/m/internal/limiter\"\n\ntype Limiter = limiter.Limiter\n",
				"internal/limiter/l.go": "package limiter\n\ntype Limiter struct{}\n\nfunc New() Limiter {\n\treturn Limiter{}\n}\n",
			},
			want: []string{"R48 reservation.go:3"},
		},
		{
			name: "blank import of a vocabulary package is reported",
			files: map[string]string{
				"container.go":            unit("ratelimit internal/limiter"),
				"reservation.go":          "package ratelimit\n\nimport _ \"example.com/m/internal/vocab\"\n",
				"internal/vocab/vocab.go": vocabularyPackage,
				"internal/limiter/l.go":   unit("limiter"),
			},
			want: []string{"R48 reservation.go:3"},
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
