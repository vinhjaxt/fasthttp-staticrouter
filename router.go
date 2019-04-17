// Copyright 2018 vinhjaxt. All rights reserved.
// license that can be found in the LICENSE file.

package router

import (
	"log"
	"runtime/debug"
	"sync"
	"sync/atomic"

	"github.com/valyala/fasthttp" // faster than net/http
)

type (
	handler struct {
		m string
		h func(*Context)
	}

	handlerList struct {
		n int
		h []*handler
	}

	// Context of request
	Context struct {
		*fasthttp.RequestCtx
		abort int32
	}

	// Router struct
	Router struct {
		handlers                 map[string]*handlerList // final map
		middlewares              map[string][]*handler   // final map
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

var poolNew = func() interface{} {
	return &Context{}
}
var pool = sync.Pool{
	New: poolNew,
} // Context pool

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

const (
	notFoundFlag         = 0x01
	methodNotAllowedFlag = 0x02
)

// New create a router
func New() (r *Router) {
	r = &Router{
		handlers:                 map[string]*handlerList{},
		middlewares:              map[string][]*handler{},
		recoverFunction:          recoverFunction,
		notFoundFuction:          notFoundFuction,
		methodNotAllowedFunction: methodNotAllowedFunction,
	}
	return
}

// BuildHandler to pass to fasthttp
func (r *Router) BuildHandler() (h func(ctx *fasthttp.RequestCtx)) {
	h = buildHandler(r.handlers, r.recoverFunction, r.notFoundFuction, r.methodNotAllowedFunction)
	r.handlers = nil
	r.middlewares = nil
	r.recoverFunction = nil
	r.notFoundFuction = nil
	r.methodNotAllowedFunction = nil
	return
}

func buildHandler(handlerMap map[string]*handlerList, recoverFunction func(*Context), notFoundFuction func(*Context), methodNotAllowedFunction func(*Context)) func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		path := string(ctx.Path())
		method := string(ctx.Method())
		status := notFoundFlag
		c := pool.Get().(*Context)
		c.RequestCtx = ctx
		atomic.StoreInt32(&c.abort, 0)
		defer pool.Put(c)
		defer recoverFunction(c)
		if handlers, ok := handlerMap[path]; ok {
			status = methodNotAllowedFlag
			hh := handlers.h
			l := handlers.n
			var i int // 0 by default
			// i think the number of handlers <= 10, so i use for loop array instead of map
			for i < l {
				handle := hh[i]
				m := handle.m
				h := handle.h
				if m == "" {
					// middleware
					h(c)
					if atomic.LoadInt32(&c.abort) == 1 {
						return
					}
				} else if m == "*" || m == method {
					// handler
					status = 0
					h(c)
					if atomic.LoadInt32(&c.abort) == 1 {
						return
					}
				}
				i++
			}
		}
		if status == 0 {
			return
		} else if status == methodNotAllowedFlag {
			methodNotAllowedFunction(c)
		} else if status == notFoundFlag {
			notFoundFuction(c)
		}
	}
}

// Router
func (r *Router) add(path, method string, hh ...func(*Context)) {
	for _, h := range hh {
		handle := &handler{
			m: method,
			h: h,
		}
		if handlers, ok := r.handlers[path]; ok {
			// path exist, middlewares also added
			handlers.h = append(handlers.h, handle)
			handlers.n++
		} else {
			// path not exist, add middlewares handlers
			hm := r.findMiddlewares(path)
			hm = append(hm, handle)
			handlers := handlerList{
				n: len(hm),
				h: hm,
			}
			r.handlers[path] = &handlers
		}
	}
}

func (r *Router) findMiddlewares(path string) []*handler {
	hh := []*handler{}
	if h, ok := r.middlewares[path]; ok {
		hh = append(hh, h...)
	}
	// i := len(path)
	// for i > 0 {
	// 	i--
	// Priority father first
	i := 0
	for i < len(path) {
		if path[i] == '/' {
			mpath := path[0:i]
			if h, ok := r.middlewares[mpath]; ok {
				hh = append(hh, h...)
			}
		}
		i++
	}
	return hh
}

func (r *Router) addUse(path string, hh ...func(*Context)) {
	middlewares := []*handler{}
	// add middlewares
	for _, h := range hh {
		handle := &handler{
			m: "",
			h: h,
		}
		middlewares = append(middlewares, handle)
		if handles, ok := r.middlewares[path]; ok {
			r.middlewares[path] = append(handles, handle)
		} else {
			r.middlewares[path] = []*handler{handle}
		}
	}

	// find router match this path
	lpath := len(path)
	for k, h := range r.handlers {
		lk := len(k)
		if k == path || (lk >= lpath && k[0:lpath] == path && k[lpath] == '/') {
			// append middlewares
			h.h = append(h.h, middlewares...)
			h.n += len(middlewares)
			r.handlers[k] = h
		}
	}
}

// Get method
func (r *Router) Get(path string, h ...func(*Context)) {
	r.add(path, "GET", h...)
}

// Post method
func (r *Router) Post(path string, h ...func(*Context)) {
	r.add(path, "POST", h...)
}

// Put method
func (r *Router) Put(path string, h ...func(*Context)) {
	r.add(path, "PUT", h...)
}

// Patch method
func (r *Router) Patch(path string, h ...func(*Context)) {
	r.add(path, "PATCH", h...)
}

// Delete method
func (r *Router) Delete(path string, h ...func(*Context)) {
	r.add(path, "DELETE", h...)
}

// Options method
func (r *Router) Options(path string, h ...func(*Context)) {
	r.add(path, "OPTIONS", h...)
}

// Head method
func (r *Router) Head(path string, h ...func(*Context)) {
	r.add(path, "HEAD", h...)
}

// Any method
func (r *Router) Any(path string, h ...func(*Context)) {
	r.add(path, "*", h...)
}

// Use middlewares
func (r *Router) Use(h ...func(*Context)) {
	r.addUse("", h...)
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
	g.router.add(g.path+path, "GET", h...)
}

// Post method
func (g *GroupRouter) Post(path string, h ...func(*Context)) {
	g.router.add(g.path+path, "POST", h...)
}

// Put method
func (g *GroupRouter) Put(path string, h ...func(*Context)) {
	g.router.add(g.path+path, "PUT", h...)
}

// Patch method
func (g *GroupRouter) Patch(path string, h ...func(*Context)) {
	g.router.add(g.path+path, "PATCH", h...)
}

// Delete method
func (g *GroupRouter) Delete(path string, h ...func(*Context)) {
	g.router.add(g.path+path, "DELETE", h...)
}

// Options method
func (g *GroupRouter) Options(path string, h ...func(*Context)) {
	g.router.add(g.path+path, "OPTIONS", h...)
}

// Head method
func (g *GroupRouter) Head(path string, h ...func(*Context)) {
	g.router.add(g.path+path, "HEAD", h...)
}

// Any method
func (g *GroupRouter) Any(path string, h ...func(*Context)) {
	g.router.add(g.path+path, "*", h...)
}

// Use middlewares
func (g *GroupRouter) Use(h ...func(*Context)) {
	g.router.addUse(g.path, h...)
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
	atomic.StoreInt32(&c.abort, 1)
}
