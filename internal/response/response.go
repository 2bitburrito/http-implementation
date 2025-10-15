package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/2bitburrito/http-implementation/internal/headers"
)

type StatusCode int

const (
	OK                  StatusCode = 200
	BadRequest          StatusCode = 400
	InternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	d := ""
	switch statusCode {
	case 200:
		d = "HTTP/1.1 200 OK\r\n"
	case 400:
		d = "HTTP/1.1 200 Bad Request\r\n"
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
	buf := make([]byte, 0)
	for k, v := range headers {
		buf = fmt.Appendf(buf, "%s: %s\r\n", k, v)
	}
	buf = fmt.Appendf(buf, "\r\n")
	_, err := w.Write(buf)
	if err != nil {
		return err
	}
	return nil
}
