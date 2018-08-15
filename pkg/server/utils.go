package server

import (
	"github.com/valyala/fasthttp"
)

var (
	okResp = []byte(`{"ok":true}`)
	okType = []byte("application/json")
)

// Simple health check 200-response
func ok(ctx *fasthttp.RequestCtx) {
	ctx.SetContentTypeBytes(okType)
	ctx.Write(okResp)
}

var (
	corsAllowMethods     = "HEAD,GET,POST,PUT,DELETE,OPTIONS"
	corsAllowHeaders     = "*"
	corsAllowOrigin      = "*"
	corsAllowCredentials = "true"
)

// CORS support for the request
func CORS(handle fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Set("Access-Control-Allow-Credentials", corsAllowCredentials)
		ctx.Response.Header.Set("Access-Control-Allow-Headers", corsAllowHeaders)
		ctx.Response.Header.Set("Access-Control-Allow-Methods", corsAllowMethods)
		ctx.Response.Header.Set("Access-Control-Allow-Origin", corsAllowOrigin)

		handle(ctx)
	}
}
