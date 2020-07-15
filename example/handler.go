package main

import (
	"log"

	"net/http"
	_ "net/http/pprof"

	"github.com/valyala/fasthttp"
	router "github.com/vinhjaxt/fasthttp-staticrouter"
)

func buildHTTPHandler(staticDir string) func(*fasthttp.RequestCtx) {
	go func() {
		log.Println(http.ListenAndServe(":6060", nil))
	}()
	r := router.New()

	api := r.Group("/api")
	api.Use(func(ctx *fasthttp.RequestCtx) (b bool) {
		b = true

		ctx.Response.Header.SetContentType("application/json;charset=utf-8")
		// Do authorization
		auth := true // Do something
		if !auth {
			// Abort request
			return
		}

		// Next router
		b = false
		return
	})

	apiv1 := api.Group("/v1.0")
	apiv1.Use(func(ctx *fasthttp.RequestCtx) (b bool) {
		ctx.Response.Header.Set("X-API-Version", "1.0")
		return
	})
	apiv1.Get("/", func(ctx *fasthttp.RequestCtx) (_ bool) {
		ctx.SetBodyString(`"Hello world"`)
		return
	})
	apiv1.Post("/", func(ctx *fasthttp.RequestCtx) (_ bool) {
		ctx.SetBodyString(`"Hello world"`)
		return
	})

	// Not found handler
	notFoundHandler := func(ctx *fasthttp.RequestCtx) {
		ctx.SetStatusCode(404)
	}
	if len(staticDir) != 0 {
		log.Println("Static file serve:", staticDir)
		// Go to: https://github.com/valyala/fasthttp/blob/master/examples/fileserver/fileserver.go
		// for more informations
		fs := &fasthttp.FS{
			Root:               staticDir,
			IndexNames:         []string{"index.html"},
			GenerateIndexPages: false,
			Compress:           true,
			AcceptByteRange:    true,
			PathNotFound:       notFoundHandler,
		}
		notFoundHandler = fs.NewRequestHandler()
	}
	r.NotFound(notFoundHandler)
	return r.BuildHandler()
}
