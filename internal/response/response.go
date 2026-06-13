package response

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/dinno7/httpfromtcp/internal/headers"
)

const CRLF = "\r\n"

var ErrInvalidState = errors.New("invalid state of response")

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
	StatusCodeNotFound:            "Not Found",
	StatusCodeInternalServerError: "Internal Server Error",
}

type responseState string

const (
	responseStateStatusLine responseState = "status_line"
	responseStateBody       responseState = "body"
	responseStateDone       responseState = "done"
)

type Response struct {
	writer  io.Writer
	headers headers.Headers
	state   responseState
}

func NewResponse(writer io.Writer) *Response {
	return &Response{
		writer:  writer,
		headers: headers.NewHeaders(),
		state:   responseStateStatusLine,
	}
}

func (r *Response) Headers() headers.Headers {
	return r.headers
}

func (r *Response) WriteStatusLine(statusCode StatusCode) error {
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s%s", statusCode, statusReasons[statusCode], CRLF)
	if _, err := r.writer.Write([]byte(statusLine)); err != nil {
		return err
	}
	r.state = responseStateBody
	return nil
}

func (r *Response) Write(p []byte) (n int, err error) {
	if r.state == responseStateStatusLine {
		if err := r.WriteStatusLine(StatusCodeOk); err != nil {
			return 0, err
		}
	}

	r.setDefaultHeaders(len(p))

	responseData := new(strings.Builder)

	// NOTE: Writing headers content
	r.headers.ForEach(func(key string, value string) {
		fmt.Fprintf(responseData, "%s: %s%s", key, value, CRLF)
	})

	// NOTE: Write headers separator
	responseData.WriteString(CRLF)

	// NOTE: Writing body content
	responseData.Write(p)

	r.state = responseStateDone
	return r.writer.Write([]byte(responseData.String()))
}

func (r *Response) setDefaultHeaders(contentLength int) {
	defaultHeaders := map[string]string{
		"Content-Length": fmt.Sprintf("%d", contentLength),
		"Connection":     "close",
		"Date":           time.Now().Format("Mon, 02 Jan 2006 15:04:05 GTM"),
		"Cache-Control":  "no-cache",
	}

	for key, value := range defaultHeaders {
		if r.headers.Get(key) == "" {
			r.headers.Set(key, value)
		}
	}
}

func GetDefaultHeaders() headers.Headers {
	h := headers.NewHeaders()
	h.Set("Connection", "close")
	h.Set("Date", time.Now().Format("Mon, 02 Jan 2006 15:04:05 GTM"))
	h.Set("Cache-Control", "no-cache")
	return h
}

func GetContentHeaders(contentLen int, contentType string) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Content-Type", "text/plain")
	return h
}

func (r *Response) checkState(expectedState responseState) error {
	if r.state != expectedState {
		return fmt.Errorf(
			"%w: expect response be %s state",
			ErrInvalidState,
			expectedState,
		)
	}
	return nil
}
