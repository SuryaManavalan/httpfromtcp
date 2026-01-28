package main

import (
	"fmt"
	"log"
	"net"

	"surya.httpfromtcp/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatalf("failed to listen: %s", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to accept connection: %s", err)
			continue
		}

		fmt.Println("connection accepted")

		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Printf("failed to parse request: %s", err)
			conn.Close()
			continue
		}

		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", req.RequestLine.Method)
		fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)

		conn.Close()
		fmt.Println("connection closed")
	}
}
