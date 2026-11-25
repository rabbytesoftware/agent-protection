package realinfra

import (
	"path"
	"slices"
	"strings"
)

const (
	netPath             = "net"
	httpPath            = "net/http"
	httptestPath        = "net/http/httptest"
	osPath              = "os"
	bufconnPath         = "google.golang.org/grpc/test/bufconn"
	miniredisPath       = "github.com/alicebob/miniredis"
	kindClient          = "client"
	kindDialer          = "dialer"
	kindResolver        = "resolver"
	kindTransport       = "transport"
	kindServer          = "server"
	pathSeparator       = "/"
	versionPrefix       = "v"
	digits              = "0123456789"
	listenPrefix        = "Listen"
	ephemeralPort       = ":0"
	testDirectoryMethod = "TempDir"
	addressMethod       = "Addr"
	stringQuotes        = "\"`"
)

func normalizedPath(
	importPath string,
) (normalized string) {
	return path.Join(slices.DeleteFunc(
		strings.Split(
			importPath,
			pathSeparator,
		),
		isMajorVersion,
	)...)
}

func isMajorVersion(
	element string,
) (version bool) {
	number, isPrefixed := strings.CutPrefix(
		element,
		versionPrefix,
	)
	return isPrefixed && number != "" && strings.Trim(
		number,
		digits,
	) == ""
}

func liveFunctions() (functions map[string][]string) {
	return map[string][]string{
		httpPath: {
			"Get",
			"Head",
			"Post",
			"PostForm",
		},
		"database/sql":         {"Open", "OpenDB"},
		"crypto/tls":           {"Dial", "DialWithDialer"},
		"github.com/jackc/pgx": {"Connect", "ConnectConfig"},
		"github.com/jackc/pgx/pgxpool": {
			"New",
			"NewWithConfig",
			"Connect",
			"ConnectConfig",
		},
		"github.com/redis/go-redis":           redisConstructors(),
		"github.com/go-redis/redis":           redisConstructors(),
		"google.golang.org/grpc":              {"Dial", "DialContext", "NewClient"},
		"go.mongodb.org/mongo-driver/mongo":   {"Connect", "NewClient"},
		"github.com/rabbitmq/amqp091-go":      {"Dial", "DialTLS", "DialConfig"},
		"github.com/streadway/amqp":           {"Dial", "DialTLS", "DialConfig"},
		"github.com/nats-io/nats.go":          {"Connect"},
		"github.com/segmentio/kafka-go":       {"Dial", "DialContext", "DialLeader"},
		"github.com/gomodule/redigo/redis":    {"Dial", "DialURL", "DialContext"},
		"github.com/ClickHouse/clickhouse-go": {"Open"},
		"github.com/jmoiron/sqlx": {
			"Open",
			"Connect",
			"MustConnect",
			"MustOpen",
		},
		"gorm.io/gorm": {"Open"},
	}
}

func redisConstructors() (names []string) {
	return []string{
		"NewClient",
		"NewClusterClient",
		"NewFailoverClient",
		"NewUniversalClient",
		"NewRing",
	}
}

func livePrefixes() (prefixes map[string][]string) {
	return map[string][]string{
		netPath: {
			"Dial",
			listenPrefix,
			"Lookup",
			"Resolve",
		},
	}
}

func liveMethods() (methods map[string][]string) {
	return map[string][]string{
		kindClient: {
			"Do",
			"Get",
			"Head",
			"Post",
			"PostForm",
		},
		kindDialer:    {"Dial", "DialContext"},
		kindTransport: {"RoundTrip"},
		kindResolver: {
			"LookupAddr",
			"LookupCNAME",
			"LookupHost",
			"LookupIP",
			"LookupIPAddr",
			"LookupMX",
			"LookupNS",
			"LookupNetIP",
			"LookupPort",
			"LookupSRV",
			"LookupTXT",
		},
	}
}

func typeKinds() (kinds map[string]map[string]string) {
	return map[string]map[string]string{
		httpPath: {"Client": kindClient, "Transport": kindTransport},
		netPath: {
			"Dialer":       kindDialer,
			"Resolver":     kindResolver,
			"Listener":     kindServer,
			"TCPListener":  kindServer,
			"UnixListener": kindServer,
		},
		httptestPath:  {"Server": kindServer},
		bufconnPath:   {"Listener": kindServer},
		miniredisPath: {"Miniredis": kindServer},
	}
}

func packageValues() (values map[string]map[string]string) {
	return map[string]map[string]string{
		httpPath: {"DefaultClient": kindClient, "DefaultTransport": kindTransport},
		netPath:  {"DefaultResolver": kindResolver},
	}
}

func ownedConstructors() (constructors map[string][]string) {
	return map[string][]string{
		httptestPath:  {"NewServer", "NewTLSServer", "NewUnstartedServer"},
		bufconnPath:   {"Listen"},
		miniredisPath: {"Run", "RunT", "NewMiniRedis"},
		osPath:        {"CreateTemp", "MkdirTemp"},
	}
}
