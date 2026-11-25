package comments_test

import "testing"

func TestWalker(
	t *testing.T,
) {
	cases := []sourceCase{
		{name: "nested source directory is scanned", path: "src/deep/a.ts", code: "// x\n", want: []string{"R07 src/deep/a.ts:1"}},
		{name: "node_modules is skipped", path: "node_modules/p/a.js", code: "// x\n", want: nil},
		{name: "target is skipped", path: "target/debug/build/a.rs", code: "// x\n", want: nil},
		{name: "dist is skipped", path: "dist/a.js", code: "// x\n", want: nil},
		{name: "build is skipped", path: "build/a.js", code: "// x\n", want: nil},
		{name: "out is skipped", path: "out/a.js", code: "// x\n", want: nil},
		{name: "coverage is skipped", path: "coverage/a.js", code: "// x\n", want: nil},
		{name: "vendor is skipped", path: "vendor/a.rs", code: "// x\n", want: nil},
		{name: "git directory is skipped", path: ".git/a.ts", code: "// x\n", want: nil},
		{name: "next directory is skipped", path: ".next/a.js", code: "// x\n", want: nil},
		{name: "turbo directory is skipped", path: ".turbo/a.js", code: "// x\n", want: nil},
		{name: "any hidden directory is skipped", path: ".cache/a.ts", code: "// x\n", want: nil},
		{name: "a directory merely containing a skipped name is scanned", path: "outbox/a.ts", code: "// x\n", want: []string{"R07 outbox/a.ts:1"}},
		{name: "every script extension is scanned", path: "a.cts", code: "// x\n", want: []string{"R07 a.cts:1"}},
		{name: "declaration files are scanned", path: "types/a.d.ts", code: "/** x */\n", want: []string{"R07 types/a.d.ts:1"}},
	}
	for _, tc := range cases {
		assertScanned(
			t,
			tc,
		)
	}
}
