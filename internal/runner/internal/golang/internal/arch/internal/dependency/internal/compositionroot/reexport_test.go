package compositionroot_test

import "testing"

const (
	vocabularyPackage = "package vocab\n\ntype Reservation struct{ Seats int }\n\nconst Limit = 3\n"
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
