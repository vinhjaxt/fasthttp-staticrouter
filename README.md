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
`wrk -t20 -c1000 -d30s http://127.0.0.1:8080/api/v1.0/`
### /index.html
```
Running 30s test @ http://127.0.0.1:8080/
  20 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    18.09ms   10.35ms 123.02ms   77.51%
    Req/Sec     2.84k     0.95k   19.18k    82.24%
  1686971 requests in 30.09s, 313.72MB read
Requests/sec:  56071.61
Transfer/sec:     10.43MB
```

### /api/v1.0/
```
Running 30s test @ http://127.0.0.1:8080/api/v1.0/
  20 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    16.47ms    9.74ms 122.89ms   77.09%
    Req/Sec     3.13k     1.10k   19.13k    81.63%
  1857419 requests in 30.09s, 274.56MB read
Requests/sec:  61721.53
Transfer/sec:      9.12MB
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
29.5 MB

## License
- MIT
