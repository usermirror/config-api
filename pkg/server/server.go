package server

import (
	"fmt"
	"runtime/debug"

	"github.com/usermirror/config-api/pkg/storage"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

// Server is the config API http server
type Server struct {
	// Addr is the address where `config-api` listens for incoming requests.
	Addr string
	// StorageBackend is the default storage backend that will be used.
	StorageBackend string
	// EtcdAddr is the address of the running etcd cluster.
	EtcdAddr string
	// RedisAddr is the address of the running redis cluster.
	RedisAddr string
	// VaultAddr is the address of the running vault server.
	VaultAddr string
	// VaultToken is root token for vault to read/write secure configurations.
	VaultToken string
	// PostgresAddr is the address of the running postgres database.
	PostgresAddr string
	// CheckAuth will validate write tokens against the token for the namespace.
	CheckAuth bool
}

// Listen starts the proxy server
func (server *Server) Listen() error {
	switch server.StorageBackend {
	case "etcd":
		if etcd, err := storage.NewEtcd(server.EtcdAddr); err == nil {
			store = etcd
			defer etcd.Close()
		} else {
			panic(err)
		}
	case "postgres":
		if postgres, err := storage.NewPostgres(server.PostgresAddr); err == nil {
			store = postgres
			defer postgres.Close()
			if err = postgres.Init(); err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	case "vault":
		if vault, err := storage.NewVault(server.VaultAddr, server.VaultToken); err == nil {
			store = vault
		} else {
			panic(err)
		}
	default:
		// Use default redis backend
		server.StorageBackend = "redis"
		fmt.Println(fmt.Sprintf("redis.connect: %s", server.RedisAddr))
	}

	fmt.Println(fmt.Sprintf("server.config: using %s as default storage backend", server.StorageBackend))

	router := fasthttprouter.New()

	router.PanicHandler = handlePanic

	router.GET("/", CORS(ok))
	router.GET("/internal/health", CORS(ok))

	router.OPTIONS("/v1/namespaces/:namespaceId/configs", CORS(ok))
	router.GET("/v1/namespaces/:namespaceId/configs", CORS(server.ScanHandler))
	router.POST("/v1/namespaces/:namespaceId/configs", CORS(server.PostHandler))

	router.OPTIONS("/v1/namespaces/:namespaceId/configs/:configId", CORS(ok))
	router.GET("/v1/namespaces/:namespaceId/configs/:configId", CORS(server.GetHandler))
	router.PUT("/v1/namespaces/:namespaceId/configs/:configId", CORS(server.PutHandler))

	fmt.Println(fmt.Sprintf("server.listen: %s", server.Addr))
	return fasthttp.ListenAndServe(server.Addr, router.Handler)
}

func handlePanic(ctx *fasthttp.RequestCtx, err interface{}) {
	if err != nil {
		fmt.Println(fmt.Sprintf("server.panic: %s", err))
		debug.PrintStack()
		ctx.Write(toJSON(map[string]interface{}{
			"error":   true,
			"message": fmt.Sprintf("%s", err),
		}))
	} else {
		ctx.Write(toJSON(map[string]interface{}{
			"error":   true,
			"message": "unknown",
		}))
	}
}
