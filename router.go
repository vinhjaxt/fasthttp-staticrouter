// Copyright 2018 vinhjaxt. All rights reserved.
// license that can be found in the LICENSE file.

package router

import (
	"log"
	"runtime/debug"
	"sync"
	"time"

	"github.com/valyala/fasthttp" // faster than net/http
)

type (
	// Context of request
	Context struct {
		*fasthttp.RequestCtx
		Store map[string]interface{}
		abort bool
	}

	// Router struct
	Router struct {
		pool                     sync.Pool // Context pool
		handles                  map[string]func(*Context)
		middlewares              map[string]func(*Context)
		notFoundFuction          func(*Context)
		recoverFunction          func(*Context)
		methodNotAllowedFunction func(*Context)
		cache                    map[string][]func(*Context)
	}

	// GroupRouter struct
	GroupRouter struct {
		router *Router
		path   string
	}
)

// Default handler functions
func recoverFunction(c *Context) {
	if r := recover(); r != nil {
		log.Println("Recovered from panic:", r, string(debug.Stack()))
		c.Error("Error", fasthttp.StatusInternalServerError)
	}
}

func notFoundFuction(c *Context) {
	c.Error("not found", fasthttp.StatusNotFound)
}

func methodNotAllowedFunction(c *Context) {
	c.Error("request method not allowed", fasthttp.StatusMethodNotAllowed)
}

var methods = [...]string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}

const (
	notFoundFlag         = 0x01
	methodNotAllowedFlag = 0x02
)

// New create a router
func New() (r *Router) {
	r = &Router{
		handles:                  map[string]func(*Context){},
		middlewares:              map[string]func(*Context){},
		recoverFunction:          recoverFunction,
		notFoundFuction:          notFoundFuction,
		methodNotAllowedFunction: methodNotAllowedFunction,
		cache:                    map[string][]func(*Context){},
	}
	r.pool.New = func() interface{} {
		return &Context{
			RequestCtx: nil,
			Store:      nil,
			abort:      false,
		}
	}
	return
}

func timeTrack(start int64, name string) {
	end := time.Now().UnixNano()
	log.Printf("%s start: %d, end: %d, diff: %dns", name, start, end, end-start)
}

// Handler need to pass to fasthttp
func (r *Router) Handler(ctx *fasthttp.RequestCtx) {
	// defer timeTrack(time.Now().UnixNano(), "Handler")
	var (
		ok      bool
		c       *Context
		status  int8
		i       int
		handle  func(*Context)
		handles []func(*Context)
		uPath   string
		mPath   string
	)
	path := string(ctx.Path())
	method := string(ctx.Method())
	mPath = method + path
	for i, uPath = range methods {
		if uPath == method {
			goto OK
		}
	}
	status = methodNotAllowedFlag // not found = false, methodNotAllowed = true
	goto methodNotSupported
OK:
	c = r.pool.Get().(*Context)
	defer r.pool.Put(c) // this will be called after recoverFunction called
	c.RequestCtx = ctx
	c.Store = nil
	c.abort = false
	defer r.recoverFunction(c)
	status = notFoundFlag // not found = true, methodNotAllowed = false

	// cache method,path => handles
	if handles, ok = r.cache[mPath]; ok {
		for i, handle = range handles {
			handle(c)
			if c.abort {
				goto abort
			}
		}
		goto abort
	}

	handles = []func(*Context){}

	// middlewares
	i = len(path)
	for i > 0 {
		i--
		if path[i] == '/' {
			uPath = path[0:i]
			if handle, ok = r.middlewares[uPath]; ok {
				handles = append(handles, handle)
				if c.abort == false {
					handle(c)
				}
			}
		}
	}
	// any method
	if handle, ok = r.handles["*"+path]; ok {
		handles = append(handles, handle)
		status = 0 // not found = false, methodNotAllowed = false
		if c.abort == false {
			handle(c)
		}
	}
	// methods
	if handle, ok = r.handles[mPath]; ok {
		handles = append(handles, handle)
		status = 0
		if c.abort == false {
			handle(c)
		}
	}

	if (status & notFoundFlag) == 0 {
		r.cache[mPath] = handles
	}

methodNotSupported:
	// path exist, but no method found
	if (status & methodNotAllowedFlag) != 0 {
		r.methodNotAllowedFunction(c)
		goto abort
	}

	// path not found
	if (status & notFoundFlag) != 0 {
		r.notFoundFuction(c)
	}
abort:
}

