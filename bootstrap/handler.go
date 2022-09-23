package bootstrap

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/velmie/alternea/httpbackend"
	"github.com/velmie/alternea/manipulation"
	"github.com/velmie/alternea/route"
	"github.com/velmie/alternea/service"

	"github.com/pkg/errors"
)

type HTTPRoutesInitializer interface {
	InitRoutes(router route.Router, config *ServerConfig) error
}

type HTTPRoutesInitializers []HTTPRoutesInitializer

func (h HTTPRoutesInitializers) InitRoutes(router route.Router, config *ServerConfig) error {
	for _, initializer := range h {
		if err := initializer.InitRoutes(router, config); err != nil {
			return err
		}
	}
	return nil
}

type ProxyRoutesInitializer struct {
	backendErrorHandler func(err error) (proceed bool)
	errorHandler        func(err error, w http.ResponseWriter, r *http.Request)
}

func NewProxyRoutesInitializer(
	backendErrorHandler func(err error) (proceed bool),
	errorHandler func(err error, w http.ResponseWriter, r *http.Request),
) *ProxyRoutesInitializer {
	return &ProxyRoutesInitializer{
		backendErrorHandler,
		errorHandler,
	}
}

func (p *ProxyRoutesInitializer) InitRoutes(router route.Router, config *ServerConfig) error {
	for _, srv := range config.ProxyServices {
		if err := p.setHandler(srv, router); err != nil {
			return err
		}
	}
	return nil
}

func (p *ProxyRoutesInitializer) setHandler(
	srv *ProxyServiceConfig,
	router route.Router,
) error {
	method := http.MethodGet
	if srv.Method != "" {
		method = srv.Method
	}
	targetURL, err := url.Parse(srv.Backend.TargetURL)
	if err != nil {
		return errors.Wrap(err, "ProxyRoutesInitializer: cannot parse target url")
	}

	transformer, err := p.createTransformer(&srv.Transformer)
	if err != nil {
		return err
	}

	requestIterator := p.requestIterator()

	targetPathTemplate := route.ColonParamsReplaceTemplate(targetURL.Path)

	// path will be added by the route.PathSubstitution
	targetURL.Path = ""

	// TODO: add client options
	reverseProxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxyBackend := httpbackend.NewProxyHandler(reverseProxy, targetURL, GetLogger())
	handlerConfig := &service.TransformerHandlerConfig{
		ErrorHandler: p.backendErrorHandler,
	}
	if len(srv.Backend.SuccessHTTPStatusCodes) > 0 {
		handlerConfig.SuccessHTTPStatusCodes = srv.Backend.SuccessHTTPStatusCodes
	}
	transformerHandler := service.NewTransformerHandler(
		transformer,
		requestIterator,
		proxyBackend,
		handlerConfig,
	)
	transformerHandler.SetFlushInterval(srv.FlushInterval)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for name, value := range srv.SetHeader {
			w.Header().Set(name, value)
		}
		err = transformerHandler.Handle(r.Context(), w, r)

		if err != nil {
			for name := range srv.SetHeader {
				w.Header().Del(name)
			}
			if p.errorHandler != nil {
				p.errorHandler(err, w, r)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
	})

	err = router.Handle(method, srv.PathTemplate, route.PathSubstitution(targetPathTemplate, handler))
	if err != nil {
		return errors.Wrap(err, "ProxyRoutesInitializer: cannot create new route")
	}
	return nil
}

func (p *ProxyRoutesInitializer) createTransformer(cfg *DynamicConfig) (manipulation.DataTransformer, error) {
	transformerCfg, err := cfg.ToConfig()
	if err != nil {
		return nil, errors.Wrap(err, "ProxyRoutesInitializer: cannot get transformer config")
	}
	return Transformer.Create(cfg.Name, transformerCfg)
}

func (p *ProxyRoutesInitializer) requestIterator() service.RequestIterator {
	return service.NewDirectRequestIterator()
}

func DefaultErrorHandler(log func(v ...any)) func(err error, w http.ResponseWriter, r *http.Request) {
	return func(err error, w http.ResponseWriter, r *http.Request) {
		log(fmt.Sprintf("%s %s, error: %s", r.Method, r.URL.Path, err))
		if httpErr, ok := errors.Cause(err).(*service.HTTPError); ok {
			w.WriteHeader(httpErr.StatusCode)
			if _, err := io.Copy(w, httpErr.Body); err != nil {
				log(fmt.Sprintf("failed to copy http error body: %s", err))
			}
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func DefaultBackendErrorHandler(log func(v ...any)) func(err error) (proceed bool) {
	return func(err error) (proceed bool) {
		log("backend returned error: ", err)
		return false
	}
}

type StaticContentRoutesInitializer struct {
}

func NewStaticContentRoutesInitializer() *StaticContentRoutesInitializer {
	return &StaticContentRoutesInitializer{}
}

func (s *StaticContentRoutesInitializer) InitRoutes(router route.Router, config *ServerConfig) error {
	for _, srv := range config.StaticServices {
		if err := s.setHandler(srv, router); err != nil {
			return err
		}
	}
	return nil
}

func (s *StaticContentRoutesInitializer) setHandler(
	srv *StaticServiceConfig,
	router route.Router,
) error {
	method := http.MethodGet
	if srv.Method != "" {
		method = srv.Method
	}
	var content []byte
	if srv.Content != "" {
		content = []byte(srv.Content)
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for name, value := range srv.SetHeader {
			w.Header().Set(name, value)
		}
		if srv.ResponseCode != 0 {
			w.WriteHeader(srv.ResponseCode)
		}
		if len(content) > 0 {
			_, _ = w.Write(content)
		}
	})
	return router.Handle(
		method,
		srv.PathTemplate,
		func(w http.ResponseWriter, r *http.Request, _ route.NamedPathParameters) {
			handler.ServeHTTP(w, r)
		},
	)
}
