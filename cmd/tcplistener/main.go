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

		r, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal("error reading from connection: ", err)
		}
		fmt.Printf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n",
			r.RequestLine.Method,
			r.RequestLine.RequestTarget,
			r.RequestLine.HTTPVersion)
		fmt.Printf("Headers:\n")
		for key, value := range r.Headers {
			fmt.Printf("- %s: %s\n", key, value)
		}
		fmt.Printf("Body:\n%s\n", string(r.Body))
		fmt.Printf("Connection Closed From: %s\n", conn.RemoteAddr())
	}
}
