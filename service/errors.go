package service

import (
	"fmt"
	"io"
	"net/http"
)

type HTTPError struct {
	StatusCode int
	Body       io.Reader
}

func (e HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d / %s", e.StatusCode, http.StatusText(e.StatusCode))
}
