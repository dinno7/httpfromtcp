package server

import (
	"fmt"
	"net"
	"sync/atomic"

	"github.com/dinno7/httpfromtcp/internal/request"
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

	if _, err := conn.Write(
		[]byte(
			"HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nHello World!",
		),
	); err != nil {
		fmt.Printf("error sending response: %s", err)
	}
	conn.Close()
}
