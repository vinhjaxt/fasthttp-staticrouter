package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/valyala/fasthttp"
)

func main() {
	bindTCP := flag.String("bind-tcp", ":80", "Bind TCP address")
	bindSockFile := flag.String("bind-file", "", "Bind Unix socket file")
	staticDir := flag.String("static-dir", "", "Static file server directory")
	flag.Parse()

	s := &fasthttp.Server{
		// ErrorHandler: nil,
		Handler:               buildHTTPHandler(*staticDir),
		NoDefaultServerHeader: true, // Don't send Server: fasthttp
		// Name: "nginx",  // Send Server header
		ReadBufferSize:                2 * 4096, // Make sure these are big enough.
		WriteBufferSize:               4096,
		ReadTimeout:                   5 * time.Second,
		WriteTimeout:                  time.Second,
		IdleTimeout:                   time.Minute, // This can be long for keep-alive connections.
		DisableHeaderNamesNormalizing: false,       // If you're not going to look at headers or know the casing you can set this.
		// NoDefaultContentType: true, // Don't send Content-Type: text/plain if no Content-Type is set manually.
		MaxRequestBodySize: 2 * 1024 * 1024, // 2MB
		DisableKeepalive:   false,
		KeepHijackedConns:  true,
		// NoDefaultDate: len(*staticDir) == 0,
		ReduceMemoryUsage: true,
		TCPKeepalive:      true,
		// TCPKeepalivePeriod: 10 * time.Second,
		// MaxRequestsPerConn: 1000,
		// MaxConnsPerIP: 20,
	}
	if len(*bindSockFile) == 0 {
		log.Println("Listening on", *bindTCP)
		log.Panicln(s.ListenAndServe(*bindTCP))
	} else {
		log.Println("Listening on", *bindSockFile)
		log.Panicln(s.ListenAndServeUNIX(*bindSockFile, os.ModePerm))
	}
}
