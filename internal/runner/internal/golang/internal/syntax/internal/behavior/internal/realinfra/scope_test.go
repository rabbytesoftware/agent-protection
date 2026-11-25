package realinfra_test

import "testing"

func TestRealinfraScope(
	t *testing.T,
) {
	assertCases(
		t,
		[]fixture{
			{
				name: "client, dialer, resolver and transport values make live method calls",
				files: map[string]string{
					"p/p_test.go": "package p\n\nimport (\n\t\"net\"\n\t\"net/http\"\n\t\"testing\"\n)\n\nfunc TestF(\n\tt *testing.T,\n) {\n\tclient := &http.Client{}\n\tclient.Get(\"https://example.com\")\n\tvar dialer net.Dialer\n\tdialer.DialContext(nil, \"tcp\", \"db:1\")\n\tresolver := new(net.Resolver)\n\tresolver.LookupHost(nil, \"example.com\")\n\thttp.DefaultClient.Do(nil)\n\thttp.DefaultTransport.RoundTrip(nil)\n\tnet.DefaultResolver.LookupIP(nil, \"ip\", \"example.com\")\n\t(&http.Client{}).Post(\"https://example.com\", \"text/plain\", nil)\n}\n",
				},
				want: []string{
					"p/p_test.go:13",
					"p/p_test.go:15",
					"p/p_test.go:17",
					"p/p_test.go:18",
					"p/p_test.go:19",
					"p/p_test.go:20",
					"p/p_test.go:21",
				},
			},
			{
				name: "a client field of a test fixture makes live calls",
				files: map[string]string{
					"p/p_test.go": "package p\n\nimport (\n\t\"net/http\"\n\t\"testing\"\n)\n\ntype fixture struct {\n\tclient *http.Client\n\tlabel  string\n}\n\nfunc (f fixture) fetch() {\n\tf.client.Get(\"https://example.com\")\n\tf.other.Get(\"x\")\n}\n",
				},
				want: []string{
					"p/p_test.go:14",
				},
			},
			{
				name: "a request built on the test server url may go through any client",
				files: map[string]string{
					"p/p_test.go": "package p\n\nimport (\n\t\"net/http\"\n\t\"net/http/httptest\"\n\t\"testing\"\n)\n\nfunc TestF(\n\tt *testing.T,\n) {\n\tsrv := httptest.NewTLSServer(nil)\n\trequest, err := http.NewRequest(\"GET\", srv.URL, nil)\n\thttp.DefaultClient.Do(request)\n\t_ = err\n}\n",
				},
			},
			{
				name: "a non-ephemeral listener is live while the ephemeral one owns its address",
				files: map[string]string{
					"p/p_test.go": "package p\n\nimport (\n\t\"net\"\n\t\"testing\"\n)\n\nfunc TestF(\n\tt *testing.T,\n) {\n\tlistener, err := net.Listen(\"tcp\", \"127.0.0.1:0\")\n\taddress := listener.Addr().String()\n\tnet.Dial(\"tcp\", address)\n\tnet.ListenPacket(\"udp\", \":0\")\n\tnet.Listen(\"tcp\", \"127.0.0.1:8080\")\n\t_ = err\n}\n",
				},
				want: []string{
					"p/p_test.go:15",
				},
			},
			{
				name: "a non-literal address is not ephemeral and a unix socket in a temp dir is owned",
				files: map[string]string{
					"p/p_test.go": "package p\n\nimport (\n\t\"net\"\n\t\"path/filepath\"\n\t\"testing\"\n)\n\nfunc TestF(\n\tt *testing.T,\n) {\n\tnet.Listen(\"tcp\", address)\n\tnet.Listen(network, `:0`+suffix)\n\tsocket := filepath.Join(t.TempDir(), \"s.sock\")\n\tnet.Listen(\"unix\", socket)\n\tnet.Dial(\"unix\", socket)\n\tnet.Dial(\"unix\", \"/var/run/docker.sock\")\n}\n",
				},
				want: []string{
					"p/p_test.go:12",
					"p/p_test.go:13",
					"p/p_test.go:17",
				},
			},
		},
	)
}
