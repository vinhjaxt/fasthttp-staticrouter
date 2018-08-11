# fasthttp-staticrouter
Simple fasthttp static router not support params in url but fast

## Features

- Static http routing which simple fast

## Usage
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

#### License
- MIT
