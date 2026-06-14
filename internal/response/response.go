package response

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/dinno7/httpfromtcp/internal/headers"
)

const CRLF = "\r\n"

var ErrInvalidState = errors.New("invalid state of response")

type StatusCode uint16

const (
	StatusCodeOk                  StatusCode = 200
	StatusCodeBadRequest          StatusCode = 400
	StatusCodeNotFound            StatusCode = 404
	StatusCodeInternalServerError StatusCode = 500
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
	responseStateHeaders    responseState = "headers"
	responseStateBody       responseState = "body"
	responseStateDone       responseState = "done"
)

type Response struct {
	writer     io.Writer
	statusCode StatusCode
	headers    headers.Headers
	state      responseState
}

func NewResponse(writer io.Writer) *Response {
	return &Response{
		writer:     writer,
		headers:    headers.NewHeaders(),
		state:      responseStateStatusLine,
		statusCode: StatusCodeOk,
	}
}

func (r *Response) Headers() headers.Headers {
	return r.headers
}

func (r *Response) SetStatusCode(statusCode StatusCode) {
	r.statusCode = statusCode
}

func (r *Response) Write(p []byte) (n int, err error) {
	// NOTE: Writing status line
	if n, err := r.writeStatusLine(); err != nil {
		return n, err
	}

	// NOTE: Writing headers content
	if r.headers.Get("Content-Length") == "" {
		r.headers.Set("Content-Length", fmt.Sprintf("%d", len(p)))
	}
	if n, err := r.writeHeaders(); err != nil {
		return n, err
	}

	// NOTE: Writing body content
	r.state = responseStateDone
	return r.writer.Write(p)
}

func (r *Response) WriteChunkedBody(p []byte) (int, error) {
	// NOTE: Write statusline and go to header state
	if r.state == responseStateStatusLine {
		if n, err := r.writeStatusLine(); err != nil {
			return n, err
		}
	}

	// NOTE: Write statusline and go to body state
	if r.state == responseStateHeaders {
		r.headers.Delete("Content-Length")
		r.Headers().Set("Transfer-Encoding", "chunked")
		if n, err := r.writeHeaders(); err != nil {
			return n, err
		}
	}

	if r.state != responseStateBody {
		return 0, ErrInvalidState
	}

	chunk := new(strings.Builder)
	sizeStr := strconv.FormatInt(int64(len(p)), 16)
	chunk.Write([]byte(sizeStr + CRLF))
	chunk.Write(p)
	chunk.Write([]byte(CRLF))

	return r.writer.Write([]byte(chunk.String()))
}

func (r *Response) WriteChunkedBodyDone() (int, error) {
	// NOTE: Writing body content
	sizeStr := strconv.FormatInt(int64(0), 16)
	r.state = responseStateDone
	return r.writer.Write([]byte(sizeStr + CRLF + CRLF))
}

func (r *Response) WriteTrailers(h headers.Headers) (int, error) {
	headerBytes, err := h.RawBytes()
	if err != nil {
		return 0, err
	}
	return r.writer.Write(headerBytes)
}

func (r *Response) writeStatusLine() (int, error) {
	if r.statusCode == 0 {
		r.statusCode = StatusCodeOk
	}
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s%s", r.statusCode, statusReasons[r.statusCode], CRLF)
	r.state = responseStateHeaders
	return r.writer.Write([]byte(statusLine))
}

func (r *Response) writeHeaders() (int, error) {
	r.setDefaultHeaders()

	headerBytes, err := r.headers.RawBytes()
	if err != nil {
		return 0, err
	}

	r.state = responseStateBody
	return r.writer.Write(headerBytes)
}

func (r *Response) setDefaultHeaders() {
	defaultHeaders := map[string]string{
		"Connection":    "close",
		"Date":          time.Now().UTC().Format(time.RFC1123),
		"Cache-Control": "no-cache",
	}

	for key, value := range defaultHeaders {
		if r.headers.Get(key) == "" {
			r.headers.Set(key, value)
		}
	}
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
