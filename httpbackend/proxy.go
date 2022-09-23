package httpbackend

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/velmie/alternea/app"
)

type ProxyHandler struct {
	proxyHandler http.Handler
	targetURL    *url.URL
	log          app.Logger
}

func NewProxyHandler(proxyHandler http.Handler, targetURL *url.URL, log app.Logger) *ProxyHandler {
	return &ProxyHandler{
		proxyHandler,
		targetURL,
		log,
	}
}

func (h *ProxyHandler) HandleRequest(request *http.Request) (*Response, error) {
	start := time.Now()
	proxyWriter := NewProxyResponseWriter()
	request.Host = h.targetURL.Host

	encoding := request.Header.Get("Accept-Encoding")
	// prevents encoding requested by client
	request.Header.Del("Accept-Encoding")

	h.proxyHandler.ServeHTTP(proxyWriter, request)
	if encoding != "" {
		request.Header.Set("Accept-Encoding", encoding)
	}

	response := proxyWriter.Response()
	response.Request = request

	if h.log.Level() >= app.InfoLevel {
		h.log.Infof(
			"httpbackend.proxyHandler: [%s] %s %s served in %s",
			http.StatusText(response.StatusCode),
			request.Method,
			request.URL,
			time.Since(start),
		)
	}

	return response, nil
}

// ProxyResponseWriter helps to create Response by using it in the http.Handler
type ProxyResponseWriter struct {
	response *Response
	writer   io.Writer
}

func NewProxyResponseWriter() *ProxyResponseWriter {
	const defaultBufferSize = 4096
	buf := bytes.NewBuffer(make([]byte, 0, defaultBufferSize))
	response := &Response{
		Body:   buf,
		Header: make(http.Header),
	}
	return &ProxyResponseWriter{
		response: response,
		writer:   buf,
	}
}

func (p *ProxyResponseWriter) Header() http.Header {
	return p.response.Header
}

func (p *ProxyResponseWriter) Write(bytes []byte) (int, error) {
	return p.writer.Write(bytes)
}

func (p *ProxyResponseWriter) WriteHeader(statusCode int) {
	p.response.StatusCode = statusCode
}

func (p *ProxyResponseWriter) Response() *Response {
	return p.response
}
