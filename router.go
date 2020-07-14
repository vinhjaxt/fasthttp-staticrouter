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
		im, ia bool                                    // is Middlware, is Any
		h      func(*fasthttp.RequestCtx) (abort bool) // handle
		m      string                                  // method
	}

	handlerList struct {
		h []*handler
	}

	// Router struct
	Router struct {
		handlers                map[string]*handlerList
		RecoverHanlder          func(*fasthttp.RequestCtx)
		NotFoundHandler         func(*fasthttp.RequestCtx)
		MethodNotAllowedHandler func(*fasthttp.RequestCtx)
		middlewares             map[string][]*handler
	}

	// GroupRouter struct
	GroupRouter struct {
		router *Router
		path   string
	}
)

// StrRecoverPanic string
var StrRecoverPanic = "Recovered from panic:"

// RecoverHanlder is default RecoverHanlder
func RecoverHanlder(c *fasthttp.RequestCtx) {
	if r := recover(); r != nil {
		log.Println(StrRecoverPanic, r, debug.Stack())
		c.SetStatusCode(fasthttp.StatusInternalServerError)
	}
}

// NotFoundHandler is default NotFoundHandler
func NotFoundHandler(c *fasthttp.RequestCtx) {
	c.SetStatusCode(fasthttp.StatusNotFound)
}

// MethodNotAllowedHandler is default MethodNotAllowedHandler
func MethodNotAllowedHandler(c *fasthttp.RequestCtx) {
	c.SetStatusCode(fasthttp.StatusMethodNotAllowed)
}

// New create a router
func New() (r *Router) {
	r = &Router{
		handlers:                map[string]*handlerList{},
		RecoverHanlder:          RecoverHanlder,
		NotFoundHandler:         NotFoundHandler,
		MethodNotAllowedHandler: MethodNotAllowedHandler,
		middlewares:             map[string][]*handler{},
	}
	return
}

// Handler ate request :v
func (r *Router) Handler(ctx *fasthttp.RequestCtx) {
	defer r.RecoverHanlder(ctx)
	handlers, ok := r.handlers[b2s(ctx.Path())]
	if !ok {
		r.NotFoundHandler(ctx)
		return
	}
	// I think the numbers of handlers <= 10, so I use loop instead of map
	var handle *handler
	method := b2s(ctx.Method())
	for _, handle = range handlers.h {
		// Is middleware?
		if handle.im {
			if handle.h(ctx) {
				return
			}
			continue
		}
		// Is "any" method or exactly method?
		if handle.ia || handle.m == method {
			handle.h(ctx)
			return
		}
	}
	r.MethodNotAllowedHandler(ctx)
}

// BuildHandler return the request handler
func (r *Router) BuildHandler() func(ctx *fasthttp.RequestCtx) {
	r.middlewares = nil
	return r.Handler
}

