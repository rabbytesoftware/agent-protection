package namedreturn_test

import "testing"

func TestVagueResults(
	t *testing.T,
) {
	cases := []struct {
		name string
		code string
		want []string
	}{
		{
			name: "single letter result on an unexported type falls back to a placeholder",
			code: "package p\n\nfunc Square() (r *square) {\n\treturn nil\n}\n",
			want: []string{"p.go:3 func Square gives results vague names (r *square): name every result after what it holds, for example (<name> *square)"},
		},
		{
			name: "res and a blank error are both named and only the vague ones are listed",
			code: "package p\n\nfunc (s *square) Parse() (v int, _ error) {\n\treturn 0, nil\n}\n",
			want: []string{"p.go:3 method square.Parse gives results vague names (v int, _ error): name every result after what it holds, for example (<name> int, err error)"},
		},
		{
			name: "denylisted words suggest a name derived from the type",
			code: "package p\n\nfunc f() (out Queue, tmp []Job, ret int) {\n\treturn Queue{}, nil, 0\n}\n",
			want: []string{"p.go:3 func f gives results vague names (out Queue, tmp []Job, ret int): name every result after what it holds, for example (queue Queue, jobs []Job, <name> int)"},
		},
		{
			name: "numbered names are vague when the stem is vague or a short conventional name",
			code: "package p\n\nfunc f() (result2 int, err2 error) {\n\treturn 0, nil\n}\n",
			want: []string{"p.go:3 func f gives results vague names (result2 int, err2 error): name every result after what it holds, for example (<name> int, err error)"},
		},
		{
			name: "a literal result is reported with the literal subject",
			code: "package p\n\nvar f = func() (data []byte) {\n\treturn nil\n}\n",
			want: []string{"p.go:3 func literal gives results vague names (data []byte): name every result after what it holds, for example (<name> []byte)"},
		},
		{
			name: "single letter of a type parameter suggests value",
			code: "package p\n\nfunc Get[V any]() (v V) {\n\treturn v\n}\n",
			want: []string{"p.go:3 func Get gives results vague names (v V): name every result after what it holds, for example (value V)"},
		},
		{
			name: "value of a concrete type is vague",
			code: "package p\n\nfunc f() (value int) {\n\treturn 0\n}\n",
			want: []string{"p.go:3 func f gives results vague names (value int): name every result after what it holds, for example (<name> int)"},
		},
		{
			name: "value of a type parameter, any or an empty interface passes, also in literals and generic receivers",
			code: "package p\n\nfunc Get[T any]() (value T) {\n\tg := func() (value T) {\n\t\treturn value\n\t}\n\treturn g()\n}\n\nfunc (b *Box[K, V]) Get() (value V) {\n\treturn value\n}\n\nfunc (b Box[T]) Any() (value any, other interface{}) {\n\treturn nil, nil\n}\n",
			want: nil,
		},
		{
			name: "a vague word that is the type's own name passes",
			code: "package p\n\nfunc f() (result FetchResult, output report.Output, results []FileResult, value reflect.Value, data *Data) {\n\treturn FetchResult{}, nil, nil, reflect.Value{}, nil\n}\n",
			want: nil,
		},
		{
			name: "err, ok, n, meaningful names and domain numbers pass",
			code: "package p\n\nfunc f() (n int, ok bool, err error, sha256 string, squared int) {\n\treturn 0, false, nil, \"\", 0\n}\n",
			want: nil,
		},
		{
			name: "a vague word that is the function's own name passes",
			code: "package p\n\nfunc (t text) value() (value string) {\n\treturn \"\"\n}\n\nfunc (m Message) Data() (data []byte) {\n\treturn nil\n}\n",
			want: nil,
		},
		{
			name: "only the vague names are listed and the rest keep their names, across composite types",
			code: "package p\n\nfunc f[K comparable, V any]() (found Box[K, V], tmp map[string]int, results Pair[K], out pkg.Queue, val interface{ M() }) {\n\treturn found, nil, nil, pkg.Queue{}, nil\n}\n",
			want: []string{"p.go:3 func f gives results vague names (tmp map[string]int, results Pair[K], out pkg.Queue, val interface{M()}): name every result after what it holds, for example (found Box[K, V], <name> map[string]int, pair Pair[K], queue pkg.Queue, <name> interface{M()})"},
		},
		{
			name: "bodiless declarations pass",
			code: "package p\n\nfunc asm() (r int)\n",
			want: nil,
		},
	}
	for _, tc := range cases {
		assertSnippets(
			t,
			tc.name,
			tc.code,
			tc.want,
		)
	}
}
