package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
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
	f, err := os.Open("messages.txt")
	if err != nil {
		log.Fatalf("failed to open file: %s", err)
	}

	ch := getLinesChannel(f)
	for line := range ch {
		fmt.Printf("read: %s \n", line)
	}
}
