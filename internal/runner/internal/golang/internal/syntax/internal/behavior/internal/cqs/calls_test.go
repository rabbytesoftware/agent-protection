package cqs_test

import "testing"

func TestCqsThroughCalls(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "calling a mutating method on the receiver is reported",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc (t *T) bump() {\n\tt.v++\n}\n\nfunc (t *T) Next() (v int) {\n\tt.bump()\n\treturn t.v\n}\n",
			want: []string{"p.go:11"},
		},
		{
			name: "passing the receiver to a mutating function is reported",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc touch(\n\tt *T,\n) {\n\tt.v = 1\n}\n\nfunc (t *T) Get() (v int) {\n\ttouch(t)\n\treturn t.v\n}\n",
			want: []string{"p.go:13"},
		},
		{
			name: "passing the address of a receiver field to a mutating function is reported",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc set(\n\ttarget *int,\n) {\n\t*target = 1\n}\n\nfunc (t *T) Get() (v int) {\n\tset(&t.v)\n\treturn t.v\n}\n",
			want: []string{"p.go:13"},
		},
		{
			name: "passing a parenthesized pointer parameter on is reported",
			code: "package p\n\nfunc set(\n\ttarget *int,\n) {\n\t*target = 1\n}\n\nfunc f(\n\tout *int,\n) (n int) {\n\tset((out))\n\treturn 1\n}\n",
			want: []string{"p.go:9"},
		},
		{
			name: "handing the receiver to a mutating method of another parameter is reported",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\ntype U struct{}\n\nfunc (u U) absorb(\n\tt *T,\n) {\n\tt.v = 1\n}\n\nfunc (t *T) Get(\n\tu U,\n) (v int) {\n\tu.absorb(t)\n\treturn t.v\n}\n",
			want: []string{"p.go:15"},
		},
		{
			name: "calling a mutating method on a pointer parameter is reported",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc (t *T) bump() {\n\tt.v++\n}\n\nfunc f(\n\tt *T,\n) (v int) {\n\tt.bump()\n\treturn t.v\n}\n",
			want: []string{"p.go:11"},
		},
		{
			name: "passing a pointer received from a channel to a mutating function passes",
			code: "package p\n\ntype T struct {\n\tv\t   int\n\tpending chan *T\n}\n\nfunc touch(\n\tt *T,\n) {\n\tt.v = 1\n}\n\nfunc (t *T) Get() (v int) {\n\ttouch(<-t.pending)\n\treturn 1\n}\n",
			want: nil,
		},
		{
			name: "calling a read-only method on the receiver passes",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc (t *T) value() (v int) {\n\treturn t.v\n}\n\nfunc (t *T) Get() (v int) {\n\treturn t.value()\n}\n",
			want: nil,
		},
		{
			name: "passing a local pointer to a mutating function passes",
			code: "package p\n\nfunc set(\n\ttarget *int,\n) {\n\t*target = 1\n}\n\nfunc f() (n int) {\n\tset(&n)\n\treturn n\n}\n",
			want: nil,
		},
		{
			name: "passing a receiver field pointer to a mutating function passes",
			code: "package p\n\ntype T struct {\n\tchild *T\n\tv int\n}\n\nfunc touch(\n\tt *T,\n) {\n\tt.v = 1\n}\n\nfunc (t *T) Get() (v int) {\n\ttouch(t.child)\n\treturn 1\n}\n",
			want: nil,
		},
		{
			name: "passing the receiver where the callee mutates another parameter passes",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc touch(\n\ta *T,\n\tb *T,\n) {\n\tb.v = 1\n}\n\nfunc (t *T) Get() (v int) {\n\ttouch(t, &T{})\n\treturn 1\n}\n",
			want: nil,
		},
		{
			name: "locking a receiver field mutex passes",
			code: "package p\n\nimport \"sync\"\n\ntype T struct {\n\tmu sync.Mutex\n\tv  int\n}\n\nfunc (t *T) Get() (v int) {\n\tt.mu.Lock()\n\tdefer t.mu.Unlock()\n\treturn t.v\n}\n",
			want: nil,
		},
		{
			name: "spreading a multi-value call into a mutating function passes",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc touch(\n\tn int,\n\tt *T,\n) {\n\tt.v = n\n}\n\nfunc pair() (n int, t *T) {\n\treturn 1, &T{}\n}\n\nfunc (t *T) Get() (v int) {\n\ttouch(pair())\n\treturn 1\n}\n",
			want: nil,
		},
		{
			name: "calling a method of an unknown receiver passes",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc (t *T) bump() {\n\tt.v++\n}\n\nfunc f(\n\tt *T,\n) (v int) {\n\tg().bump()\n\treturn 1\n}\n\nfunc g() (t *T) {\n\treturn &T{}\n}\n",
			want: nil,
		},
		{
			name: "command calling a mutating helper passes",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc (t *T) bump() {\n\tt.v++\n}\n\nfunc (t *T) Bump() (err error) {\n\tt.bump()\n\treturn nil\n}\n",
			want: nil,
		},
		{
			name: "result-less method on an interface field of the receiver is reported",
			code: "package p\n\ntype Ledger interface {\n\tRecord(\n\t\tvalue int,\n\t)\n\tCount() (count int)\n}\n\ntype T struct {\n\tledger Ledger\n}\n\nfunc (t *T) Square(\n\tv int,\n) (r int) {\n\tt.ledger.Record(v)\n\treturn v * t.ledger.Count()\n}\n",
			want: []string{"p.go:14"},
		},
		{
			name: "result-less func field called through a value receiver is reported",
			code: "package p\n\ntype T struct {\n\trecord func(value int)\n}\n\nfunc (t T) Note(\n\tv int,\n) (r int) {\n\tt.record(v)\n\treturn v\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "result-less method on an interface parameter is reported",
			code: "package p\n\ntype Sink interface {\n\tPut(\n\t\tvalue int,\n\t)\n}\n\nfunc f(\n\tsink Sink,\n) (n int) {\n\tsink.Put(1)\n\treturn 1\n}\n",
			want: []string{"p.go:9"},
		},
		{
			name: "query calling a command helper on the receiver is reported",
			code: "package p\n\ntype T struct {\n\tonHit func()\n}\n\nfunc (t *T) hit() {\n\tt.onHit()\n}\n\nfunc (t *T) Get() (v int) {\n\tt.hit()\n\treturn 1\n}\n",
			want: []string{"p.go:11"},
		},
		{
			name: "interface methods with results on fields pass",
			code: "package p\n\ntype Ledger interface {\n\tCount() (count int)\n}\n\ntype T struct {\n\tledger Ledger\n}\n\nfunc (t *T) Get() (v int) {\n\treturn t.ledger.Count()\n}\n",
			want: nil,
		},
		{
			name: "calling a func parameter passes",
			code: "package p\n\nfunc run(\n\tvisit func(value int),\n) (n int) {\n\tvisit(1)\n\treturn 1\n}\n",
			want: nil,
		},
		{
			name: "lock, cancel and logger calls pass",
			code: "package p\n\ntype locker interface {\n\tLock()\n\tUnlock()\n}\n\ntype Logger interface {\n\tInfo(\n\t\tmessage string,\n\t)\n}\n\ntype T struct {\n\tmu     locker\n\tcancel func()\n\tlog    Logger\n\taudit  auditLogger\n}\n\ntype auditLogger interface {\n\tWrite(\n\t\tline string,\n\t)\n}\n\nfunc (t *T) Get() (v int) {\n\tt.mu.Lock()\n\tdefer t.mu.Unlock()\n\tt.log.Info(\"get\")\n\tt.audit.Write(\"get\")\n\tt.cancel()\n\treturn 1\n}\n",
			want: nil,
		},
		{
			name: "storing the created value through an interface field passes",
			code: "package p\n\ntype Repo interface {\n\tAdd(\n\t\titem *Item,\n\t)\n}\n\ntype T struct {\n\trepo Repo\n}\n\nfunc (t *T) Create() (item *Item) {\n\tcreated := &Item{}\n\tt.repo.Add(created)\n\treturn created\n}\n",
			want: nil,
		},
		{
			name: "result-less command returning only an error passes",
			code: "package p\n\ntype Sink interface {\n\tPut()\n}\n\ntype T struct {\n\tsink Sink\n}\n\nfunc (t *T) Flush() (err error) {\n\tt.sink.Put()\n\treturn nil\n}\n",
			want: nil,
		},
	}
	for _, tc := range cases {
		assertCase(
			t,
			tc.name,
			tc.code,
			tc.want,
		)
	}
}
