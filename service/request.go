package service

import (
	"context"
	"io"
	"net/http"
)

type (
	RequestHandler interface {
		Handle(ctx context.Context, w io.Writer, request *http.Request) error
	}
	// RequestIterator is used in order to produce required requests to a backend
	RequestIterator interface {
		Next(prevRequest *http.Request, prevResponseData []byte) (*http.Request, error)
	}
)

type DirectRequestIterator struct {
}

func NewDirectRequestIterator() *DirectRequestIterator {
	return &DirectRequestIterator{}
}

func (d *DirectRequestIterator) Next(
	prevRequest *http.Request,
	prevResponseData []byte,
) (*http.Request, error) {
	if prevResponseData == nil {
		return prevRequest, nil
	}
	return nil, nil
}
