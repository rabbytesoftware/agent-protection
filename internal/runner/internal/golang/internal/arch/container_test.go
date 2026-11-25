package arch_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/arch"
	"github.com/rabbytesoftware/agent-protection/internal/runner/internal/golang/internal/source"
)

func TestArch(
	t *testing.T,
) {
	cases := []struct {
		name  string
		files map[string]string
		fails bool
		want  []string
	}{
		{
			name: "every child check contributes its findings in check order",
			files: map[string]string{
				"go.mod":                  "module example.com/m\n\ngo 1.24\n",
				"internal/a/service.go":   "package a\n\nimport \"example.com/m/internal/b\"\n\nfunc A() {\n\tb.B()\n}\n",
				"internal/a/misc_test.go": "package a_test\n",
				"internal/b/b.go":         "package b\n\nimport \"example.com/m/internal/c\"\n\nfunc B() {\n\tc.C()\n}\n",
				"internal/b/b_test.go":    "package b_test\n\nfunc TestB() {}\n",
				"internal/c/c.go":         "package c\n\nfunc C() {}\n",
				"internal/c/c_test.go":    "package c_test\n\nfunc TestC() {}\n",
			},
			want: []string{
				"R42 internal/a/service.go",
				"R43 internal/a/misc_test.go",
				"R16 internal/a/service.go:3",
				"R19 internal/b",
				"R19 internal/c",
			},
		},
		{
			name: "module without go.mod fails instead of reporting partial findings",
			files: map[string]string{
				"internal/a/a.go": "package a\n",
			},
			fails: true,
			want:  nil,
		},
	}
	for _, tc := range cases {
		assertRules(
			t,
			tc.name,
			tc.files,
			tc.fails,
			tc.want,
		)
	}
}

