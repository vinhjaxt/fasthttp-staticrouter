package main

import (
	"flag"
	"log"
	"strconv"

	"github.com/valyala/fasthttp"
	router "github.com/vinhjaxt/fasthttp-staticrouter"
)

func main() {
	port := flag.Int("port", 8080, "HTTP port")
	flag.Parse()

	r := router.New()

	api := r.Group("/api")

	api.Use(func(ctx *fasthttp.RequestCtx) bool {
		ctx.Response.Header.Set("Content-Type", "application/json")
		return false // return true to abort next handler
	})

	apiv1 := api.Group("/v1")

	apiv1.Use(func(ctx *fasthttp.RequestCtx) bool {
		ctx.Response.Header.Set("X-Hello", "Hello") // dont trust the client
		ctx.SetUserValue("user", map[string]string{"Name": "Guest"})
		return false
	})

	// apiv1.Get

	apiv1.Any("/", func(ctx *fasthttp.RequestCtx) bool {
		user, ok := ctx.UserValue("user").(map[string]string)
		name, ok := user["Name"]
		if !ok {
			ctx.Error("Error", fasthttp.StatusInternalServerError)
			return true
		}
		ctx.SetBodyString("\"Hello " + name + "\"")
		return false
	})

	fs := &fasthttp.FS{
		Root:               "./public_web",
		IndexNames:         []string{"index.html"},
		AcceptByteRange:    true,
		GenerateIndexPages: false,
		Compress:           true,
		// CacheDuration: ,
	}
	r.NotFound(fs.NewRequestHandler())

	api = nil
	apiv1 = nil

	// Start HTTP server.
	s := &fasthttp.Server{
		Handler: r.BuildHandler(),
	}
	log.Println("Server running on", *port)
	log.Panicln(s.ListenAndServe(":" + strconv.Itoa(*port)))
}
