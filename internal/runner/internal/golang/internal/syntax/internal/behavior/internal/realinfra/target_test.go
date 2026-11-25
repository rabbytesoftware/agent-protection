package realinfra_test

import "testing"

func TestRealinfraTargets(
	t *testing.T,
) {
	assertCases(
		t,
		[]fixture{
			{
				name: "every net dial, listen, lookup and resolve call is live",
				files: map[string]string{
					"p/p_test.go": "package p\n\nimport (\n\t\"net\"\n\t\"testing\"\n)\n\nfunc TestF(\n\tt *testing.T,\n) {\n\tnet.Dial(\"tcp\", \"10.0.0.1:1\")\n\tnet.Listen(\"tcp\", \":8080\")\n\tnet.LookupHost(\"example.com\")\n\tnet.ResolveTCPAddr(\"tcp\", \"example.com:1\")\n}\n",
				},
				want: []string{
					"p/p_test.go:11",
					"p/p_test.go:12",
					"p/p_test.go:13",
					"p/p_test.go:14",
				},
			},
			{
				name: "database drivers are live, versioned and aliased paths included",
				files: map[string]string{
					"p/p_test.go": "package p\n\nimport (\n\t\"github.com/jackc/pgx/v5\"\n\tpool \"github.com/jackc/pgx/v5/pgxpool\"\n\t\"github.com/redis/go-redis/v9\"\n\t\"google.golang.org/grpc\"\n\t\"go.mongodb.org/mongo-driver/v2/mongo\"\n\t\"github.com/nats-io/nats.go\"\n\t\"github.com/segmentio/kafka-go\"\n\t_ \"embed\"\n\t. \"strings\"\n\t\"testing\"\n)\n\nfunc TestF(\n\tt *testing.T,\n) {\n\tpgx.Connect(nil, \"postgres://db\")\n\tpool.New(nil, \"postgres://db\")\n\tredis.NewClient(nil)\n\tgrpc.NewClient(\"dns:///api.example.com:443\")\n\tmongo.Connect(nil)\n\tnats.Connect(\"nats://bus\")\n\tkafka.DialLeader(nil, \"tcp\", \"k:9092\", \"t\", 0)\n}\n",
				},
				want: []string{
					"p/p_test.go:19",
					"p/p_test.go:20",
					"p/p_test.go:21",
					"p/p_test.go:22",
					"p/p_test.go:23",
					"p/p_test.go:24",
					"p/p_test.go:25",
				},
			},
			{
				name: "an in-memory database is owned by the test, a file or server dsn is live",
				files: map[string]string{
					"p/p_test.go": "package p\n\nimport (\n\t\"database/sql\"\n\t\"testing\"\n)\n\nfunc TestF(\n\tt *testing.T,\n) {\n\tsql.Open(\"sqlite\", \":memory:\")\n\tsql.Open(\"sqlite3\", \"file:x?mode=memory&cache=shared\")\n\tsql.Open(\"postgres\", \"postgres://localhost/db\")\n}\n",
				},
				want: []string{
					"p/p_test.go:13",
				},
			},
			{
				name: "an unrelated package exposing the same names is not live",
				files: map[string]string{
					"p/p_test.go": "package p\n\nimport (\n\t\"example.com/cache\"\n\t\"testing\"\n)\n\nfunc TestF(\n\tt *testing.T,\n) {\n\tcache.Open(\"x\")\n\tcache.Get(\"x\")\n\tcache.Dial(\"x\")\n}\n",
				},
			},
		},
	)
}