func TestArchEncapsulationAdviceConverges(
	t *testing.T,
) {
	cases := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{
			name: "leaf importer: R19 asks to merge the package into it",
			files: map[string]string{
				"calculator.go":                      "package arith\n\ntype Calculator interface {\n\tSquare(\n\t\tvalue int,\n\t) (result int)\n\tCube(\n\t\tvalue int,\n\t) (result int)\n}\n",
				"calculator_test.go":                 "package arith\n\nimport \"testing\"\n\nfunc TestCalculator(\n\tt *testing.T,\n) {\n\tvar calculator Calculator = Arith()\n\tif calculator == nil {\n\t\tt.Fatal(\"nil\")\n\t}\n}\n",
				"container.go":                       "package arith\n\nimport (\n\t\"example.com/arith/internal/cube\"\n\t\"example.com/arith/internal/square\"\n)\n\nfunc Arith() (calculator Calculator) {\n\treturn cube.Cube(square.Square())\n}\n",
				"container_test.go":                  "package arith\n\nimport \"testing\"\n\nfunc TestArith(\n\tt *testing.T,\n) {\n\tif Arith().Cube(2) != 8 {\n\t\tt.Fatal(\"cube\")\n\t}\n}\n",
				"go.mod":                             "module example.com/arith\n\ngo 1.24\n",
				"internal/cube/cube.go":              "package cube\n\ntype cube struct {\n\tsquarer Squarer\n}\n\nfunc Cube(\n\tsquarer Squarer,\n) (calculator *cube) {\n\treturn &cube{squarer: squarer}\n}\n\nfunc (c *cube) Square(\n\tvalue int,\n) (result int) {\n\treturn c.squarer.Square(value)\n}\n\nfunc (c *cube) Cube(\n\tvalue int,\n) (result int) {\n\treturn c.squarer.Square(value) * value\n}\n",
				"internal/cube/cube_test.go":         "package cube\n\nimport \"testing\"\n\ntype identity struct{}\n\nfunc (identity) Square(\n\tvalue int,\n) (result int) {\n\treturn value * value\n}\n\nfunc TestCube(\n\tt *testing.T,\n) {\n\tif Cube(identity{}).Cube(2) != 8 {\n\t\tt.Fatal(\"cube\")\n\t}\n}\n",
				"internal/cube/squarer.go":           "package cube\n\ntype Squarer interface {\n\tSquare(\n\t\tvalue int,\n\t) (result int)\n}\n",
				"internal/cube/squarer_test.go":      "package cube\n\nimport \"testing\"\n\nfunc TestSquarer(\n\tt *testing.T,\n) {\n\tvar squarer Squarer = Cube(nil)\n\tif squarer == nil {\n\t\tt.Fatal(\"nil\")\n\t}\n}\n",
				"internal/rounding/rounding.go":      "package rounding\n\nconst step = 1\n\nfunc Rounding(\n\tvalue int,\n) (rounded int) {\n\treturn value / step * step\n}\n",
				"internal/rounding/rounding_test.go": "package rounding\n\nimport \"testing\"\n\nfunc TestRounding(\n\tt *testing.T,\n) {\n\tif Rounding(19) != 19 {\n\t\tt.Fatal(\"rounding\")\n\t}\n}\n",
				"internal/square/square.go":          "package square\n\nimport \"example.com/arith/internal/rounding\"\n\ntype square struct {\n\tfactor int\n}\n\nfunc Square() (squarer *square) {\n\treturn &square{factor: 1}\n}\n\nfunc (s *square) Square(\n\tvalue int,\n) (result int) {\n\treturn rounding.Rounding(value * value * s.factor)\n}\n",
				"internal/square/square_test.go":     "package square\n\nimport \"testing\"\n\nfunc TestSquare(\n\tt *testing.T,\n) {\n\tif Square().Square(3) != 9 {\n\t\tt.Fatal(\"square\")\n\t}\n}\n",
			},
			want: []string{
				"R19 internal/rounding",
			},
		},
		{
			name: "leaf importer after the merge passes",
			files: map[string]string{
				"calculator.go":                    "package arith\n\ntype Calculator interface {\n\tSquare(\n\t\tvalue int,\n\t) (result int)\n\tCube(\n\t\tvalue int,\n\t) (result int)\n}\n",
				"calculator_test.go":               "package arith\n\nimport \"testing\"\n\nfunc TestCalculator(\n\tt *testing.T,\n) {\n\tvar calculator Calculator = Arith()\n\tif calculator == nil {\n\t\tt.Fatal(\"nil\")\n\t}\n}\n",
				"container.go":                     "package arith\n\nimport (\n\t\"example.com/arith/internal/cube\"\n\t\"example.com/arith/internal/square\"\n)\n\nfunc Arith() (calculator Calculator) {\n\treturn cube.Cube(square.Square())\n}\n",
				"container_test.go":                "package arith\n\nimport \"testing\"\n\nfunc TestArith(\n\tt *testing.T,\n) {\n\tif Arith().Cube(2) != 8 {\n\t\tt.Fatal(\"cube\")\n\t}\n}\n",
				"go.mod":                           "module example.com/arith\n\ngo 1.24\n",
				"internal/cube/cube.go":            "package cube\n\ntype cube struct {\n\tsquarer Squarer\n}\n\nfunc Cube(\n\tsquarer Squarer,\n) (calculator *cube) {\n\treturn &cube{squarer: squarer}\n}\n\nfunc (c *cube) Square(\n\tvalue int,\n) (result int) {\n\treturn c.squarer.Square(value)\n}\n\nfunc (c *cube) Cube(\n\tvalue int,\n) (result int) {\n\treturn c.squarer.Square(value) * value\n}\n",
				"internal/cube/cube_test.go":       "package cube\n\nimport \"testing\"\n\ntype identity struct{}\n\nfunc (identity) Square(\n\tvalue int,\n) (result int) {\n\treturn value * value\n}\n\nfunc TestCube(\n\tt *testing.T,\n) {\n\tif Cube(identity{}).Cube(2) != 8 {\n\t\tt.Fatal(\"cube\")\n\t}\n}\n",
				"internal/cube/squarer.go":         "package cube\n\ntype Squarer interface {\n\tSquare(\n\t\tvalue int,\n\t) (result int)\n}\n",
				"internal/cube/squarer_test.go":    "package cube\n\nimport \"testing\"\n\nfunc TestSquarer(\n\tt *testing.T,\n) {\n\tvar squarer Squarer = Cube(nil)\n\tif squarer == nil {\n\t\tt.Fatal(\"nil\")\n\t}\n}\n",
				"internal/square/rounding.go":      "package square\n\nconst step = 1\n\nfunc rounding(\n\tvalue int,\n) (rounded int) {\n\treturn value / step * step\n}\n",
				"internal/square/rounding_test.go": "package square\n\nimport \"testing\"\n\nfunc TestRounding(\n\tt *testing.T,\n) {\n\tif rounding(19) != 19 {\n\t\tt.Fatal(\"rounding\")\n\t}\n}\n",
				"internal/square/square.go":        "package square\n\ntype square struct {\n\tfactor int\n}\n\nfunc Square() (squarer *square) {\n\treturn &square{factor: 1}\n}\n\nfunc (s *square) Square(\n\tvalue int,\n) (result int) {\n\treturn rounding(value * value * s.factor)\n}\n",
				"internal/square/square_test.go":   "package square\n\nimport \"testing\"\n\nfunc TestSquare(\n\tt *testing.T,\n) {\n\tif Square().Square(3) != 9 {\n\t\tt.Fatal(\"square\")\n\t}\n}\n",
			},
			want: nil,
		},
		{
			name: "module importer: R19 asks to nest the package under its internal/",
			files: map[string]string{
				"calculator.go":                                "package arith\n\ntype Calculator interface {\n\tSquare(\n\t\tvalue int,\n\t) (result int)\n\tCube(\n\t\tvalue int,\n\t) (result int)\n}\n",
				"calculator_test.go":                           "package arith\n\nimport \"testing\"\n\nfunc TestCalculator(\n\tt *testing.T,\n) {\n\tvar calculator Calculator = Arith()\n\tif calculator == nil {\n\t\tt.Fatal(\"nil\")\n\t}\n}\n",
				"container.go":                                 "package arith\n\nimport (\n\t\"example.com/arith/internal/cube\"\n\t\"example.com/arith/internal/square\"\n)\n\nfunc Arith() (calculator Calculator) {\n\treturn cube.Cube(square.Square())\n}\n",
				"container_test.go":                            "package arith\n\nimport \"testing\"\n\nfunc TestArith(\n\tt *testing.T,\n) {\n\tif Arith().Cube(2) != 8 {\n\t\tt.Fatal(\"cube\")\n\t}\n}\n",
				"go.mod":                                       "module example.com/arith\n\ngo 1.24\n",
				"internal/cube/cube.go":                        "package cube\n\ntype cube struct {\n\tsquarer Squarer\n}\n\nfunc Cube(\n\tsquarer Squarer,\n) (calculator *cube) {\n\treturn &cube{squarer: squarer}\n}\n\nfunc (c *cube) Square(\n\tvalue int,\n) (result int) {\n\treturn c.squarer.Square(value)\n}\n\nfunc (c *cube) Cube(\n\tvalue int,\n) (result int) {\n\treturn c.squarer.Square(value) * value\n}\n",
				"internal/cube/cube_test.go":                   "package cube\n\nimport \"testing\"\n\ntype identity struct{}\n\nfunc (identity) Square(\n\tvalue int,\n) (result int) {\n\treturn value * value\n}\n\nfunc TestCube(\n\tt *testing.T,\n) {\n\tif Cube(identity{}).Cube(2) != 8 {\n\t\tt.Fatal(\"cube\")\n\t}\n}\n",
				"internal/cube/squarer.go":                     "package cube\n\ntype Squarer interface {\n\tSquare(\n\t\tvalue int,\n\t) (result int)\n}\n",
				"internal/cube/squarer_test.go":                "package cube\n\nimport \"testing\"\n\nfunc TestSquarer(\n\tt *testing.T,\n) {\n\tvar squarer Squarer = Cube(nil)\n\tif squarer == nil {\n\t\tt.Fatal(\"nil\")\n\t}\n}\n",
				"internal/rounding/rounding.go":                "package rounding\n\nconst step = 1\n\nfunc Rounding(\n\tvalue int,\n) (rounded int) {\n\treturn value / step * step\n}\n",
				"internal/rounding/rounding_test.go":           "package rounding\n\nimport \"testing\"\n\nfunc TestRounding(\n\tt *testing.T,\n) {\n\tif Rounding(19) != 19 {\n\t\tt.Fatal(\"rounding\")\n\t}\n}\n",
				"internal/square/container.go":                 "package square\n\nimport (\n\t\"example.com/arith/internal/rounding\"\n\t\"example.com/arith/internal/square/internal/power\"\n)\n\nfunc Square() (raiser Raiser) {\n\treturn power.Power(rounding.Rounding)\n}\n",
				"internal/square/container_test.go":            "package square\n\nimport \"testing\"\n\nfunc TestSquare(\n\tt *testing.T,\n) {\n\tif Square().Square(4) != 16 {\n\t\tt.Fatal(\"square\")\n\t}\n}\n",
				"internal/square/internal/power/power.go":      "package power\n\ntype power struct {\n\tround func(int) int\n}\n\nfunc Power(\n\tround func(int) int,\n) (squarer *power) {\n\treturn &power{round: round}\n}\n\nfunc (p *power) Square(\n\tvalue int,\n) (result int) {\n\treturn p.round(value * value)\n}\n",
				"internal/square/internal/power/power_test.go": "package power\n\nimport \"testing\"\n\nfunc identity(\n\tvalue int,\n) (result int) {\n\treturn value\n}\n\nfunc TestPower(\n\tt *testing.T,\n) {\n\tif Power(identity).Square(3) != 9 {\n\t\tt.Fatal(\"power\")\n\t}\n}\n",
				"internal/square/raiser.go":                    "package square\n\ntype Raiser interface {\n\tSquare(\n\t\tvalue int,\n\t) (result int)\n}\n",
				"internal/square/raiser_test.go":               "package square\n\nimport \"testing\"\n\nfunc TestRaiser(\n\tt *testing.T,\n) {\n\tvar raiser Raiser = Square()\n\tif raiser == nil {\n\t\tt.Fatal(\"nil\")\n\t}\n}\n",
			},
			want: []string{
				"R51 internal/square/internal/power",
				"R19 internal/rounding",
			},
		},
		{
			name: "module importer after nesting passes",
			files: map[string]string{
				"calculator.go":                                      "package arith\n\ntype Calculator interface {\n\tSquare(\n\t\tvalue int,\n\t) (result int)\n\tCube(\n\t\tvalue int,\n\t) (result int)\n}\n",
				"calculator_test.go":                                 "package arith\n\nimport \"testing\"\n\nfunc TestCalculator(\n\tt *testing.T,\n) {\n\tvar calculator Calculator = Arith()\n\tif calculator == nil {\n\t\tt.Fatal(\"nil\")\n\t}\n}\n",
				"container.go":                                       "package arith\n\nimport (\n\t\"example.com/arith/internal/cube\"\n\t\"example.com/arith/internal/square\"\n)\n\nfunc Arith() (calculator Calculator) {\n\treturn cube.Cube(square.Square())\n}\n",
				"container_test.go":                                  "package arith\n\nimport \"testing\"\n\nfunc TestArith(\n\tt *testing.T,\n) {\n\tif Arith().Cube(2) != 8 {\n\t\tt.Fatal(\"cube\")\n\t}\n}\n",
				"go.mod":                                             "module example.com/arith\n\ngo 1.24\n",
				"internal/cube/cube.go":                              "package cube\n\ntype cube struct {\n\tsquarer Squarer\n}\n\nfunc Cube(\n\tsquarer Squarer,\n) (calculator *cube) {\n\treturn &cube{squarer: squarer}\n}\n\nfunc (c *cube) Square(\n\tvalue int,\n) (result int) {\n\treturn c.squarer.Square(value)\n}\n\nfunc (c *cube) Cube(\n\tvalue int,\n) (result int) {\n\treturn c.squarer.Square(value) * value\n}\n",
				"internal/cube/cube_test.go":                         "package cube\n\nimport \"testing\"\n\ntype identity struct{}\n\nfunc (identity) Square(\n\tvalue int,\n) (result int) {\n\treturn value * value\n}\n\nfunc TestCube(\n\tt *testing.T,\n) {\n\tif Cube(identity{}).Cube(2) != 8 {\n\t\tt.Fatal(\"cube\")\n\t}\n}\n",
				"internal/cube/squarer.go":                           "package cube\n\ntype Squarer interface {\n\tSquare(\n\t\tvalue int,\n\t) (result int)\n}\n",
				"internal/cube/squarer_test.go":                      "package cube\n\nimport \"testing\"\n\nfunc TestSquarer(\n\tt *testing.T,\n) {\n\tvar squarer Squarer = Cube(nil)\n\tif squarer == nil {\n\t\tt.Fatal(\"nil\")\n\t}\n}\n",
				"internal/square/container.go":                       "package square\n\nimport (\n\t\"example.com/arith/internal/square/internal/power\"\n\t\"example.com/arith/internal/square/internal/rounding\"\n)\n\nfunc Square() (raiser Raiser) {\n\treturn power.Power(rounding.Rounding)\n}\n",
				"internal/square/container_test.go":                  "package square\n\nimport \"testing\"\n\nfunc TestSquare(\n\tt *testing.T,\n) {\n\tif Square().Square(4) != 16 {\n\t\tt.Fatal(\"square\")\n\t}\n}\n",
				"internal/square/internal/power/power.go":            "package power\n\ntype power struct {\n\tround func(int) int\n}\n\nfunc Power(\n\tround func(int) int,\n) (squarer *power) {\n\treturn &power{round: round}\n}\n\nfunc (p *power) Square(\n\tvalue int,\n) (result int) {\n\treturn p.round(value * value)\n}\n",
				"internal/square/internal/power/power_test.go":       "package power\n\nimport \"testing\"\n\nfunc identity(\n\tvalue int,\n) (result int) {\n\treturn value\n}\n\nfunc TestPower(\n\tt *testing.T,\n) {\n\tif Power(identity).Square(3) != 9 {\n\t\tt.Fatal(\"power\")\n\t}\n}\n",
				"internal/square/internal/rounding/rounding.go":      "package rounding\n\nconst step = 1\n\nfunc Rounding(\n\tvalue int,\n) (rounded int) {\n\treturn value / step * step\n}\n",
				"internal/square/internal/rounding/rounding_test.go": "package rounding\n\nimport \"testing\"\n\nfunc TestRounding(\n\tt *testing.T,\n) {\n\tif Rounding(19) != 19 {\n\t\tt.Fatal(\"rounding\")\n\t}\n}\n",
				"internal/square/raiser.go":                          "package square\n\ntype Raiser interface {\n\tSquare(\n\t\tvalue int,\n\t) (result int)\n}\n",
				"internal/square/raiser_test.go":                     "package square\n\nimport \"testing\"\n\nfunc TestRaiser(\n\tt *testing.T,\n) {\n\tvar raiser Raiser = Square()\n\tif raiser == nil {\n\t\tt.Fatal(\"nil\")\n\t}\n}\n",
			},
			want: nil,
		},
	}
	for _, tc := range cases {
		assertRules(
			t,
			tc.name,
			tc.files,
			false,
			tc.want,
		)
	}
}

