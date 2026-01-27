package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)
		defer f.Close()

		str := ""
		for {
			data := make([]byte, 8)
			n, err := f.Read(data)
			if err != nil {
				break
			}

			data = data[:n]
			if i := bytes.IndexByte(data, '\n'); i >= 0 {
				str += string(data[:i])
				data = data[i+1:]
				ch <- str
				str = ""
			}

			str += string(data)
		}

		if len(str) != 0 {
			ch <- str
		}
	}()

	return ch
}

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

		ch := getLinesChannel(conn)
		for line := range ch {
			fmt.Println(line)
		}

		fmt.Println("connection closed")
	}
}
