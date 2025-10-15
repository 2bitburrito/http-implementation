// Package server is the main entrypoint for the full server implementation
package server

import (
	"bytes"
	"fmt"
	"io"
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
type HandlerError struct {
	StatusCode int
	Err        error
}
type Handler func(w io.Writer, req *request.Request) *HandlerError

func (he *HandlerError) Error() string {
	return fmt.Sprintf("%d error: %s", he.StatusCode, he.Err.Error())
}

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
	buf := bytes.NewBuffer(*new([]byte))
	statusCode := 200
	if err := s.Handler(buf, req); err != nil {
		fmt.Println("there was an error in the server handler: ", err.Error())
		statusCode = err.StatusCode
	}

	// Writing status line
	if err := response.WriteStatusLine(conn, response.StatusCode(statusCode)); err != nil {
		fmt.Println("error while writing status line: ", err)
		return
	}

	// Writing headers
	headers := response.GetDefaultHeaders(buf.Len())
	if err := response.WriteHeaders(conn, headers); err != nil {
		fmt.Println("error while writing headers: ", err)
		return
	}

	// Writing Body:
	conn.Write(buf.Bytes())
}

func (s *Server) Close() {
	s.isOpen.Store(false)
	s.listener.Close()
}
