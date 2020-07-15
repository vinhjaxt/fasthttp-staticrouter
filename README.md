## fasthttp-staticrouter
Simple fasthttp static router not support params in url but fast

## Features

- Static http routing which simple fast

## Usage
`go get -u github.com/vinhjaxt/fasthttp-staticrouter`

```go
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
```
Checkout [Example](example/main.go)

## Available api:
  - r.New : Create a new router
  - r.Use : Use a Middleware
  - r.Get, r.Post, r.Put, r.Patch, r.Delete, r.Options, r.Head, r.Method, r.Any : HTTP methods
  - r.Group : Create group of routers
  - r.NotFound : Set not found handler
  - r.MethodNotAllowed : Set MethodNotAllowed handler
  - r.Recover or r.OnError : Set panic handler 

## Performance on example
### Serve
`./example -bind-tcp :8080 -static-dir ./public_web/`
### Test
`wrk -t20 -c1000 -d30s -R1000000 http://127.0.0.1:8080/api/v1.0/`
### /index.html
```
Running 30s test @ http://127.0.0.1:8080/
  20 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    18.65s     5.45s   28.62s    57.66%
    Req/Sec     2.68k   216.40     3.21k    70.00%
  1614174 requests in 29.99s, 300.18MB read
Requests/sec:  53814.91
Transfer/sec:     10.01MB
```

### /api/v1.0/
```
Running 30s test @ http://127.0.0.1:8080/api/v1.0/
  20 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    18.56s     5.35s   28.44s    57.98%
    Req/Sec     3.01k   178.65     3.49k    70.00%
  1805231 requests in 30.00s, 266.85MB read
Requests/sec:  60183.75
Transfer/sec:      8.90MB
```
### Test Machine
```
OS: Parrot GNU/Linux 4.10 x86_64 
Host: 4290CTO ThinkPad X220 
Kernel: 5.6.0-2parrot1-amd64 
CPU: Intel i5-2410M (4) @ 2.900GHz 
GPU: Intel 2nd Generation Core Proces 
Memory: 4407MiB / 9856MiB 
```
### Mem usage
31.9 MB

## License
- MIT
