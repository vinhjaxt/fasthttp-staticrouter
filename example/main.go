package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/valyala/fasthttp"
	"github.com/vinhjaxt/fasthttp-staticrouter"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	r := router.New()

	api := r.Group("/api")

	api.Use(func(c *router.Context) {
		c.Response.Header.Set("Content-Type", "application/json")
	})

	apiv1 := api.Group("/v1")

	apiv1.Use(func(c *router.Context) {
		c.Response.Header.Set("X-Hello", "Hello")
		c.SetUserValue("user", map[string]string{"Name": "Guest"})
	})

	// apiv1.Get

	apiv1.Any("/", func(c *router.Context) {
		user, ok := c.UserValue("user").(map[string]string)
		name, ok := user["Name"]
		if !ok {
			c.Error("Error", fasthttp.StatusInternalServerError)
			return
		}
		c.SetBodyString("\"Hello " + name + "\"")
	})

	fs := &fasthttp.FS{
		Root:               "./public_web",
		IndexNames:         []string{"index.html", "index.htm"},
		AcceptByteRange:    true,
		GenerateIndexPages: false,
		Compress:           true,
		// CacheDuration: ,
	}
	FSHandler := fs.NewRequestHandler()
	r.NotFound(func(c *router.Context) {
		FSHandler(c.RequestCtx)
	})
	requestHandler := r.BuildHandler()
	api = nil
	apiv1 = nil
	r = nil

	go func() {
		HTTPPort, err := strconv.Atoi(strings.Trim(os.Getenv("GO_APP_HTTP_PORT"), " \r\n\t"))
		if err != nil {
			HTTPPort = 80
		}
		// Start HTTP server.
		HTTPServer := &fasthttp.Server{
			Handler:              requestHandler,
			Name:                 "nginx",
			ReadTimeout:          120 * 1000000000, // 120s
			WriteTimeout:         120 * 1000000000,
			MaxKeepaliveDuration: 120 * 1000000000,
			MaxRequestBodySize:   2 * 1048576, // 2MB
		}
		log.Printf("\r\nHTTP Server running on port %d", HTTPPort)
		if err := HTTPServer.ListenAndServe(fmt.Sprintf(":%d", HTTPPort)); err != nil {
			log.Fatalf("error in ListenAndServe: %s", err)
		}
	}()

	go func() {
		HTTPSPort, err := strconv.Atoi(strings.Trim(os.Getenv("GO_APP_HTTPS_PORT"), " \r\n\t"))
		if err == nil {
			// Start HTTPS server.
			HTTPSServer := &fasthttp.Server{
				Handler:              requestHandler,
				Name:                 "nginx",
				ReadTimeout:          120 * 1000000000, // 120s
				WriteTimeout:         120 * 1000000000,
				MaxKeepaliveDuration: 120 * 1000000000,
				MaxRequestBodySize:   2 * 1048576, // 2MB
			}
			certFile := ""
			keyFile := ""
			log.Printf("Starting HTTPS server on port %d", HTTPSPort)
			if err := HTTPSServer.ListenAndServeTLS(fmt.Sprintf(":%d", HTTPSPort), certFile, keyFile); err != nil {
				log.Fatalf("error in ListenAndServeTLS: %s", err)
			}
		}
	}()

	// Wait forever.
	select {}
}
