package interfacesize_test

import (
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/syntax/internal/interfacesize"
)

const (
	lifecycle = "package role\n\ntype Lifecycle interface {\n\tStart()\n\tStop()\n}\n"
	producer  = "package role\n\ntype Producer interface {\n\tEnqueue()\n\tCancel()\n}\n"
	advice    = "(max 3): a struct that embeds interfaces is one large interface; return each small interface from its own constructor and let every consumer depend on the one it calls"
)

func TestInterfacesizeSurface(
	t *testing.T,
) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{
			name:  "struct embedding small interfaces from another package is reported",
			files: map[string]string{"contracts/role/lifecycle.go": lifecycle, "contracts/role/producer.go": producer, "queue.go": "package jobq\n\nimport \"example.com/jobq/contracts/role\"\n\ntype Queue struct {\n\trole.Lifecycle\n\trole.Producer\n}\n"},
			want:  []string{"R28 queue.go:5 struct Queue exposes 4 methods through embedded role.Lifecycle, role.Producer " + advice},
		},
		{
			name:  "own exported methods add to the promoted ones",
			files: map[string]string{"internal/cube/reader.go": "package cube\n\ntype Reader interface {\n\tLoad()\n\tKeys()\n}\n", "internal/cube/box.go": "package cube\n\ntype Box struct {\n\tReader\n}\n\nfunc (b *Box) Store() {}\n\nfunc (b Box) Reset() {}\n\nfunc (b Box) hidden() {}\n"},
			want:  []string{"R28 internal/cube/box.go:3 struct Box exposes 4 methods through embedded Reader plus 2 own methods " + advice},
		},
		{
			name:  "decorator overriding a method of its one embedded interface adds no surface",
			files: map[string]string{"internal/cube/big.go": "package cube\n\ntype Big interface {\n\tA()\n\tB()\n\tC()\n\tD()\n}\n", "internal/cube/traced.go": "package cube\n\ntype traced struct {\n\tBig\n}\n\nfunc (t traced) A() {}\n"},
			want:  nil,
		},
		{
			name:  "struct with only own methods is left alone",
			files: map[string]string{"internal/cube/plain.go": "package cube\n\ntype Plain struct{}\n\nfunc (Plain) A() {}\n\nfunc (Plain) B() {}\n\nfunc (Plain) C() {}\n\nfunc (Plain) D() {}\n"},
			want:  nil,
		},
		{
			name:  "embedding concrete structs without interfaces is left alone",
			files: map[string]string{"internal/cube/base.go": "package cube\n\ntype Base struct{}\n\nfunc (Base) A() {}\n\nfunc (Base) B() {}\n", "internal/cube/full.go": "package cube\n\ntype Full struct {\n\t*Base\n}\n\nfunc (Full) C() {}\n\nfunc (Full) D() {}\n"},
			want:  nil,
		},
		{
			name: "interfaces promoted through an embedded struct are counted",
			files: map[string]string{
				"internal/role/lifecycle.go": lifecycle,
				"internal/role/producer.go":  producer,
				"internal/role/half.go":      "package role\n\ntype half struct {\n\tLifecycle\n}\n",
				"internal/role/whole.go":     "package role\n\ntype Whole[T any] struct {\n\t*half\n\tProducer\n}\n",
			},
			want: []string{"R28 internal/role/whole.go:3 struct Whole exposes 4 methods through embedded *half, Producer " + advice},
		},
		{
			name:  "named fields promote nothing",
			files: map[string]string{"internal/role/lifecycle.go": lifecycle, "internal/role/producer.go": producer, "internal/role/queue.go": "package role\n\ntype Queue struct {\n\tlife Lifecycle\n\tProducer\n}\n"},
			want:  nil,
		},
		{
			name:  "fakes in test files embedding interfaces are allowed",
			files: map[string]string{"internal/role/lifecycle.go": lifecycle, "internal/role/producer.go": producer, "internal/role/fake_test.go": "package role\n\ntype fake struct {\n\tLifecycle\n\tProducer\n}\n"},
			want:  nil,
		},
		{
			name:  "structs in mock files are allowed",
			files: map[string]string{"internal/role/lifecycle.go": lifecycle, "internal/role/producer.go": producer, "internal/role/mocks/queue.go": "package mocks\n\nimport \"example.com/jobq/internal/role\"\n\ntype Queue struct {\n\trole.Lifecycle\n\trole.Producer\n}\n"},
			want:  nil,
		},
		{
			name:  "cyclic struct embedding terminates",
			files: map[string]string{"internal/role/lifecycle.go": lifecycle, "internal/role/ping.go": "package role\n\ntype ping struct {\n\t*pong\n\tLifecycle\n}\n\ntype pong struct {\n\t*ping\n}\n"},
			want:  nil,
		},
	}
	for _, tc := range cases {
		assertFindings(
			t,
			tc.name,
			interfacesize.Interfacesize(parsedFiles(
				t,
				tc.files,
			)),
			tc.want,
		)
	}
}
