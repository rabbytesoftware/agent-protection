package sqllex_test

import (
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/commentscan/internal/sqllex"
	"testing"
)

func TestMigrationMarkers(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []int
	}{
		{name: "goose up and down pass", code: "-- +goose Up\nCREATE TABLE t (a INT);\n-- +goose Down\nDROP TABLE t;\n", want: nil},
		{name: "goose statement markers pass", code: "-- +goose StatementBegin\nSELECT 1;\n-- +goose StatementEnd\n", want: nil},
		{name: "goose no transaction passes", code: "-- +goose NO TRANSACTION\n", want: nil},
		{name: "goose marker with prose is reported", code: "-- +goose Up because the index is wrong\n", want: []int{1}},
		{name: "migrate markers pass", code: "-- +migrate Up\n-- +migrate Down notransaction\n", want: nil},
		{name: "sqlc annotation passes", code: "-- name: GetUser :one\nSELECT 1;\n", want: nil},
		{name: "sqlc annotation with prose is reported", code: "-- name: GetUser :one returns the user\n", want: []int{1}},
		{name: "trailing marker is reported", code: "SELECT 1; -- +goose Up\n", want: []int{1}},
	}
	for _, tc := range cases {
		assertLines(
			t,
			tc.name,
			sqllex.Sqllex(),
			tc.code,
			tc.want,
		)
	}
}
