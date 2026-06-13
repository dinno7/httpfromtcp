package response

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/dinno7/httpfromtcp/internal/headers"
)

type StatusCode uint

const (
	StatusCodeOk                  = 200
	StatusCodeBadRequest          = 400
	StatusCodeNotFound            = 404
	StatusCodeInternalServerError = 500
)

var statusReasons = map[StatusCode]string{
	StatusCodeOk:                  "Ok",
	StatusCodeBadRequest:          "Bad Request",
	StatusCodeInternalServerError: "Internal Server Error",
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, statusReasons[statusCode])
	_, err := w.Write([]byte(statusLine))
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	h.Set("Date", time.Now().Format("Mon, 02 Jan 2006 15:04:05 GTM"))
	h.Set("Cache-Control", "no-cache")
	return h
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	headersStr := new(strings.Builder)
	headers.ForEach(func(key string, value string) {
		fmt.Fprintf(headersStr, "%s: %s\r\n", key, value)
	})
	_, err := w.Write([]byte(headersStr.String()))
	_, err2 := w.Write([]byte("\r\n"))
	return errors.Join(err, err2)
}
