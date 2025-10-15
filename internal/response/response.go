package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/2bitburrito/http-implementation/internal/headers"
)

type (
	Writer     struct{}
	StatusCode int
)

const (
	OK                  StatusCode = 200
	BadRequest          StatusCode = 400
	InternalServerError StatusCode = 500
)

// func (w *Writer) WriteStatusLine(statusCode StatusCode) error
//
// func (w *Writer) WriteHeaders(headers headers.Headers) error
//
// func (w *Writer) WriteBody(p []byte) (int, error)
func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
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
	_, err := w.Write([]byte(d))
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
		"content-type":   "text/plain",
	}
	return headers
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for k, v := range headers {
		_, err := fmt.Fprintf(w, "%s: %s\r\n", k, v)
		if err != nil {
			return err
		}
	}
	w.Write([]byte("\r\n"))
	fmt.Println("Finished writing headers")
	return nil
}
