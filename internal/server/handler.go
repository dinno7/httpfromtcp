package server

import (
	"fmt"
	"io"
	"net"

	"github.com/dinno7/httpfromtcp/internal/request"
	"github.com/dinno7/httpfromtcp/internal/response"
)

type Handler func(w io.Writer, req *request.Request) *HandlerError

type HandlerError struct {
	StatusCode response.StatusCode
	Message    []byte
}

func (he *HandlerError) Write(conn net.Conn) error {
	if err := response.WriteStatusLine(conn, he.StatusCode); err != nil {
		return err
	}
	headers := response.GetDefaultHeaders(len(he.Message))
	if err := response.WriteHeaders(conn, headers); err != nil {
		return err
	}
	_, err := conn.Write(he.Message)
	return err
}

func NewHandlerError(code response.StatusCode, message []byte) *HandlerError {
	return &HandlerError{
		StatusCode: code,
		Message:    message,
	}
}

func NewHandlerErrorBadRequest(message []byte) *HandlerError {
	return NewHandlerError(response.StatusCodeBadRequest, message)
}

func NewHandlerErrorNotFound(message []byte) *HandlerError {
	return NewHandlerError(response.StatusCodeNotFound, message)
}

func NewHandlerErrorInternalServerError(message []byte) *HandlerError {
	return NewHandlerError(response.StatusCodeInternalServerError, message)
}

func (h *HandlerError) Error() string {
	return fmt.Sprintf("%d - %s", h.StatusCode, h.Message)
}
