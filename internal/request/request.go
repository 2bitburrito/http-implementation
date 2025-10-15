// Package request processes an HTTP Request
package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/2bitburrito/http-implementation/internal/headers"
)

type Request struct {
	RequestLine       RequestLine
	Headers           headers.Headers
	State             RequestState
	Body              []byte
	reportedConentLen int
	totalBodyParsed   int
}

type RequestState int

const (
	requestStateInitialised RequestState = iota
	requestParsingHeaders
	requestParsingBody
	requestStateDone
)

type RequestLine struct {
	Method        string
	HTTPVersion   string
	RequestTarget string
}

var allowedMethods = []string{"GET", "POST", "PUT", "DETELE", "PATCH"}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buffer := make([]byte, 8)
	currReadIdx := 0

	req := &Request{
		State:   requestStateInitialised,
		Headers: headers.NewHeaders(),
	}

	for req.State != requestStateDone {
		if currReadIdx >= len(buffer) {
			newSlice := make([]byte, len(buffer)*2)
			copy(newSlice, buffer)
			buffer = newSlice
		}
		numberBytesRead, err := reader.Read(buffer[currReadIdx:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				req.State = requestStateDone
				break
			}
			return nil, err
		}
		if numberBytesRead == 0 {
			return nil, fmt.Errorf("only reading 0 bytes")
		}
		currReadIdx += numberBytesRead

		numberBytesParsed, err := req.parse(buffer[:currReadIdx])
		if err != nil {
			return nil, fmt.Errorf("error while parsing request: %w", err)
		}

		// numberBytesParsed is coming back as too long in large body:
		copy(buffer, buffer[numberBytesParsed:])
		currReadIdx -= numberBytesParsed
	}
	if req.RequestLine.HTTPVersion == "" ||
		req.RequestLine.Method == "" ||
		req.RequestLine.RequestTarget == "" {
		return nil, fmt.Errorf("unknown error while parsing request line")
	}
	if len(req.Body) < req.reportedConentLen {
		return req, fmt.Errorf("body is shorter than content-length")
	}
	return req, nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.State {
	case requestStateInitialised:
		req, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		r.RequestLine = *req
		r.State = requestParsingHeaders
		return n, nil
	case requestParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return n, err
		}
		if n == 0 {
			return 0, nil
		}
		if done {
			r.State = requestParsingBody
		}
		return n, nil
	case requestParsingBody:
		contentLen, ok := r.Headers.Get("content-length")
		if !ok {
			r.State = requestStateDone
			r.reportedConentLen = 0
			return len(r.Body), nil
		}
		lenInt, err := strconv.Atoi(string(contentLen))
		if err != nil {
			return 0, fmt.Errorf("malformed content length: %s", contentLen)
		}
		r.reportedConentLen = lenInt
		r.Body = append(r.Body, data...)
		if len(r.Body) > lenInt {
			return len(r.Body), fmt.Errorf("actual body length is longer than reported")
		}
		if len(r.Body) == lenInt {
			r.State = requestStateDone
			return len(data), nil
		}
		r.totalBodyParsed += len(data)
		return len(data), nil
	default:
		return 0, fmt.Errorf("error: trying to read data in an invalid state")
	}
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	requestLineEnd := bytes.Index(data, []byte("\r\n"))
	if requestLineEnd == -1 {
		return nil, 0, nil
	}
	msg := string(data[:requestLineEnd])

	splitMsg := strings.Split(msg, " ")
	if len(splitMsg) != 3 {
		return nil, 0, fmt.Errorf("incorrect number of parts in req line: %+v", splitMsg)
	}
	httpVersionSplit := strings.Split(splitMsg[2], "/")
	if len(httpVersionSplit) != 2 {
		return nil, 0, fmt.Errorf("incorrect formatting of http version: only version 1.1 allowed")
	}
	if httpVersionSplit[0] != "HTTP" {
		return nil, 0, fmt.Errorf("invalid version: only supporting http. Got: %q", httpVersionSplit[0])
	}
	httpVersion := httpVersionSplit[1]
	if httpVersion != "1.1" {
		return nil, 0, fmt.Errorf("implementation only covers http version 1.1, got: %v", httpVersion)
	}

	method := splitMsg[0]
	if !slices.Contains(allowedMethods, method) {
		return nil, 0, fmt.Errorf("unknown method: %s", method)
	}

	target := splitMsg[1]
	// Check for whitespace
	if strings.Contains(target, " ") ||
		strings.Contains(target, "\n") ||
		strings.Contains(target, "\r") {
		return nil, 0, fmt.Errorf("bad target path: path is malformed: contains whitespaces")
	}

	reqLines := &RequestLine{
		Method:        method,
		RequestTarget: target,
		HTTPVersion:   httpVersion,
	}
	return reqLines, requestLineEnd + 2, nil
}
