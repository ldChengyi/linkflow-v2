package server

import (
	"log/slog"
	"net/http"
	"strings"
)

type registeredRoute struct {
	Method  string
	Pattern string
}

type routeRecorder struct {
	mux    *http.ServeMux
	routes []registeredRoute
}

func newRouteRecorder(mux *http.ServeMux) *routeRecorder {
	return &routeRecorder{mux: mux}
}

func (r *routeRecorder) Handle(pattern string, handler http.Handler) {
	r.routes = append(r.routes, parseRoutePattern(pattern))
	r.mux.Handle(pattern, handler)
}

func (r *routeRecorder) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	r.routes = append(r.routes, parseRoutePattern(pattern))
	r.mux.HandleFunc(pattern, handler)
}

func (r *routeRecorder) Log(log *slog.Logger) {
	if log == nil {
		log = slog.Default()
	}
	for _, route := range r.routes {
		log.Info("registered http route", "method", route.Method, "pattern", route.Pattern)
	}
}

func parseRoutePattern(pattern string) registeredRoute {
	method, path, ok := strings.Cut(pattern, " ")
	if !ok {
		return registeredRoute{Method: "*", Pattern: pattern}
	}
	return registeredRoute{Method: method, Pattern: path}
}
