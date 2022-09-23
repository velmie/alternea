package httpbackend

import (
	"io"
	"net/http"
)

type (
	// Response represents the response from the backend
	Response struct {
		Request    *http.Request
		Body       io.Reader
		Header     http.Header
		StatusCode int
	}
	// RequestHandler represents an entity that can handle a request
	RequestHandler interface {
		HandleRequest(request *http.Request) (*Response, error)
	}
)
