package realinfra_test

import "testing"

func TestRealinfraOwnership(
	t *testing.T,
) {
	assertCases(
		t,
		[]fixture{
			{
				name: "a helper that builds a test server or a temp socket owns what it returns",
				files: map[string]string{
					"p/helpers_test.go": "package p\n\nimport (\n\t\"net/http/httptest\"\n\t\"os\"\n\t\"testing\"\n)\n\nfunc newServer(\n\tt *testing.T,\n) (label string, srv *httptest.Server) {\n\treturn \"x\", nil\n}\n\nfunc socketPath() (path string) {\n\tfile, _ := os.CreateTemp(\"\", \"s\")\n\treturn file.Name()\n}\n\nfunc (f fixture) method() {}\n",
					"p/p_test.go":       "package p\n\nimport (\n\t\"net\"\n\t\"net/http\"\n\t\"testing\"\n)\n\nfunc TestF(\n\tt *testing.T,\n) {\n\t_, srv := newServer(t)\n\thttp.Get(srv.URL)\n\tpath := socketPath()\n\tnet.Dial(\"unix\", path)\n}\n",
				},
			},
			{
				name: "a listener returned by the code under test owns its Addr",
				files: map[string]string{
					"p/p_test.go": "package p\n\nimport (\n\t\"example.com/gateway\"\n\t\"net\"\n\t\"testing\"\n)\n\nfunc TestF(\n\tt *testing.T,\n) {\n\tlistener, err := gateway.New(\"tcp://127.0.0.1:0\")\n\tnet.Dial(\"tcp\", listener.Addr().String())\n\tnet.Dial(\"tcp\", remote.Addr().String())\n\t_ = err\n}\n",
				},
				want: []string{
					"p/p_test.go:14",
				},
			},
			{
				name: "server parameters, server fields and multi-value declarations carry ownership",
				files: map[string]string{
					"p/p_test.go": "package p\n\nimport (\n\t\"net/http\"\n\t\"net/http/httptest\"\n\t\"net\"\n\t\"github.com/alicebob/miniredis/v2\"\n\t\"github.com/redis/go-redis/v9\"\n\t\"testing\"\n)\n\ntype suite struct {\n\tserver *httptest.Server\n}\n\nfunc (s suite) fetch(\n\tother *httptest.Server,\n) {\n\thttp.Get(s.server.URL)\n\thttp.Get(other.URL)\n\tvar a, b = other.URL, \"https://example.com\"\n\thttp.Get(a)\n\thttp.Get(b)\n\tvar c, d string = other.URL, \"x\"\n\thttp.Get(c + d)\n\tvar listener net.Listener\n\tnet.Dial(\"tcp\", listener.Addr().String())\n\tvar empty string\n\thttp.Get(empty)\n\tmini := miniredis.RunT(nil)\n\tredis.NewClient(&redis.Options{Addr: mini.Addr()})\n\tfirst, second := other.URL, \"https://example.com\"\n\thttp.Get(first)\n\thttp.Get(second)\n\tfixture.field, plain := other.URL, \"\"\n\thttp.Get(plain)\n}\n",
				},
				want: []string{
					"p/p_test.go:23",
					"p/p_test.go:29",
					"p/p_test.go:34",
					"p/p_test.go:36",
				},
			},
			{
				name: "a closure inside the test sees the test's ownership",
				files: map[string]string{
					"p/p_test.go": "package p\n\nimport (\n\t\"net/http\"\n\t\"net/http/httptest\"\n\t\"testing\"\n)\n\nfunc TestF(\n\tt *testing.T,\n) {\n\tsrv := httptest.NewServer(nil)\n\tfunc(target string) {\n\t\thttp.Get(target)\n\t}(srv.URL)\n\trun(func(t *testing.T) {\n\t\thttp.Get(srv.URL)\n\t})\n}\n",
				},
				want: []string{
					"p/p_test.go:14",
				},
			},
		},
	)
}
