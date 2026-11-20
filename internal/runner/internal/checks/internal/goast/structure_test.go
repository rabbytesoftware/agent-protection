package goast

import "testing"

func TestPackageHygiene(t *testing.T) {
	testCases := []struct {
		name string
		src  string
		want int
	}{
		{name: "banned generic package name is flagged", src: "package util\nfunc A() {}\n", want: 1},
		{name: "ordinary lowercase name is clean", src: "package domain\nfunc A() {}\n", want: 0},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			files := parseOne(t, dir, "f.go", tc.src)
			got := PackageHygiene(dir, files)
			if len(got) != tc.want {
				t.Errorf("PackageHygiene() = %d, want %d: %v", len(got), tc.want, got)
			}
		})
	}
}

func TestNoStutter(t *testing.T) {
	testCases := []struct {
		name string
		src  string
		want int
	}{
		{name: "stuttering type name is flagged", src: "package user\ntype UserService struct{}\n", want: 1},
		{name: "exact package-name match is the one allowed exception", src: "package vault\ntype Vault struct{}\n", want: 0},
		{name: "non-stuttering name is clean", src: "package user\ntype Service struct{}\n", want: 0},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			files := parseOne(t, dir, "f.go", tc.src)
			got := NoStutter(dir, files)
			if len(got) != tc.want {
				t.Errorf("NoStutter() = %d, want %d: %v", len(got), tc.want, got)
			}
		})
	}
}

func TestOneTypePerFile(t *testing.T) {
	testCases := []struct {
		name string
		src  string
		want int
	}{
		{name: "two types in one file is flagged", src: "package x\ntype A struct{}\ntype B struct{}\n", want: 1},
		{name: "one type is clean", src: "package x\ntype A struct{}\n", want: 0},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			files := parseOne(t, dir, "f.go", tc.src)
			got := OneTypePerFile(dir, files)
			if len(got) != tc.want {
				t.Errorf("OneTypePerFile() = %d, want %d: %v", len(got), tc.want, got)
			}
		})
	}
}

func TestNoElse(t *testing.T) {
	testCases := []struct {
		name string
		src  string
		want int
	}{
		{name: "else branch is flagged", src: "package x\nfunc F(a bool) int {\n\tif a {\n\t\treturn 1\n\t} else {\n\t\treturn 2\n\t}\n}\n", want: 1},
		{name: "guard clause with no else is clean", src: "package x\nfunc F(a bool) int {\n\tif a {\n\t\treturn 1\n\t}\n\treturn 2\n}\n", want: 0},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			files := parseOne(t, dir, "f.go", tc.src)
			got := NoElse(dir, files)
			if len(got) != tc.want {
				t.Errorf("NoElse() = %d, want %d: %v", len(got), tc.want, got)
			}
		})
	}
}

func TestNakedReturn(t *testing.T) {
	testCases := []struct {
		name string
		src  string
		want int
	}{
		{name: "naked return on named results is flagged", src: "package x\nfunc F() (result int, err error) {\n\treturn\n}\n", want: 1},
		{name: "explicit return on named results is clean", src: "package x\nfunc F() (result int, err error) {\n\treturn 1, nil\n}\n", want: 0},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			files := parseOne(t, dir, "f.go", tc.src)
			got := NakedReturn(dir, files)
			if len(got) != tc.want {
				t.Errorf("NakedReturn() = %d, want %d: %v", len(got), tc.want, got)
			}
		})
	}
}

func TestCtxFirstParam(t *testing.T) {
	testCases := []struct {
		name string
		src  string
		want int
	}{
		{name: "ctx not first is flagged", src: "package x\nimport \"context\"\nfunc F(a int, ctx context.Context) {}\n", want: 1},
		{name: "ctx first is clean", src: "package x\nimport \"context\"\nfunc F(ctx context.Context, a int) {}\n", want: 0},
		{name: "ctx stored in struct is flagged", src: "package x\nimport \"context\"\ntype T struct{ ctx context.Context }\n", want: 1},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			files := parseOne(t, dir, "f.go", tc.src)
			got := CtxFirstParam(dir, files)
			if len(got) != tc.want {
				t.Errorf("CtxFirstParam() = %d, want %d: %v", len(got), tc.want, got)
			}
		})
	}
}

func TestFileLineLimit(t *testing.T) {
	longBody := "package x\n"
	for i := 0; i < 650; i++ {
		longBody += "// line\n"
	}
	testCases := []struct {
		name string
		src  string
		want int
	}{
		{name: "file over 600 lines is flagged", src: longBody, want: 1},
		{name: "short file is clean", src: "package x\nfunc A() {}\n", want: 0},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			files := parseOne(t, dir, "f.go", tc.src)
			got := FileLineLimit(dir, files)
			if len(got) != tc.want {
				t.Errorf("FileLineLimit() = %d, want %d: %v", len(got), tc.want, got)
			}
		})
	}
}

func TestNoInfraInUnitTests(t *testing.T) {
	testCases := []struct {
		name string
		src  string
		want int
	}{
		{
			name: "real outbound http call in a unit test is flagged",
			src:  "package x\nimport (\"net/http\"\n\"testing\")\nfunc TestF(t *testing.T) {\n\thttp.Get(\"https://example.com\")\n}\n",
			want: 1,
		},
		{
			name: "loopback listener call is clean (same idiom validated against real repos)",
			src:  "package x\nimport (\"net\"\n\"testing\")\nfunc TestF(t *testing.T) {\n\tln, _ := net.Listen(\"tcp\", \"127.0.0.1:0\")\n\taddr := ln.Addr().String()\n\tnet.Dial(\"tcp\", addr)\n}\n",
			want: 0,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFile(t, dir, "f_test.go", tc.src)
			got := NoInfraInUnitTests(dir, nil)
			if len(got) != tc.want {
				t.Errorf("NoInfraInUnitTests() = %d, want %d: %v", len(got), tc.want, got)
			}
		})
	}
}
