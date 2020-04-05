# fasthttp-staticrouter
Simple fasthttp static router not support params in url but fast

## Features

- Static http routing which simple fast

## Usage
`go get -u github.com/vinhjaxt/fasthttp-staticrouter`

```go
r := router.New()

api := r.Group("/api")

api.Use(func(ctx *fasthttp.RequestCtx) bool {
  ctx.Response.Header.Set("Content-Type", "application/json")
  return false
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
```
Checkout [Example](example/main.go)

# Available api:
  - r.New : Create new router
  - r.Use : Middleware
  - r.Get, r.Post, r.Put, r.Patch, r.Delete, r.Options, r.Head : HTTP methods
  - r.Group : Create group of routers
  - r.NotFound : Set not found function
  - r.MethodNotAllowed : Set MethodNotAllowed function
  - r.OnError : Set panic handler function 
#### License
- MIT
