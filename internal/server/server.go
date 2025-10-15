// Package server is the main entrypoint for the full server implementation
package server

import (
	"fmt"
	"net"
	"sync/atomic"

	"github.com/2bitburrito/http-implementation/internal/response"
)

type Server struct {
	listener net.Listener
	isOpen   *atomic.Bool
}

// Serve implements [net.Listener]
func Serve(port int) (*Server, error) {
	isOpen := atomic.Bool{}
	isOpen.Store(true)

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		listener: listener,
		isOpen:   &isOpen,
	}
	go server.listen()
	return server, nil
}

func (s *Server) listen() {
	for s.isOpen.Load() {
		conn, err := s.listener.Accept()
		if err != nil {
			if !s.isOpen.Load() {
				return
			}
			fmt.Printf("error accepting connection: %s", err)
			return
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	headers := response.GetDefaultHeaders(0)
	if err := response.WriteStatusLine(conn, 200); err != nil {
		fmt.Println("error while writing status line: ", err)
		return
	}
	if err := response.WriteHeaders(conn, headers); err != nil {
		fmt.Println("error while writing headers: ", err)
		return
	}
	conn.Close()
}

func (s *Server) Close() {
	s.isOpen.Store(false)
	s.listener.Close()
}
