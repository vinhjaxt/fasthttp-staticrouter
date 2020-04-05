// Copyright 2018 vinhjaxt. All rights reserved.
// license that can be found in the LICENSE file.

package router

import (
	"log"
	"reflect"
	"runtime/debug"
	"unsafe"

	"github.com/valyala/fasthttp"
)

type (
	handler struct {
		m string
		h func(*fasthttp.RequestCtx) (abort bool)
	}

	handlerList struct {
		h []*handler
	}

	// Router struct
	Router struct {
		handlers                 map[string]*handlerList // final map
		middlewares              map[string][]*handler   // final map
		notFoundFuction          func(*fasthttp.RequestCtx)
		recoverFunction          func(*fasthttp.RequestCtx)
		methodNotAllowedFunction func(*fasthttp.RequestCtx)
	}

	// GroupRouter struct
	GroupRouter struct {
		router *Router
		path   string
	}
)

// Default handler functions
func recoverFunction(c *fasthttp.RequestCtx) {
	if r := recover(); r != nil {
		log.Printf("Recovered from panic: %v %s\n", r, debug.Stack())
		c.Error("Error", fasthttp.StatusInternalServerError)
	}
}

func notFoundFuction(c *fasthttp.RequestCtx) {
	c.Error("not found", fasthttp.StatusNotFound)
}

func methodNotAllowedFunction(c *fasthttp.RequestCtx) {
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
	return func(ctx *fasthttp.RequestCtx) {
		status := notFoundFlag
		defer r.recoverFunction(ctx)
		if handlers, ok := r.handlers[b2s(ctx.Path())]; ok {
			status = methodNotAllowedFlag
			// i think the number of handlers <= 10, so i use for loop array instead of map
			for _, handle := range handlers.h {
				if len(handle.m) == 0 {
					// middleware
					if handle.h(ctx) {
						return
					}
				} else if handle.m == "*" || handle.m == b2s(ctx.Method()) {
					// handler
					status = 0
					if handle.h(ctx) {
						return
					}
				}
			}
		}
		if status == 0 {
			return
		} else if status == methodNotAllowedFlag {
			r.methodNotAllowedFunction(ctx)
		} else if status == notFoundFlag {
			r.notFoundFuction(ctx)
		}
	}
}

// Router
func (r *Router) add(path, method string, hh ...func(*fasthttp.RequestCtx) (abort bool)) {
	for _, h := range hh {
		handle := &handler{
			m: method,
			h: h,
		}
		if handlers, ok := r.handlers[path]; ok {
			// path exist, middlewares also added
			handlers.h = append(handlers.h, handle)
		} else {
			// path not exist, add middlewares handlers
			hm := r.findMiddlewares(path)
			hm = append(hm, handle)
			r.handlers[path] = &handlerList{
				h: hm,
			}
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

func (r *Router) addUse(path string, hh ...func(*fasthttp.RequestCtx) (abort bool)) {
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
			r.handlers[k] = h
		}
	}
}

// Get method
func (r *Router) Get(path string, h ...func(*fasthttp.RequestCtx) (abort bool)) {
	r.add(path, "GET", h...)
}

// Post method
func (r *Router) Post(path string, h ...func(*fasthttp.RequestCtx) (abort bool)) {
	r.add(path, "POST", h...)
}

// Put method
func (r *Router) Put(path string, h ...func(*fasthttp.RequestCtx) (abort bool)) {
	r.add(path, "PUT", h...)
}

// Patch method
func (r *Router) Patch(path string, h ...func(*fasthttp.RequestCtx) (abort bool)) {
	r.add(path, "PATCH", h...)
}

// Delete method
func (r *Router) Delete(path string, h ...func(*fasthttp.RequestCtx) (abort bool)) {
	r.add(path, "DELETE", h...)
}

// Options method
func (r *Router) Options(path string, h ...func(*fasthttp.RequestCtx) (abort bool)) {
	r.add(path, "OPTIONS", h...)
}

// Head method
func (r *Router) Head(path string, h ...func(*fasthttp.RequestCtx) (abort bool)) {
	r.add(path, "HEAD", h...)
}

// Any method
func (r *Router) Any(path string, h ...func(*fasthttp.RequestCtx) (abort bool)) {
	r.add(path, "*", h...)
}

// Use middlewares
func (r *Router) Use(h ...func(*fasthttp.RequestCtx) (abort bool)) {
	r.addUse("", h...)
}

// NotFound handler
func (r *Router) NotFound(h func(*fasthttp.RequestCtx)) {
	r.notFoundFuction = h
}

// MethodNotAllowed handler
func (r *Router) MethodNotAllowed(h func(*fasthttp.RequestCtx)) {
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
func (r *Router) OnError(h func(*fasthttp.RequestCtx)) {
	r.recoverFunction = h
}

// GroupRouter

// Get method
func (g *GroupRouter) Get(path string, h ...func(*fasthttp.RequestCtx) (abort bool)) {
	g.router.add(g.path+path, "GET", h...)
}

// Post method
func (g *GroupRouter) Post(path string, h ...func(*fasthttp.RequestCtx) (abort bool)) {
	g.router.add(g.path+path, "POST", h...)
}

// Put method
func (g *GroupRouter) Put(path string, h ...func(*fasthttp.RequestCtx) (abort bool)) {
	g.router.add(g.path+path, "PUT", h...)
}

// Patch method
func (g *GroupRouter) Patch(path string, h ...func(*fasthttp.RequestCtx) (abort bool)) {
	g.router.add(g.path+path, "PATCH", h...)
}

// Delete method
func (g *GroupRouter) Delete(path string, h ...func(*fasthttp.RequestCtx) (abort bool)) {
	g.router.add(g.path+path, "DELETE", h...)
}

// Options method
func (g *GroupRouter) Options(path string, h ...func(*fasthttp.RequestCtx) (abort bool)) {
	g.router.add(g.path+path, "OPTIONS", h...)
}

// Head method
func (g *GroupRouter) Head(path string, h ...func(*fasthttp.RequestCtx) (abort bool)) {
	g.router.add(g.path+path, "HEAD", h...)
}

// Any method
func (g *GroupRouter) Any(path string, h ...func(*fasthttp.RequestCtx) (abort bool)) {
	g.router.add(g.path+path, "*", h...)
}

// Use middlewares
func (g *GroupRouter) Use(h ...func(*fasthttp.RequestCtx) (abort bool)) {
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

// b2s converts byte slice to a string without memory allocation.
// See https://groups.google.com/forum/#!msg/Golang-Nuts/ENgbUzYvCuU/90yGx7GUAgAJ .
//
// Note it may break if string and/or slice header will change
// in the future go versions.
func b2s(b []byte) string {
	/* #nosec G103 */
	return *(*string)(unsafe.Pointer(&b))
}

// s2b converts string to a byte slice without memory allocation.
//
// Note it may break if string and/or slice header will change
// in the future go versions.
func s2b(s string) (b []byte) {
	/* #nosec G103 */
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	/* #nosec G103 */
	sh := *(*reflect.StringHeader)(unsafe.Pointer(&s))
	bh.Data = sh.Data
	bh.Len = sh.Len
	bh.Cap = sh.Len
	return b
}
