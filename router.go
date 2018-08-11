// Copyright 2018 vinhjaxt. All rights reserved.
// license that can be found in the LICENSE file.

package router

import (
	"log"
	"runtime/debug"
	"strings"
	"sync"

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
		handles                  map[string]map[string][]func(*Context)
		middlewares              map[string][]func(*Context)
		notFoundFuction          func(*Context)
		recoverFunction          func(*Context)
		methodNotAllowedFunction func(*Context)
	}

	// GroupRouter struct
	GroupRouter struct {
		router *Router
		path   string
	}
)

// Default handler functions
var recoverFunction = func(c *Context) {
	if r := recover(); r != nil {
		log.Println("Recovered from panic:", r, string(debug.Stack()))
		c.Error("Error", fasthttp.StatusInternalServerError)
	}
}
var notFoundFuction = func(c *Context) {
	c.Error("not found", fasthttp.StatusNotFound)
}
var methodNotAllowedFunction = func(c *Context) {
	c.Error("request method not allowed", fasthttp.StatusMethodNotAllowed)
}

// New create a router
func New() (r *Router) {
	r = &Router{
		handles:                  map[string]map[string][]func(*Context){},
		middlewares:              map[string][]func(*Context){},
		recoverFunction:          recoverFunction,
		notFoundFuction:          notFoundFuction,
		methodNotAllowedFunction: methodNotAllowedFunction,
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

// Handler need to pass to fasthttp
func (r *Router) Handler(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Path())
	method := string(ctx.Method())
	c := r.pool.Get().(*Context)
	defer r.pool.Put(c) // this will be called after recoverFunction called
	c.RequestCtx = ctx
	c.Store = nil
	c.abort = false
	defer r.recoverFunction(c)
	notFound := true
	methodNotAllowed := false

	// middlewares
	for p, mhandles := range r.middlewares {
		lenP := len(p)
		if strings.HasPrefix(path, p) && (lenP == len(path) || path[lenP] == '/') {
			for _, handle := range mhandles {
				handle(c)
				if c.abort {
					goto abort
				}
			}
		}
	}

	if handles, ok := r.handles[path]; ok {
		methodNotAllowed = true

		// any method
		if mhandles, ok := handles["*"]; ok {
			methodNotAllowed = false
			for _, handle := range mhandles {
				handle(c)
				notFound = false
				if c.abort {
					goto abort
				}
			}
		}

		// methods
		if mhandles, ok := handles[method]; ok {
			methodNotAllowed = false
			for _, handle := range mhandles {
				handle(c)
				notFound = false
				if c.abort {
					goto abort
				}
			}
		}
	}

	// path exist, but no method found
	if methodNotAllowed {
		r.methodNotAllowedFunction(c)
		goto abort
	}

	// path not found
	if notFound {
		if c.abort == false {
			r.notFoundFuction(c)
		}
	}
abort:
}

// Router
func (r *Router) add(path, method string, h []func(*Context)) {
	mhandles, ok := r.handles[path]
	if !ok {
		mhandles = map[string][]func(*Context){}
		r.handles[path] = mhandles
	}
	handles, ok := mhandles[method]
	if ok {
		r.handles[path][method] = append(handles, h...)
	} else {
		r.handles[path][method] = h
	}
}

// Get method
func (r *Router) Get(path string, h ...func(*Context)) {
	r.add(path, "GET", h)
}

// Post method
func (r *Router) Post(path string, h ...func(*Context)) {
	r.add(path, "POST", h)
}

// Put method
func (r *Router) Put(path string, h ...func(*Context)) {
	r.add(path, "PUT", h)
}

// Patch method
func (r *Router) Patch(path string, h ...func(*Context)) {
	r.add(path, "PATCH", h)
}

// Delete method
func (r *Router) Delete(path string, h ...func(*Context)) {
	r.add(path, "DELETE", h)
}

// Options method
func (r *Router) Options(path string, h ...func(*Context)) {
	r.add(path, "OPTIONS", h)
}

// Head method
func (r *Router) Head(path string, h ...func(*Context)) {
	r.add(path, "HEAD", h)
}

// Any method
func (r *Router) Any(path string, h ...func(*Context)) {
	r.add(path, "*", h)
}

// Use middlewares
func (r *Router) Use(h ...func(*Context)) {
	path := ""
	handles, ok := r.middlewares[path]
	if ok {
		r.middlewares[path] = append(handles, h...)
	} else {
		r.middlewares[path] = h
	}
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
func (g *GroupRouter) Get(path string, h ...func(*Context)) {
	g.router.add(g.path+path, "GET", h)
}

// Post method
func (g *GroupRouter) Post(path string, h ...func(*Context)) {
	g.router.add(g.path+path, "POST", h)
}

// Put method
func (g *GroupRouter) Put(path string, h ...func(*Context)) {
	g.router.add(g.path+path, "PUT", h)
}

// Patch method
func (g *GroupRouter) Patch(path string, h ...func(*Context)) {
	g.router.add(g.path+path, "PATCH", h)
}

// Delete method
func (g *GroupRouter) Delete(path string, h ...func(*Context)) {
	g.router.add(g.path+path, "DELETE", h)
}

// Options method
func (g *GroupRouter) Options(path string, h ...func(*Context)) {
	g.router.add(g.path+path, "OPTIONS", h)
}

// Head method
func (g *GroupRouter) Head(path string, h ...func(*Context)) {
	g.router.add(g.path+path, "HEAD", h)
}

// Any method
func (g *GroupRouter) Any(path string, h ...func(*Context)) {
	g.router.add(g.path+path, "*", h)
}

// Use middlewares
func (g *GroupRouter) Use(h ...func(*Context)) {
	path := g.path
	handles, ok := g.router.middlewares[path]
	if ok {
		g.router.middlewares[path] = append(handles, h...)
	} else {
		g.router.middlewares[path] = h
	}
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
	if c.Store == nil {
		return nil
	}
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
