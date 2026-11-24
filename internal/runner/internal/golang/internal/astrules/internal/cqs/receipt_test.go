package cqs_test

import "testing"

func TestCqsCreatedValue(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "returning the id it generated and registered through a helper passes",
			code: "package p\n\nimport \"ids\"\n\ntype B struct {\n\tsubs map[string]int\n}\n\nfunc (b *B) register(\n\tid string,\n) {\n\tb.subs[id] = 1\n}\n\nfunc (b *B) Subscribe(\n\tpattern string,\n) (subID string, err error) {\n\tif pattern == \"\" {\n\t\treturn \"\", nil\n\t}\n\n\tsubID = ids.New()\n\tb.register(subID)\n\treturn subID, nil\n}\n",
			want: nil,
		},
		{
			name: "returning a created value written directly into the receiver passes",
			code: "package p\n\ntype S struct {\n\titems map[int]*Item\n}\n\ntype Item struct{}\n\nfunc (s *S) Add() (item *Item) {\n\titem = &Item{}\n\ts.items[1] = item\n\treturn item\n}\n",
			want: nil,
		},
		{
			name: "returning a value read from the receiver is reported",
			code: "package p\n\ntype S struct {\n\tn int\n}\n\nfunc (s *S) Next() (id int) {\n\tid = s.counter()\n\ts.n = id\n\treturn id\n}\n\nfunc (s *S) counter() (n int) {\n\treturn s.n + 1\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "returning a created value that never reaches the write is reported",
			code: "package p\n\nimport \"ids\"\n\ntype S struct {\n\tn int\n}\n\nfunc (s *S) Touch() (id string) {\n\tid = ids.New()\n\ts.n++\n\treturn id\n}\n",
			want: []string{"p.go:9"},
		},
		{
			name: "returning a parameter that was stored is reported",
			code: "package p\n\ntype S struct {\n\tlast string\n}\n\nfunc (s *S) Swap(\n\tnext string,\n) (previous string) {\n\ts.last = next\n\treturn next\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "a value declared without a constructor is reported",
			code: "package p\n\ntype S struct {\n\tlast string\n}\n\nfunc (s *S) Mark() (stamp string) {\n\tvar label string\n\ts.last = label\n\treturn label\n}\n",
			want: []string{"p.go:7"},
		},
		{
			name: "a value created by var and returned alongside the error passes",
			code: "package p\n\nimport \"ids\"\n\ntype S struct {\n\tlast string\n}\n\nfunc (s *S) Mark() (stamp string, err error) {\n\tvar label = ids.New()\n\ts.last = label\n\treturn label, nil\n}\n",
			want: nil,
		},
		{
			name: "returning a call result directly is reported",
			code: "package p\n\nimport \"ids\"\n\ntype S struct {\n\tlast string\n}\n\nfunc (s *S) Mark() (stamp string) {\n\ts.last = \"x\"\n\treturn ids.New()\n}\n",
			want: []string{"p.go:9"},
		},
		{
			name: "a value received from a channel is reported",
			code: "package p\n\ntype S struct {\n\tlast string\n}\n\nfunc (s *S) Take(\n\tsource chan string,\n) (taken string) {\n\ttaken = <-source\n\ts.last = taken\n\treturn taken\n}\n",
			want: []string{"p.go:7"},
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

func TestCqsStoredValue(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "a created value stored through a nested helper chain passes",
			code: "package p\n\ntype S struct {\n\tm map[int]int\n}\n\nfunc (s *S) put(\n\tv int,\n) {\n\ts.store(v)\n}\n\nfunc (s *S) store(\n\tv int,\n) {\n\ts.m[v] = v\n}\n\nfunc (s *S) Add() (v int) {\n\tv = len(make([]int, 3))\n\ts.put(v)\n\treturn v\n}\n",
			want: nil,
		},
		{
			name: "a created value assigned in a tuple with an error passes",
			code: "package p\n\nimport \"ids\"\n\ntype S struct {\n\tlast string\n}\n\nfunc (s *S) Mark() (stamp string, err error) {\n\tstamp, err = ids.Parse(\"x\")\n\tif err != nil {\n\t\treturn \"\", err\n\t}\n\n\ts.last = stamp\n\treturn stamp, nil\n}\n",
			want: nil,
		},
		{
			name: "a value only read near a write is reported",
			code: "package p\n\nimport \"ids\"\n\ntype S struct {\n\tlast string\n}\n\nfunc (s *S) Mark() (stamp string) {\n\tstamp = ids.New()\n\tif stamp != \"\" {\n\t\ts.last = \"x\"\n\t}\n\n\treturn stamp\n}\n",
			want: []string{"p.go:9"},
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
