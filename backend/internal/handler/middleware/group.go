package middleware

import (
	"net/http"
	"strings"
)

type Middleware func(http.Handler) http.Handler

type Group struct {
	mux         *http.ServeMux
	prefix      string
	middlewares []Middleware
}

// NewGroup 建立有 prefix 的群組
func NewGroup(mux *http.ServeMux, prefix string, middlewares ...Middleware) *Group {
	return &Group{mux: mux, prefix: prefix, middlewares: middlewares}
}

// NewGroupNoPrefix 建立無 prefix 的群組
func NewGroupNoPrefix(mux *http.ServeMux, middlewares ...Middleware) *Group {
	return &Group{mux: mux, middlewares: middlewares}
}

func (g *Group) Handle(pattern string, handler http.Handler) {
	g.mux.Handle(g.buildPattern(pattern), Chain(handler, g.middlewares...))
}

func (g *Group) HandleFunc(pattern string, fn func(http.ResponseWriter, *http.Request)) {
	g.Handle(pattern, http.HandlerFunc(fn))
}

// SubGroup Group 可以衍生子群組，繼承 prefix 與 middleware
func (g *Group) SubGroup(prefix string, middlewares ...Middleware) *Group {
	merged := make([]Middleware, len(g.middlewares), len(g.middlewares)+len(middlewares))
	copy(merged, g.middlewares)
	return &Group{
		mux:         g.mux,
		prefix:      g.prefix + prefix,
		middlewares: append(merged, middlewares...),
	}
}

// buildPattern 有 prefix 時插入 method 與 path 之間，無 prefix 直接回傳
func (g *Group) buildPattern(pattern string) string {
	if g.prefix == "" {
		return pattern
	}
	method, path, hasMethod := strings.Cut(pattern, " ")
	if !hasMethod {
		return g.prefix + pattern
	}
	return method + " " + g.prefix + strings.TrimSpace(path)
}

func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
