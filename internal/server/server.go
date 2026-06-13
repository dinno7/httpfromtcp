package server

import (
	"bytes"
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

	isClosed := new(atomic.Bool)
	isClosed.Store(false)

	server := &Server{
		listener:  listener,
		isClosed:  isClosed,
		handlerFn: handler,
	}
	go server.listen()
	return server, nil
}

func (s *Server) Close() error {
	s.isClosed.Store(true)
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

	responseBodyBuf := new(bytes.Buffer)
	responseWriter := response.NewResponse(responseBodyBuf)

	req, err := request.RequestFromReader(conn)
	if err != nil {
		fmt.Printf("error in parsing request: %s", err)
		hErr := &HandlerError{
			StatusCode: response.StatusCodeBadRequest,
			Message:    []byte(err.Error()),
		}
		_ = hErr.Write(responseWriter)
		if _, err := conn.Write(responseBodyBuf.Bytes()); err != nil {
			fmt.Println("Something went wrong", err)
		}
		return
	}

	if err := s.handlerFn(responseWriter, req); err != nil {
		_ = err.Write(responseWriter)
		if _, err := conn.Write(responseBodyBuf.Bytes()); err != nil {
			fmt.Println("Something went wrong", err)
		}
		return
	}

	if _, err := conn.Write(responseBodyBuf.Bytes()); err != nil {
		fmt.Println("Something went wrong", err)
	}
}