// Router
func (r *Router) add(path, method string, h func(*Context)) {
	r.handles[method+path] = h
}

// Get method
func (r *Router) Get(path string, h func(*Context)) {
	r.add(path, "GET", h)
}

// Post method
func (r *Router) Post(path string, h func(*Context)) {
	r.add(path, "POST", h)
}

// Put method
func (r *Router) Put(path string, h func(*Context)) {
	r.add(path, "PUT", h)
}

// Patch method
func (r *Router) Patch(path string, h func(*Context)) {
	r.add(path, "PATCH", h)
}

// Delete method
func (r *Router) Delete(path string, h func(*Context)) {
	r.add(path, "DELETE", h)
}

// Options method
func (r *Router) Options(path string, h func(*Context)) {
	r.add(path, "OPTIONS", h)
}

// Head method
func (r *Router) Head(path string, h func(*Context)) {
	r.add(path, "HEAD", h)
}

// Any method
func (r *Router) Any(path string, h func(*Context)) {
	r.add(path, "*", h)
}

// Use middlewares
func (r *Router) Use(h func(*Context)) {
	r.middlewares[""] = h
}

// NotFound handler
func (r *Router) NotFound(h func(*Context)) {
	r.notFoundFuction = h
}

// MethodNotAllowed handler
func (r *Router) MethodNotAllowed(h func(*Context)) {
	r.methodNotAllowedFunction = h
}

// Group make a group of routers
func (r *Router) Group(path string) (g *GroupRouter) {
	g = &GroupRouter{
		path:   path,
		router: r,
	}
	return
}

// OnError Recover from panic
func (r *Router) OnError(h func(*Context)) {
	r.recoverFunction = h
}

// GroupRouter

// Get method
func (g *GroupRouter) Get(path string, h func(*Context)) {
	g.router.add(g.path+path, "GET", h)
}

// Post method
func (g *GroupRouter) Post(path string, h func(*Context)) {
	g.router.add(g.path+path, "POST", h)
}

// Put method
func (g *GroupRouter) Put(path string, h func(*Context)) {
	g.router.add(g.path+path, "PUT", h)
}

// Patch method
func (g *GroupRouter) Patch(path string, h func(*Context)) {
	g.router.add(g.path+path, "PATCH", h)
}

// Delete method
func (g *GroupRouter) Delete(path string, h func(*Context)) {
	g.router.add(g.path+path, "DELETE", h)
}

// Options method
func (g *GroupRouter) Options(path string, h func(*Context)) {
	g.router.add(g.path+path, "OPTIONS", h)
}

// Head method
func (g *GroupRouter) Head(path string, h func(*Context)) {
	g.router.add(g.path+path, "HEAD", h)
}

// Any method
func (g *GroupRouter) Any(path string, h func(*Context)) {
	g.router.add(g.path+path, "*", h)
}

// Use middlewares
func (g *GroupRouter) Use(h func(*Context)) {
	g.router.middlewares[g.path] = h
}

// Group create another group
func (g *GroupRouter) Group(path string) (cg *GroupRouter) {
	cg = &GroupRouter{
		path:   g.path + path,
		router: g.router,
	}
	return
}

// Context

// Abort next handler from context
func (c *Context) Abort() {
	c.abort = true
}

// GetData from other handler
func (c *Context) GetData(name string) interface{} {
	if r, ok := c.Store[name]; ok {
		return r
	}
	return nil
}

// SetData from other handler
func (c *Context) SetData(name string, data interface{}) {
	if c.Store == nil {
		c.Store = map[string]interface{}{}
	}
	c.Store[name] = data
}
