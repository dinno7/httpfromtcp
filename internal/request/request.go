package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
)

const (
	VALID_HTTP_VERSION = "1.1"
	SEPARATOR          = "\r\n"
)

var HTTP_METHODS = []string{
	"GET", "POST", "HEAD", "OPTION", "DELETE", "PATCH",
}

var (
	ErrUnsupportedHttpVersion = errors.New("http version not support")
	ErrInvalidRequestHeader   = errors.New("invalid request header")
	ErrInvalidHttpMethod      = errors.New("invalid http method")
	ErrInvalidTargetPath      = errors.New("invalid target path")
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine
}

func RequestFromReader(r io.Reader) (*Request, error) {
	rawReq, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("unable to read raw request: %w", err)
	}

	requestLine, _, err := parseRequestLine(rawReq)
	if err != nil {
		return nil, errors.Join(errors.New("faild parse request line"), err)
	}

	return &Request{
		RequestLine: *requestLine,
	}, nil
}

func parseRequestLine(rawRequest []byte) (*RequestLine, int, error) {
	separatorIndex := bytes.Index(rawRequest, []byte(SEPARATOR))
	firstLine := rawRequest[:separatorIndex]

	line := bytes.SplitN(firstLine, []byte{' '}, 3)
	if len(line) != 3 {
		return nil, 0, ErrInvalidRequestHeader
	}

	method, targetPath, version := line[0], line[1], line[2]

	parsedRequestLine := new(RequestLine)
	parsedRequestLine.HttpVersion = string(bytes.Replace(version, []byte("HTTP/"), []byte(""), 1))
	parsedRequestLine.Method = string(method)
	parsedRequestLine.RequestTarget = string(targetPath)
	if err := parsedRequestLine.validate(); err != nil {
		return nil, 0, err
	}

	return parsedRequestLine, len(firstLine) + len(SEPARATOR), nil
}

func (rl *RequestLine) validate() error {
	if !slices.Contains(HTTP_METHODS, rl.Method) {
		return ErrInvalidHttpMethod
	}

	if !strings.HasPrefix(rl.RequestTarget, "/") {
		return ErrInvalidTargetPath
	}

	if rl.HttpVersion != VALID_HTTP_VERSION {
		return ErrUnsupportedHttpVersion
	}

	return nil
}
