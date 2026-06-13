package server

import (
	"fmt"

	"github.com/dinno7/httpfromtcp/internal/request"
	"github.com/dinno7/httpfromtcp/internal/response"
)

type Handler func(w *response.Response, req *request.Request) *HandlerError

type HandlerError struct {
	StatusCode response.StatusCode
	Message    []byte
}

func (he *HandlerError) Write(r *response.Response) error {
	if err := r.WriteStatusLine(he.StatusCode); err != nil {
		return err
	}
	_, err := r.Write(he.Message)
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
