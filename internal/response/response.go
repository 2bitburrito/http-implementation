package response

import (
	"fmt"
	"net"
	"strconv"

	"github.com/2bitburrito/http-implementation/internal/headers"
)

type (
	StatusCode int
)

type Writer struct {
	Conn net.Conn
}

const (
	OK                  StatusCode = 200
	BadRequest          StatusCode = 400
	InternalServerError StatusCode = 500
)

func (w *Writer) WriteBody(p []byte) (int, error) {
	return w.Conn.Write(p)
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	d := ""
	switch statusCode {
	case 200:
		d = "HTTP/1.1 200 OK\r\n"
	case 400:
		d = "HTTP/1.1 400 Bad Request\r\n"
	case 500:
		d = "HTTP/1.1 500 Internal Server Error\r\n"
	default:
		return fmt.Errorf("unsupported status code: %q", statusCode)
	}
	_, err := w.Conn.Write([]byte(d))
	if err != nil {
		return err
	}
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	contentLenStr := strconv.Itoa(contentLen)
	headers := map[string]string{
		"content-length": contentLenStr,
		"connection":     "close",
	}
	return headers
}

func (w *Writer) WriteHeaders(h headers.Headers) error {
	for k, v := range h {
		_, err := fmt.Fprintf(w.Conn, "%s: %s\r\n", k, v)
		if err != nil {
			return err
		}
	}
	w.Conn.Write([]byte("\r\n"))
	return nil
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	t := 0
	n, err := fmt.Fprintf(w.Conn, "%x\r\n", len(p))
	if err != nil {
		return t, err
	}
	t += n
	n, err = w.Conn.Write(p)
	if err != nil {
		return t, err
	}
	t += n
	n, err = w.Conn.Write([]byte("\r\n"))
	if err != nil {
		return t, err
	}
	t += n
	return t, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	t, err := w.Conn.Write([]byte("0\r\n"))
	if err != nil {
		return 0, err
	}
	return t, nil
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	for k, v := range h {
		_, err := fmt.Fprintf(w.Conn, "%s: %s\r\n", k, v)
		if err != nil {
			return err
		}
	}
	w.Conn.Write([]byte("\r\n"))
	return nil
}
