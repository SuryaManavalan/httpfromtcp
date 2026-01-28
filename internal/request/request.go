package request

import (
	"errors"
	"io"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	requestText := string(data)
	requestLine, err := parseRequestLine(requestText)
	if err != nil {
		return nil, err
	}

	return &Request{
		RequestLine: *requestLine,
	}, nil
}

func parseRequestLine(requestText string) (*RequestLine, error) {
	// Split by \r\n to get the first line
	lines := strings.Split(requestText, "\r\n")
	if len(lines) == 0 {
		return nil, errors.New("empty request")
	}

	firstLine := lines[0]

	// Split the request line into parts
	parts := strings.Split(firstLine, " ")
	if len(parts) != 3 {
		return nil, errors.New("invalid request line: expected 3 parts")
	}

	method := parts[0]
	requestTarget := parts[1]
	httpVersionFull := parts[2]

	// Verify method contains only capital alphabetic characters
	for _, ch := range method {
		if !unicode.IsUpper(ch) {
			return nil, errors.New("invalid method: must contain only capital letters")
		}
	}

	// Parse and verify HTTP version
	if !strings.HasPrefix(httpVersionFull, "HTTP/") {
		return nil, errors.New("invalid HTTP version format")
	}

	version := strings.TrimPrefix(httpVersionFull, "HTTP/")
	if version != "1.1" {
		return nil, errors.New("unsupported HTTP version: only HTTP/1.1 is supported")
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   version,
	}, nil
}
