package routing

import (
	"net/http"

	"github.com/logitools/gw/web"
)

type Router interface {
	// ServeHTTP is invoked before invoking the route-matched handler's ServeHTTP
	// for every request regardless of the request url.
	// This can be easily implemented by embedding *http.ServeMux
	ServeHTTP(w http.ResponseWriter, r *http.Request)

	Handle(pattern string, handler http.Handler, handlerWrappers ...web.HandlerWrapper)
	HandleFunc(pattern string, handleFunc func(http.ResponseWriter, *http.Request), handlerWrappers ...web.HandlerWrapper)
	Group(prefix string, batch func(*RouteGroup), handlerWrappers ...web.HandlerWrapper) *RouteGroup
}
