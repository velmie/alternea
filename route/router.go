package route

import (
	"fmt"
	"net/http"
)

type Handler func(w http.ResponseWriter, r *http.Request, params NamedPathParameters)

type Router interface {
	GET(path string, h Handler) error
	POST(path string, h Handler) error
	DELETE(path string, h Handler) error
	PATCH(path string, h Handler) error
	PUT(path string, h Handler) error
	Handle(method, path string, h Handler) error
	Match(method, path string) http.Handler
	http.Handler
}

type MatcherFactory func(template string) (PathMatcher, error)

type route struct {
	template string
	matcher  PathMatcher
	handler  Handler
}

type DefaultRouter struct {
	createMatcher   MatcherFactory
	routes          map[string][]*route
	notFoundHandler http.Handler
}

func NewDefaultRouter(
	matcherFactory MatcherFactory,
) *DefaultRouter {
	return &DefaultRouter{
		matcherFactory,
		make(map[string][]*route),
		http.HandlerFunc(notFoundHandler),
	}
}

func (r *DefaultRouter) Match(method, path string) http.Handler {
	rt := r.match(method, path)
	if rt == nil {
		return r.notFoundHandler
	}
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		rt.handler(w, req, rt.matcher.RetrieveParameters(req.URL.Path))
	})
}

func (r *DefaultRouter) GET(path string, h Handler) error {
	return r.Handle(http.MethodGet, path, h)
}

func (r *DefaultRouter) POST(path string, h Handler) error {
	return r.Handle(http.MethodPost, path, h)
}

func (r *DefaultRouter) DELETE(path string, h Handler) error {
	return r.Handle(http.MethodDelete, path, h)
}

func (r *DefaultRouter) PATCH(path string, h Handler) error {
	return r.Handle(http.MethodPatch, path, h)
}

func (r *DefaultRouter) PUT(path string, h Handler) error {
	return r.Handle(http.MethodPut, path, h)
}

var allowedMethods = []string{
	http.MethodGet,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
}

func (r *DefaultRouter) Handle(method, path string, handler Handler) error {
	allowed := false
	for _, allowedMethod := range allowedMethods {
		if allowedMethod == method {
			allowed = true
			break
		}
	}
	if !allowed {
		return fmt.Errorf("method %s could not be processed", method)
	}

	if rt := r.match(method, path); rt != nil {
		return fmt.Errorf(
			"handler is already registered for path %s %s, it matches the path %s",
			method,
			path,
			rt.template,
		)
	}

	matcher, err := r.createMatcher(path)
	if err != nil {
		return err
	}
	r.routes[method] = append(r.routes[method], &route{
		matcher:  matcher,
		template: path,
		handler:  handler,
	})

	return nil
}

func (r *DefaultRouter) SetNotFoundHandler(h http.Handler) {
	r.notFoundHandler = h
}

func (r *DefaultRouter) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	r.Match(request.Method, request.URL.Path).ServeHTTP(writer, request)
}

func (r *DefaultRouter) match(method, path string) *route {
	for _, rt := range r.routes[method] {
		if rt.matcher.Match(path) {
			return rt
		}
	}
	return nil
}

func notFoundHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
