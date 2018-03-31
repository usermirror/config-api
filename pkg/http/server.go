package http

import (
	"fmt"

	"github.com/buaazp/fasthttprouter"
	"github.com/usermirror/config-api/pkg/models/config"
	"github.com/valyala/fasthttp"
)

// Server is the config API http server
type Server struct {
	Addr string
}

// Listen starts the proxy server
func (server *Server) Listen() error {
	router := fasthttprouter.New()

	router.GET("/v1/namespaces/:namespaceId/configs/:configId", config.GetHandler)
	router.PUT("/v1/namespaces/:namespaceId/configs/:configId", config.PutHandler)
	router.POST("/v1/namespaces/:namespaceId/configs", config.PostHandler)

	fmt.Println(fmt.Sprintf("server.listen: %s", server.Addr))
	return fasthttp.ListenAndServe(server.Addr, router.Handler)
}
