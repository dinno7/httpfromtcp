package server

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/dinno7/httpfromtcp/internal/request"
	"github.com/dinno7/httpfromtcp/internal/response"
)

type Server struct {
	listener  net.Listener
	isClosed  *atomic.Bool
	handlerFn Handler
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		listener:  listener,
		isClosed:  new(atomic.Bool),
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
			if s.isClosed.Load() {
				return
			}
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		fmt.Printf("error in parsing request: %s", err)
		hErr := &HandlerError{
			StatusCode: response.StatusCodeBadRequest,
			Message:    []byte(err.Error()),
		}
		_ = errors.Join(hErr.Write(conn), conn.Close())
		return
	}

	responseBodyBuf := new(bytes.Buffer)
	if err := s.handlerFn(responseBodyBuf, req); err != nil {
		_ = errors.Join(err.Write(conn), conn.Close())
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
