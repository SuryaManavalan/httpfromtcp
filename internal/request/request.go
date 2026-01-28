package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"surya.httpfromtcp/internal/headers"
)

const (
	requestStateInitialized = iota
	requestStateParsingHeaders
	requestStateParsingBody
	requestStateDone
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	state       int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := &Request{
		state:   requestStateInitialized,
		Headers: headers.NewHeaders(),
	}

	buffer := make([]byte, 8)
	var accumulated []byte

	for req.state != requestStateDone {
		n, err := reader.Read(buffer)
		if n > 0 {
			accumulated = append(accumulated, buffer[:n]...)

			consumed, parseErr := req.parse(accumulated)
			if parseErr != nil {
				return nil, parseErr
			}

			accumulated = accumulated[consumed:]
		}

		if err == io.EOF {
			if req.state != requestStateDone {
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
	if r.state == requestStateDone {
		return 0, nil
	}

	totalBytesParsed := 0
	for r.state != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return totalBytesParsed, err
		}

		if n == 0 {
			break
		}

		totalBytesParsed += n
	}

	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case requestStateInitialized:
		consumed, requestLine, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}

		if consumed == 0 {
			return 0, nil
		}

		r.RequestLine = *requestLine
		r.state = requestStateParsingHeaders

		return consumed, nil

	case requestStateParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}

		if done {
			r.state = requestStateParsingBody
		}

		return n, nil

	case requestStateParsingBody:
		contentLengthStr := r.Headers.Get("content-length")
		if contentLengthStr == "" {
			r.state = requestStateDone
			return 0, nil
		}

		contentLength, err := strconv.Atoi(contentLengthStr)
		if err != nil {
			return 0, fmt.Errorf("invalid content-length: %w", err)
		}

		r.Body = append(r.Body, data...)

		if len(r.Body) > contentLength {
			return 0, fmt.Errorf("body length exceeds content-length: got %d, expected %d", len(r.Body), contentLength)
		}

		if len(r.Body) == contentLength {
			r.state = requestStateDone
		}

		return len(data), nil

	default:
		return 0, nil
	}
}

func parseRequestLine(data []byte) (int, *RequestLine, error) {
	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		return 0, nil, nil
	}

	firstLine := string(data[:idx])

	parts := strings.Split(firstLine, " ")
	if len(parts) != 3 {
		return 0, nil, errors.New("invalid request line: expected 3 parts")
	}

	method := parts[0]
	requestTarget := parts[1]
	httpVersionFull := parts[2]

	for _, ch := range method {
		if !unicode.IsUpper(ch) {
			return 0, nil, errors.New("invalid method: must contain only capital letters")
		}
	}

	if !strings.HasPrefix(httpVersionFull, "HTTP/") {
		return 0, nil, errors.New("invalid HTTP version format")
	}

	version := strings.TrimPrefix(httpVersionFull, "HTTP/")
	if version != "1.1" {
		return 0, nil, errors.New("unsupported HTTP version: only HTTP/1.1 is supported")
	}

	consumed := idx + 2

	return consumed, &RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   version,
	}, nil
}