// Router
func (r *Router) add(path, method string, h func(*fasthttp.RequestCtx) (_ bool)) {
	var handle *handler
	if method == "*" {
		handle = &handler{
			ia: true,
			h:  h,
		}
	} else {
		handle = &handler{
			m: method,
			h: h,
		}
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

func (r *Router) addUse(path string, h func(*fasthttp.RequestCtx) (abort bool)) {
	middlewares := []*handler{}
	// add middlewares
	handle := &handler{
		im: true,
		h:  h,
	}
	middlewares = append(middlewares, handle)
	if handles, ok := r.middlewares[path]; ok {
		r.middlewares[path] = append(handles, handle)
	} else {
		r.middlewares[path] = []*handler{handle}
	}

	// find router match this path
	lpath := len(path)
	for k, h := range r.handlers {
		lk := len(k)
		if k == path || (lk >= lpath && k[0:lpath] == path && k[lpath] == '/') {
			// prepend middlewares
			h.h = append(middlewares, h.h...)
			r.handlers[k] = h
		}
	}
}

// Get method
func (r *Router) Get(path string, h func(*fasthttp.RequestCtx) (_ bool)) {
	r.add(path, "GET", h)
}

// Post method
func (r *Router) Post(path string, h func(*fasthttp.RequestCtx) (_ bool)) {
	r.add(path, "POST", h)
}

// Put method
func (r *Router) Put(path string, h func(*fasthttp.RequestCtx) (_ bool)) {
	r.add(path, "PUT", h)
}

// Patch method
func (r *Router) Patch(path string, h func(*fasthttp.RequestCtx) (_ bool)) {
	r.add(path, "PATCH", h)
}

// Delete method
func (r *Router) Delete(path string, h func(*fasthttp.RequestCtx) (_ bool)) {
	r.add(path, "DELETE", h)
}

// Options method
func (r *Router) Options(path string, h func(*fasthttp.RequestCtx) (_ bool)) {
	r.add(path, "OPTIONS", h)
}

// Head method
func (r *Router) Head(path string, h func(*fasthttp.RequestCtx) (_ bool)) {
	r.add(path, "HEAD", h)
}

// Method specific
func (r *Router) Method(path, method string, h func(*fasthttp.RequestCtx) (_ bool)) {
	r.add(path, method, h)
}

// Any method
func (r *Router) Any(path string, h func(*fasthttp.RequestCtx) (_ bool)) {
	r.add(path, "*", h)
}

// Use middlewares
func (r *Router) Use(h func(*fasthttp.RequestCtx) (abort bool)) {
	r.addUse("", h)
}

// NotFound handler
func (r *Router) NotFound(h func(*fasthttp.RequestCtx)) {
	r.NotFoundHandler = h
}

// MethodNotAllowed handler
func (r *Router) MethodNotAllowed(h func(*fasthttp.RequestCtx)) {
	r.MethodNotAllowedHandler = h
}

// Group make a group of routers
func (r *Router) Group(path string) (g *GroupRouter) {
	g = &GroupRouter{
		path:   path,
		router: r,
	}
	return
}

// OnError is an alias of r.Recover()
func (r *Router) OnError(h func(*fasthttp.RequestCtx)) {
	r.Recover(h)
}

// Recover set RecoverHanlder
func (r *Router) Recover(h func(*fasthttp.RequestCtx)) {
	r.RecoverHanlder = h
}

// GroupRouter

// Get method
func (g *GroupRouter) Get(path string, h func(*fasthttp.RequestCtx) (_ bool)) {
	g.router.add(g.path+path, "GET", h)
}

// Post method
func (g *GroupRouter) Post(path string, h func(*fasthttp.RequestCtx) (_ bool)) {
	g.router.add(g.path+path, "POST", h)
}

// Put method
func (g *GroupRouter) Put(path string, h func(*fasthttp.RequestCtx) (_ bool)) {
	g.router.add(g.path+path, "PUT", h)
}

// Patch method
func (g *GroupRouter) Patch(path string, h func(*fasthttp.RequestCtx) (_ bool)) {
	g.router.add(g.path+path, "PATCH", h)
}

// Delete method
func (g *GroupRouter) Delete(path string, h func(*fasthttp.RequestCtx) (_ bool)) {
	g.router.add(g.path+path, "DELETE", h)
}

// Options method
func (g *GroupRouter) Options(path string, h func(*fasthttp.RequestCtx) (_ bool)) {
	g.router.add(g.path+path, "OPTIONS", h)
}

// Head method
func (g *GroupRouter) Head(path string, h func(*fasthttp.RequestCtx) (_ bool)) {
	g.router.add(g.path+path, "HEAD", h)
}

// Method specific
func (g *GroupRouter) Method(path, method string, h func(*fasthttp.RequestCtx) (_ bool)) {
	g.router.add(g.path+path, method, h)
}

// Any method
func (g *GroupRouter) Any(path string, h func(*fasthttp.RequestCtx) (_ bool)) {
	g.router.add(g.path+path, "*", h)
}

// Use middlewares
func (g *GroupRouter) Use(h func(*fasthttp.RequestCtx) (abort bool)) {
	g.router.addUse(g.path, h)
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