func assertRules(
	t *testing.T,
	name string,
	files map[string]string,
	fails bool,
	want []string,
) {
	t.Helper()
	root := module(
		t,
		files,
	)
	parsed, err := source.Source(root)
	if err != nil {
		t.Fatalf(
			"%s: parse fixture module: %v",
			name,
			err,
		)
	}

	findings, err := arch.Arch(
		root,
		parsed,
	)
	hasFailed := err != nil
	if hasFailed != fails {
		t.Fatalf(
			"%s: Arch() error = %v, want failure %v",
			name,
			err,
			fails,
		)
	}

	got := []string{}
	for _, finding := range findings {
		got = append(
			got,
			finding.Rule+" "+finding.Location,
		)
	}

	if !slices.Equal(
		got,
		want,
	) {
		t.Errorf(
			"%s: findings = %v, want %v",
			name,
			got,
			want,
		)
	}
}

func module(
	t *testing.T,
	files map[string]string,
) (root string) {
	t.Helper()
	root = t.TempDir()
	for name, content := range files {
		write(
			t,
			filepath.Join(
				root,
				name,
			),
			content,
		)
	}

	return root
}

func write(
	t *testing.T,
	path string,
	content string,
) {
	t.Helper()
	err := os.MkdirAll(
		filepath.Dir(path),
		0o755,
	)
	if err != nil {
		t.Fatalf(
			"mkdir %s: %v",
			path,
			err,
		)
	}

	err = os.WriteFile(
		path,
		[]byte(content),
		0o600,
	)
	if err != nil {
		t.Fatalf(
			"write %s: %v",
			path,
			err,
		)
	}
}
