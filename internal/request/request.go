package request

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"unicode"
)

const (
	stateInitialized = iota
	stateDone
)

type Request struct {
	RequestLine RequestLine
	state       int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := &Request{
		state: stateInitialized,
	}

	buffer := make([]byte, 8)
	var accumulated []byte

	for req.state != stateDone {
		n, err := reader.Read(buffer)
		if n > 0 {
			accumulated = append(accumulated, buffer[:n]...)

			consumed, parseErr := req.parse(accumulated)
			if parseErr != nil {
				return nil, parseErr
			}

			// Remove consumed bytes from accumulated buffer
			accumulated = accumulated[consumed:]
		}

		if err == io.EOF {
			// If we still have data but haven't finished parsing, that's an error
			if req.state != stateDone {
				return nil, errors.New("incomplete request")
			}
			break
		}

		if err != nil {
			return nil, err
		}
	}

	return req, nil
}

func (r *Request) parse(data []byte) (int, error) {
	if r.state == stateDone {
		return 0, nil
	}

	// Try to parse the request line
	consumed, requestLine, err := parseRequestLine(data)
	if err != nil {
		return 0, err
	}

	// If we couldn't parse yet (no \r\n found), return 0 with no error
	if consumed == 0 {
		return 0, nil
	}

	// Successfully parsed the request line
	r.RequestLine = *requestLine
	r.state = stateDone

	return consumed, nil
}

func parseRequestLine(data []byte) (int, *RequestLine, error) {
	// Look for \r\n to find the end of the request line
	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		// Haven't received the full line yet
		return 0, nil, nil
	}

	// Extract the first line
	firstLine := string(data[:idx])

	// Split the request line into parts
	parts := strings.Split(firstLine, " ")
	if len(parts) != 3 {
		return 0, nil, errors.New("invalid request line: expected 3 parts")
	}

	method := parts[0]
	requestTarget := parts[1]
	httpVersionFull := parts[2]

	// Verify method contains only capital alphabetic characters
	for _, ch := range method {
		if !unicode.IsUpper(ch) {
			return 0, nil, errors.New("invalid method: must contain only capital letters")
		}
	}

	// Parse and verify HTTP version
	if !strings.HasPrefix(httpVersionFull, "HTTP/") {
		return 0, nil, errors.New("invalid HTTP version format")
	}

	version := strings.TrimPrefix(httpVersionFull, "HTTP/")
	if version != "1.1" {
		return 0, nil, errors.New("unsupported HTTP version: only HTTP/1.1 is supported")
	}

	// Return bytes consumed (including \r\n)
	consumed := idx + 2

	return consumed, &RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   version,
	}, nil
}
