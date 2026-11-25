package cqs_test

import "testing"

func TestCqsDirectWrites(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "incrementing a receiver field is reported",
			code: "package p\n\ntype T struct {\n\tn int\n}\n\nfunc (t *T) f() (n int) {\n\tt.n++\n\treturn t.n\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "writing into a receiver map field is reported",
			code: "package p\n\ntype T struct {\n\tm map[string]int\n}\n\nfunc (t *T) f() (n int) {\n\tt.m[\"k\"] = 1\n\treturn 1\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "assigning a nested receiver field is reported",
			code: "package p\n\ntype T struct {\n\tinner struct {\n\t\tv int\n\t}\n}\n\nfunc (t *T) f() (n int) {\n\tt.inner.v = 1\n\treturn 1\n}\n",
			want: []string{"p.go:9"},
		},
		{
			name: "assigning through a parenthesized dereference is reported",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc (t *T) f() (n int) {\n\t(*t).v = 1\n\treturn 1\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "deleting from a receiver map field is reported",
			code: "package p\n\ntype T struct {\n\tm map[string]int\n}\n\nfunc (t *T) f() (n int) {\n\tdelete(t.m, \"k\")\n\treturn 1\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "clearing a receiver field is reported",
			code: "package p\n\ntype T struct {\n\tm map[string]int\n}\n\nfunc (t *T) f() (n int) {\n\tclear(t.m)\n\treturn 1\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "copying into a receiver slice field is reported",
			code: "package p\n\ntype T struct {\n\ts []int\n}\n\nfunc (t *T) f(\n\tfrom []int,\n) (n int) {\n\treturn copy(t.s, from)\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "copying out of a receiver slice field passes",
			code: "package p\n\ntype T struct {\n\ts []int\n}\n\nfunc (t *T) f() (n int) {\n\tto := make([]int, len(t.s))\n\treturn copy(to, t.s)\n}\n",
			want: nil,
		},
		{
			name: "rebinding the pointer parameter itself passes",
			code: "package p\n\ntype T struct {\n\tv int\n}\n\nfunc f(\n\tt *T,\n) (n int) {\n\tt = &T{}\n\t(t) = nil\n\treturn 1\n}\n",
			want: nil,
		},
		{
			name: "incrementing through a pointer field of a value receiver is reported",
			code: "package p\n\ntype tally struct {\n\thits int\n}\n\ntype T struct {\n\ttally *tally\n}\n\nfunc (t T) Square(\n\tv int,\n) (r int) {\n\tt.tally.hits++\n\treturn v * t.tally.hits\n}\n",
			want: []string{"p.go:11"},
		},
		{
			name: "assigning through a nested pointer field of a value parameter is reported",
			code: "package p\n\ntype inner struct {\n\tn int\n}\n\ntype outer struct {\n\tcore struct {\n\t\tref *inner\n\t}\n}\n\nfunc f(\n\to outer,\n) (n int) {\n\to.core.ref.n = 2\n\treturn 1\n}\n",
			want: []string{"p.go:13"},
		},
		{
			name: "writing a value field of a value receiver copy passes",
			code: "package p\n\ntype tally struct {\n\thits int\n}\n\ntype T struct {\n\ttally tally\n}\n\nfunc (t T) Square() (r int) {\n\tt.tally.hits++\n\treturn t.tally.hits\n}\n",
			want: nil,
		},
		{
			name: "writing through a pointer field of a local value passes",
			code: "package p\n\ntype tally struct {\n\thits int\n}\n\ntype T struct {\n\ttally *tally\n}\n\nfunc f() (r int) {\n\tlocal := T{tally: &tally{}}\n\tlocal.tally.hits++\n\treturn local.tally.hits\n}\n",
			want: nil,
		},
		{
			name: "incrementing a local counter passes",
			code: "package p\n\nfunc f() (n int) {\n\tn++\n\treturn n\n}\n",
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

func TestCqsSharedWrites(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "writing a value receiver map field is reported",
			code: "package p\n\ntype T struct {\n\tm map[string]int\n}\n\nfunc (t T) f() (n int) {\n\tt.m[\"k\"]++\n\treturn len(t.m)\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "incrementing a map parameter entry is reported",
			code: "package p\n\nfunc f(\n\ttally map[string]int,\n) (n int) {\n\ttally[\"k\"]++\n\treturn len(tally)\n}\n",
			want: []string{"p.go:3"},
		},
		{
			name: "sending on a channel parameter is reported",
			code: "package p\n\nfunc f(\n\tevents chan<- string,\n) (n int) {\n\tevents <- \"a\"\n\treturn len(events)\n}\n",
			want: []string{"p.go:3"},
		},
		{
			name: "deleting from a map parameter is reported",
			code: "package p\n\nfunc f(\n\tm map[string]int,\n) (n int) {\n\tdelete(m, \"k\")\n\treturn len(m)\n}\n",
			want: []string{"p.go:3"},
		},
		{
			name: "clearing a slice parameter is reported",
			code: "package p\n\nfunc f(\n\tvalues []int,\n) (n int) {\n\tclear(values)\n\treturn len(values)\n}\n",
			want: []string{"p.go:3"},
		},
		{
			name: "writing a slice parameter element is reported",
			code: "package p\n\nfunc f(\n\tvalues []int,\n) (n int) {\n\tvalues[0] = 1\n\treturn len(values)\n}\n",
			want: []string{"p.go:3"},
		},
		{
			name: "appending to a value receiver field is reported",
			code: "package p\n\ntype T struct {\n\ts []int\n}\n\nfunc (t T) f() (n int) {\n\tt.s = append(t.s, 1)\n\treturn len(t.s)\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "sending on a value receiver channel field is reported",
			code: "package p\n\ntype T struct {\n\tc chan int\n}\n\nfunc (t T) f() (n int) {\n\tt.c <- 1\n\treturn len(t.c)\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "sending on a pointer receiver channel field is reported",
			code: "package p\n\ntype T struct {\n\tc chan int\n}\n\nfunc (t *T) f() (n int) {\n\tt.c <- 1\n\treturn len(t.c)\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "passing a map parameter to a helper that writes it is reported",
			code: "package p\n\nfunc put(\n\tm map[string]int,\n) {\n\tm[\"k\"] = 1\n}\n\nfunc f(\n\tm map[string]int,\n) (n int) {\n\tput(m)\n\treturn len(m)\n}\n",
			want: []string{"p.go:9"},
		},
		{
			name: "writing an array parameter copy passes",
			code: "package p\n\nfunc f(\n\tvalues [3]int,\n) (n int) {\n\tvalues[0] = 1\n\treturn values[0]\n}\n",
			want: nil,
		},
		{
			name: "rebinding a slice parameter with append passes",
			code: "package p\n\nfunc f(\n\tvalues []int,\n) (n int) {\n\tvalues = append(values, 1)\n\treturn len(values)\n}\n",
			want: nil,
		},
		{
			name: "reading a map parameter passes",
			code: "package p\n\nfunc f(\n\tm map[string]int,\n) (n int) {\n\treturn m[\"k\"]\n}\n",
			want: nil,
		},
		{
			name: "storing and returning a created value in a value receiver map passes",
			code: "package p\n\ntype Item struct {\n\tname string\n}\n\ntype T struct {\n\tm map[string]Item\n}\n\nfunc (t T) Add(\n\tname string,\n) (item Item) {\n\titem = Item{name: name}\n\tt.m[name] = item\n\treturn item\n}\n",
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

func TestCqsStdlibMutators(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "calling atomic and sync.Map mutators on receiver fields is reported",
			code: "package p\n\nimport (\n\t\"sync\"\n\t\"sync/atomic\"\n)\n\ntype T struct {\n\tmemo  *sync.Map\n\tcalls *atomic.Int64\n}\n\nfunc (t *T) Square(\n\tv int,\n) (r int) {\n\tt.calls.Add(1)\n\tr = v * v\n\tt.memo.Store(v, r)\n\treturn r\n}\n",
			want: []string{"p.go:13"},
		},
		{
			name: "storing into a value atomic field is reported",
			code: "package p\n\nimport \"sync/atomic\"\n\ntype T struct {\n\tn atomic.Int64\n}\n\nfunc (t *T) f() (n int64) {\n\tt.n.Store(1)\n\treturn 1\n}\n",
			want: []string{"p.go:9"},
		},
		{
			name: "writing into a strings.Builder field is reported",
			code: "package p\n\nimport \"strings\"\n\ntype T struct {\n\tb strings.Builder\n}\n\nfunc (t *T) f() (n int) {\n\tt.b.WriteString(\"x\")\n\treturn t.b.Len()\n}\n",
			want: []string{"p.go:9"},
		},
		{
			name: "resetting a buffer on a field of a field is reported",
			code: "package p\n\nimport \"bytes\"\n\ntype Inner struct {\n\tbuf bytes.Buffer\n}\n\ntype T struct {\n\tinner *Inner\n}\n\nfunc (t *T) f() (n int) {\n\tt.inner.buf.Reset()\n\treturn 1\n}\n",
			want: []string{"p.go:13"},
		},
		{
			name: "adding to an atomic inside an anonymous struct field is reported",
			code: "package p\n\nimport \"sync/atomic\"\n\ntype T struct {\n\tstats struct {\n\t\thits atomic.Int64\n\t}\n}\n\nfunc (t *T) f() (n int64) {\n\tt.stats.hits.Add(1)\n\treturn 1\n}\n",
			want: []string{"p.go:11"},
		},
		{
			name: "pushing onto a list parameter is reported",
			code: "package p\n\nimport \"container/list\"\n\nfunc f(\n\tl *list.List,\n) (n int) {\n\tl.PushBack(1)\n\treturn l.Len()\n}\n",
			want: []string{"p.go:5"},
		},
		{
			name: "calling a package-level atomic write on a receiver field is reported",
			code: "package p\n\nimport \"sync/atomic\"\n\ntype T struct {\n\tn int64\n}\n\nfunc (t *T) f() (n int64) {\n\treturn atomic.AddInt64(&t.n, 1)\n}\n",
			want: []string{"p.go:9"},
		},
		{
			name: "sorting a receiver slice field is reported",
			code: "package p\n\nimport \"slices\"\n\ntype T struct {\n\titems []int\n}\n\nfunc (t *T) f() (n int) {\n\tslices.Sort(t.items)\n\treturn t.items[0]\n}\n",
			want: []string{"p.go:9"},
		},
		{
			name: "sorting a slice parameter is reported",
			code: "package p\n\nimport \"sort\"\n\nfunc f(\n\tvalues []int,\n) (n int) {\n\tsort.Ints(values)\n\treturn values[0]\n}\n",
			want: []string{"p.go:5"},
		},
		{
			name: "copying into a receiver map field is reported",
			code: "package p\n\nimport \"maps\"\n\ntype T struct {\n\tm map[string]int\n}\n\nfunc (t *T) f(\n\tother map[string]int,\n) (n int) {\n\tmaps.Copy(t.m, other)\n\treturn len(t.m)\n}\n",
			want: []string{"p.go:9"},
		},
		{
			name: "calling a mutating method of a field's own type is reported",
			code: "package p\n\ntype Inner struct {\n\tv int\n}\n\nfunc (i *Inner) bump() {\n\ti.v++\n}\n\ntype T struct {\n\tinner Inner\n}\n\nfunc (t *T) f() (n int) {\n\tt.inner.bump()\n\treturn t.inner.v\n}\n",
			want: []string{"p.go:15"},
		},
		{
			name: "storing into a generic atomic pointer field is reported",
			code: "package p\n\nimport \"sync/atomic\"\n\ntype T[K comparable, V any] struct {\n\tp atomic.Pointer[V]\n}\n\nfunc (t *T[K, V]) f(\n\tv *V,\n) (n int) {\n\tt.p.Store(v)\n\treturn 1\n}\n",
			want: []string{"p.go:9"},
		},
		{
			name: "mutating through embedded and parenthesized fields is reported",
			code: "package p\n\nimport (\n\t\"strings\"\n\t\"sync\"\n)\n\ntype Inner struct {\n\tb strings.Builder\n}\n\ntype T struct {\n\t*Inner\n\tsync.Map\n}\n\nfunc (t *T) f() (n int) {\n\t(t.Inner).b.Reset()\n\treturn 1\n}\n\nfunc (t *T) g() (n int) {\n\t(*t).Map.Delete(1)\n\treturn 1\n}\n",
			want: []string{"p.go:17", "p.go:22"},
		},
		{
			name: "growing a builder field of a generic receiver is reported",
			code: "package p\n\nimport \"strings\"\n\ntype ID int\n\ntype T[K comparable] struct {\n\tb strings.Builder\n}\n\nfunc (t *T[K]) f() (n int) {\n\tt.b.Grow(1)\n\treturn 1\n}\n",
			want: []string{"p.go:11"},
		},
		{
			name: "locking a mutex while reading passes",
			code: "package p\n\nimport \"sync\"\n\ntype T struct {\n\tmu sync.Mutex\n\trw *sync.RWMutex\n\tv  int\n}\n\nfunc (t *T) f() (n int) {\n\tt.mu.Lock()\n\tdefer t.mu.Unlock()\n\tt.rw.RLock()\n\tdefer t.rw.RUnlock()\n\treturn t.v\n}\n",
			want: nil,
		},
		{
			name: "loading and ranging a sync.Map passes",
			code: "package p\n\nimport \"sync\"\n\ntype T struct {\n\tm *sync.Map\n}\n\nfunc (t *T) f() (n int) {\n\t_, found := t.m.Load(1)\n\tt.m.Range(t.visit)\n\tif found {\n\t\treturn 1\n\t}\n\n\treturn 0\n}\n\nfunc (t *T) visit(\n\tk any,\n\tv any,\n) (more bool) {\n\treturn k != v\n}\n",
			want: nil,
		},
		{
			name: "building into a local strings.Builder passes",
			code: "package p\n\nimport \"strings\"\n\nfunc f(\n\tparts []string,\n) (joined string) {\n\tvar b strings.Builder\n\tfor _, part := range parts {\n\t\tb.WriteString(part)\n\t}\n\n\treturn b.String()\n}\n",
			want: nil,
		},
		{
			name: "loading atomics passes",
			code: "package p\n\nimport \"sync/atomic\"\n\ntype T struct {\n\tn  atomic.Int64\n\tm  int64\n}\n\nfunc (t *T) f() (n int64) {\n\treturn t.n.Load() + atomic.LoadInt64(&t.m)\n}\n",
			want: nil,
		},
		{
			name: "sorting a copy of a receiver map's keys passes",
			code: "package p\n\nimport (\n\t\"maps\"\n\t\"slices\"\n)\n\ntype T struct {\n\tm map[string]int\n}\n\nfunc (t *T) f() (keys []string) {\n\treturn slices.Sorted(maps.Keys(t.m))\n}\n",
			want: nil,
		},
		{
			name: "storing a value it created and returning it passes",
			code: "package p\n\nimport \"sync\"\n\ntype T struct {\n\titems *sync.Map\n}\n\nfunc newID() (id string) {\n\treturn \"x\"\n}\n\nfunc (t *T) Register(\n\tv int,\n) (id string) {\n\tid = newID()\n\tt.items.Store(id, v)\n\treturn id\n}\n",
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
