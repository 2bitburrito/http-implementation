// Package request processes an HTTP Request
package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
)

type Request struct {
	RequestLine RequestLine
	State       RequestState
	Body        *[]byte
}

type RequestState int

const (
	requestStateInitialised RequestState = iota
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
		State: requestStateInitialised,
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

		copy(buffer, buffer[:currReadIdx])
		currReadIdx -= numberBytesParsed
	}
	if req.RequestLine.HTTPVersion == "" ||
		req.RequestLine.Method == "" ||
		req.RequestLine.RequestTarget == "" {
		return nil, fmt.Errorf("unknown error while parsing request line")
	}
	return req, nil
}

func (r *Request) parse(data []byte) (int, error) {
	if r.State != requestStateInitialised {
		return 0, fmt.Errorf("error: trying to read data in an invalid state")
	}
	req, n, err := parseRequestLine(data)
	if err != nil {
		return 0, err
	}
	if n == 0 {
		return 0, nil
	}
	r.RequestLine = *req
	r.State = requestStateDone
	return len(data), nil
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
	// Check for path correctness
	if !strings.HasPrefix(target, "/") && target != "*" {
		return nil, 0, fmt.Errorf("bad target path: should start with \"/\" received %q", target)
	}

	reqLines := &RequestLine{
		Method:        method,
		RequestTarget: target,
		HTTPVersion:   httpVersion,
	}
	return reqLines, requestLineEnd + 2, nil
}
