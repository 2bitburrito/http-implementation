// Package server is the main entrypoint for the full server implementation
package server

import (
	"fmt"
	"net"
	"sync/atomic"

	"github.com/2bitburrito/http-implementation/internal/request"
	"github.com/2bitburrito/http-implementation/internal/response"
)

type Server struct {
	listener net.Listener
	isOpen   *atomic.Bool
	Handler  Handler
}
type Handler func(w *response.Writer, req *request.Request)

func Serve(port int, hdlr Handler) (*Server, error) {
	isOpen := atomic.Bool{}
	isOpen.Store(true)

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		listener: listener,
		isOpen:   &isOpen,
		Handler:  hdlr,
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
	// Ensure we always close the request with crlf
	defer func() {
		defer conn.Close()
		_, err := fmt.Fprintf(conn, "\r\n")
		if err != nil {
			fmt.Println("error while writing crlf to connection: ", err)
		}
	}()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		fmt.Println("error reading request: ", err)
	}
	writer := &response.Writer{
		Conn: conn,
	}
	s.Handler(writer, req)
}

func (s *Server) Close() {
	s.isOpen.Store(false)
	s.listener.Close()
}
