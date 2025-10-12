package main

import (
	"fmt"
	"log"
	"net"

	"github.com/2bitburrito/http-implementation/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal("error while instansiating tcp listener: ", err)
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("error while accepting connection: ", err)
		}
		fmt.Printf("Connection Accepted From: %s\n", conn.RemoteAddr())

		request, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal("error reading from connection")
		}
		fmt.Printf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n",
			request.RequestLine.Method,
			request.RequestLine.RequestTarget,
			request.RequestLine.HTTPVersion)
		fmt.Printf("Connection Closed From: %s\n", conn.RemoteAddr())
	}
}
