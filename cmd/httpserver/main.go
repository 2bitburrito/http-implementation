package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/2bitburrito/http-implementation/internal/request"
	"github.com/2bitburrito/http-implementation/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w io.Writer, req *request.Request) *server.HandlerError {
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		_, err := w.Write([]byte("Your problem is not my problem\n"))
		if err != nil {
			fmt.Println("error while writing to writer in handler: ", err)
		}
		return &server.HandlerError{
			StatusCode: 400,
			Err:        fmt.Errorf("it's your problem"),
		}
	case "/myproblem":
		_, err := w.Write([]byte("Woopsie, my bad\n"))
		if err != nil {
			fmt.Println("error while writing to writer in handler: ", err)
		}
		return &server.HandlerError{
			StatusCode: 500,
			Err:        fmt.Errorf("my problem"),
		}
	default:
		_, err := w.Write([]byte("All good, frfr\n"))
		if err != nil {
			fmt.Println("error while writing to writer in handler: ", err)
		}
		return nil
	}
}
