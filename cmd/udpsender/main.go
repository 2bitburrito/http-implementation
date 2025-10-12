package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	address, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		fmt.Println("error resolving udp address: ", err)
		os.Exit(1)
	}
	conn, err := net.DialUDP("udp", nil, address)
	if err != nil {
		fmt.Println("error creating udp connection: ", err)
		os.Exit(1)
	}
	defer conn.Close()
	fmt.Println("Server is ready to post UDP on port 42069")

	readerIn := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf(">")
		input, err := readerIn.ReadString('\n')
		if err != nil {
			fmt.Println("error reading from stdin:", err)
			os.Exit(1)
		}
		_, err = conn.Write([]byte(input))
		if err != nil {
			fmt.Println("error writing to udp connection:", err)
			os.Exit(1)
		}
	}
}
