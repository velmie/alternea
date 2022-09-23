package service

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/velmie/alternea/httpbackend"
	"github.com/velmie/alternea/manipulation"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type TransformerHandler struct {
	transformer            manipulation.DataTransformer
	requestIterator        RequestIterator
	backend                httpbackend.RequestHandler
	errorHandler           func(err error) (proceed bool)
	flushInterval          time.Duration
	successHTTPStatusCodes []int
}

type TransformerHandlerConfig struct {
	ErrorHandler           func(err error) (proceed bool)
	SuccessHTTPStatusCodes []int
}

func NewTransformerHandler(
	transformer manipulation.DataTransformer,
	requestIterator RequestIterator,
	backend httpbackend.RequestHandler,
	config ...*TransformerHandlerConfig,
) *TransformerHandler {
	handler := &TransformerHandler{
		transformer:     transformer,
		requestIterator: requestIterator,
		backend:         backend,
	}
	if len(config) > 0 {
		handler.errorHandler = config[0].ErrorHandler
		handler.successHTTPStatusCodes = config[0].SuccessHTTPStatusCodes
	}
	if len(handler.successHTTPStatusCodes) == 0 {
		handler.successHTTPStatusCodes = []int{http.StatusOK}
	}
	return handler
}

func (h *TransformerHandler) Handle(
	ctx context.Context,
	w io.Writer,
	request *http.Request,
) error {
	transformerChanel := make(chan []byte)

	// net/http/httputil/reverseproxy.go
	if h.flushInterval != 0 {
		if wf, ok := w.(writeFlusher); ok {
			mlw := &maxLatencyWriter{
				dst:     wf,
				latency: h.flushInterval,
			}
			defer mlw.stop()

			// set up initial timer so headers get flushed even if body writes are delayed
			mlw.flushPending = true
			mlw.t = time.AfterFunc(h.flushInterval, mlw.delayedFlush)

			w = mlw
		}
	}

	errs, ctx := errgroup.WithContext(ctx)
	errs.Go(func() error {
		err := h.transformer.Transform(ctx, transformerChanel, w)
		return err
	})

	var (
		response *httpbackend.Response
		data     []byte
		err      error
	)
	for {
		request, err = h.requestIterator.Next(request, data)
		if err != nil {
			return errors.Wrap(err, "TransformerHandler: cannot get request from iterator")
		}
		if request == nil {
			close(transformerChanel)
			break
		}
		response, err = h.backend.HandleRequest(request)
		if err != nil && !h.errorHandler(err) {
			close(transformerChanel)
			return errors.Wrap(err, "TransformerHandler: cannot get response from backend")
		}
		success := false
		for _, successCode := range h.successHTTPStatusCodes {
			if response.StatusCode == successCode {
				success = true
				break
			}
		}

		if !success {
			close(transformerChanel)
			return errors.Wrapf(
				&HTTPError{
					StatusCode: response.StatusCode,
					Body:       response.Body,
				},
				"expected response code to be one of: %+v, got %d %s",
				h.successHTTPStatusCodes,
				response.StatusCode,
				http.StatusText(response.StatusCode),
			)
		}

		data, err = io.ReadAll(response.Body)
		if err != nil {
			close(transformerChanel)
			return errors.Wrap(err, "TransformerHandler: cannot read response body")
		}
		transformerChanel <- data
	}

	return errs.Wait()
}

// SetFlushInterval specifies the flush interval
// to flush to the writer while copying the
// transformed data.
// If zero, no periodic flushing is done.
// A negative value means to flush immediately
// after each write to the client.
func (h *TransformerHandler) SetFlushInterval(interval time.Duration) {
	h.flushInterval = interval
}
