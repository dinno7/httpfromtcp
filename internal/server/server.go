package server

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/dinno7/httpfromtcp/internal/request"
	"github.com/dinno7/httpfromtcp/internal/response"
)

type Server struct {
	listener  net.Listener
	isOpen    *atomic.Bool
	handlerFn Handler
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	isOpen := &atomic.Bool{}
	isOpen.Store(true)
	server := &Server{
		listener:  listener,
		isOpen:    isOpen,
		handlerFn: handler,
	}
	go server.listen()
	return server, nil
}

func (s *Server) Close() error {
	return s.listener.Close()
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
		}
		if s.isOpen.Load() {
			go s.handle(conn)
		}
	}
}

func (s *Server) handle(conn net.Conn) {
	req, err := request.RequestFromReader(conn)
	if err != nil {
		fmt.Printf("error in parsing request: %s", err)
	}

	responseBodyBuf := new(bytes.Buffer)
	if err := s.handlerFn(responseBodyBuf, req); err != nil {
		message := []byte(err.Error())
		var statusCode response.StatusCode = response.StatusCodeInternalServerError

		if err, ok := errors.AsType[*HandlerError](err); ok {
			statusCode = err.StatusCode
			message = err.Message
		}

		_ = response.WriteStatusLine(conn, statusCode)
		headers := response.GetDefaultHeaders(len(message))
		_ = response.WriteHeaders(conn, headers)
		conn.Write(message)
		conn.Close()
		return
	}

	contentLength := responseBodyBuf.Len()

	statusLineWriteErr := response.WriteStatusLine(conn, response.StatusCodeOk)
	headers := response.GetDefaultHeaders(contentLength)
	headerWriteErr := response.WriteHeaders(conn, headers)

	if contentLength > 0 {
		conn.Write(responseBodyBuf.Bytes())
	}
	closeErr := conn.Close()

	if err := errors.Join(statusLineWriteErr, headerWriteErr, closeErr); err != nil {
		fmt.Println("Something went wrong", err)
	}
}
