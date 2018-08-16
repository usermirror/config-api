package server

import (
	"fmt"

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
	// VaultAddr is the address of the running vault server.
	VaultAddr string
	// VaultToken is root token for vault to read/write secure configurations.
	VaultToken string
}

// Listen starts the proxy server
func (server *Server) Listen() error {
	switch server.StorageBackend {
	case "etcd":
		if etcd, err := storage.NewEtcd(server.EtcdAddr); err == nil {
			store = etcd
			defer etcd.Close()
		}
	case "vault":
		if vault, err := storage.NewVault(server.VaultAddr, server.VaultToken); err == nil {
			store = vault
		}
	default:
		// Use default redis backend
		server.StorageBackend = "redis"
	}

	fmt.Println(fmt.Sprintf("server.config: using %s as default storage backend", server.StorageBackend))

	router := fasthttprouter.New()

	router.GET("/internal/health", CORS(ok))

	router.OPTIONS("/v1/namespaces/:namespaceId/configs/:configId", CORS(ok))
	router.GET("/v1/namespaces/:namespaceId/configs/:configId", CORS(GetHandler))
	router.PUT("/v1/namespaces/:namespaceId/configs/:configId", CORS(PutHandler))
	router.POST("/v1/namespaces/:namespaceId/configs", CORS(PostHandler))

	fmt.Println(fmt.Sprintf("server.listen: %s", server.Addr))
	return fasthttp.ListenAndServe(server.Addr, router.Handler)
}
