package response

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"surya.httpfromtcp/internal/headers"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

const (
	writerStateInitialized = iota
	writerStateStatusWritten
	writerStateHeadersWritten
	writerStateBodyWritten
)

type Writer struct {
	w     io.Writer
	state int
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w:     w,
		state: writerStateInitialized,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != writerStateInitialized {
		return errors.New("status line already written")
	}

	var reasonPhrase string
	switch statusCode {
	case StatusOK:
		reasonPhrase = "OK"
	case StatusBadRequest:
		reasonPhrase = "Bad Request"
	case StatusInternalServerError:
		reasonPhrase = "Internal Server Error"
	default:
		reasonPhrase = ""
	}

	var statusLine string
	if reasonPhrase != "" {
		statusLine = fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reasonPhrase)
	} else {
		statusLine = fmt.Sprintf("HTTP/1.1 %d \r\n", statusCode)
	}

	_, err := w.w.Write([]byte(statusLine))
	if err != nil {
		return err
	}

	w.state = writerStateStatusWritten
	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.state != writerStateStatusWritten {
		return errors.New("headers must be written after status line and before body")
	}

	for key, value := range headers {
		line := fmt.Sprintf("%s: %s\r\n", key, value)
		_, err := w.w.Write([]byte(line))
		if err != nil {
			return err
		}
	}
	_, err := w.w.Write([]byte("\r\n"))
	if err != nil {
		return err
	}

	w.state = writerStateHeadersWritten
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != writerStateHeadersWritten {
		return 0, errors.New("body must be written after headers")
	}

	n, err := w.w.Write(p)
	w.state = writerStateBodyWritten
	return n, err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := make(headers.Headers)
	h["content-length"] = strconv.Itoa(contentLen)
	h["connection"] = "close"
	h["content-type"] = "text/plain"
	return h
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var reasonPhrase string
	switch statusCode {
	case StatusOK:
		reasonPhrase = "OK"
	case StatusBadRequest:
		reasonPhrase = "Bad Request"
	case StatusInternalServerError:
		reasonPhrase = "Internal Server Error"
	default:
		reasonPhrase = ""
	}

	var statusLine string
	if reasonPhrase != "" {
		statusLine = fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reasonPhrase)
	} else {
		statusLine = fmt.Sprintf("HTTP/1.1 %d \r\n", statusCode)
	}

	_, err := w.Write([]byte(statusLine))
	return err
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, value := range headers {
		line := fmt.Sprintf("%s: %s\r\n", key, value)
		_, err := w.Write([]byte(line))
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("\r\n"))
	return err
}
