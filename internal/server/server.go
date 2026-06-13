package server

import (
	"errors"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/dinno7/httpfromtcp/internal/request"
	"github.com/dinno7/httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
	isOpen   *atomic.Bool
}

func Serve(port int) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	isOpen := &atomic.Bool{}
	isOpen.Store(true)
	server := &Server{
		listener: listener,
		isOpen:   isOpen,
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
	_, err := request.RequestFromReader(conn)
	if err != nil {
		fmt.Printf("error in parsing request: %s", err)
	}

	statusLineWriteErr := response.WriteStatusLine(conn, response.StatusCodeOk)
	headers := response.GetDefaultHeaders(0)
	headerWriteErr := response.WriteHeaders(conn, headers)
	closeErr := conn.Close()

	if err := errors.Join(statusLineWriteErr, headerWriteErr, closeErr); err != nil {
		fmt.Println("Something went wrong", err)
	}
}
