# fasthttp-staticrouter
Simple fasthttp static router not support params in url but fast

## Features

- Static http routing which simple fast

## Usage
`go get -u github.com/vinhjaxt/fasthttp-staticrouter`

```go
	r := router.New()

	api := r.Group("/api")

	api.Use(func(c *router.Context) {
		c.Response.Header.Set("Content-Type", "application/json")
	})

	apiv1 := api.Group("/v1")

	apiv1.Use(func(c *router.Context) {
		c.Response.Header.Set("X-Hello", "Hello")
		c.SetData("usera", map[string]string{"Name": "Guest"})
	})

	// apiv1.Get

	apiv1.Any("/", func(c *router.Context) {
		user, ok := c.GetData("user").(map[string]string)
		name, ok := user["Name"]
		if !ok {
			c.Error("Error", fasthttp.StatusInternalServerError)
			return
		}
		c.SetBodyString("\"Hello " + name + "\"")
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

  - c.Abort : Abort next handler
  - c.SetData : Set data to current context
  - c.GetData : Set data of current context

#### License
- MIT
