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
	CRLF               = "\r\n"
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

type parserState string

const (
	stateParserInit  parserState = "initialized"
	stateParserError parserState = "error"
	stateParserDone  parserState = "done"
)

type Request struct {
	RequestLine
	state parserState
}

func newRequest() *Request {
	return &Request{
		state: stateParserInit,
	}
}

func (r *Request) done() bool {
	return r.state == stateParserDone || r.state == stateParserError
}

func RequestFromReader(r io.Reader) (*Request, error) {
	req := newRequest()
	buf := make([]byte, 1024)
	readBytesLen := 0
	parsedBytesLen := 0
	for !req.done() {
		n, err := r.Read(buf[readBytesLen:])
		// TODO: What to do this this?
		if err != nil {
			return nil, fmt.Errorf("unable to read raw request: %w", err)
		}
		readBytesLen += n

		parsedN, err := req.parse(buf[:readBytesLen+n])
		if err != nil {
			return nil, err
		}
		copy(buf, buf[parsedN:readBytesLen])
		readBytesLen -= parsedN
		parsedBytesLen += parsedN

	}

	return req, nil
}

// data -> all currently unparsed bytes from the buffer
func (r *Request) parse(data []byte) (int, error) {
	read := 0
outer:
	for {
		switch r.state {
		case stateParserInit:
			rl, n, err := parseRequestLine(data)
			if err != nil {
				r.state = stateParserError
				return 0, err
			}

			// NOTE: Need to read more data, so go to parent loop to read more data
			if n == 0 {
				break outer
			}

			r.RequestLine = *rl
			read += n
			r.state = stateParserDone

		case stateParserDone:
			break outer

		}
	}
	return read, nil
}

func parseRequestLine(rawRequest []byte) (*RequestLine, int, error) {
	separatorIndex := bytes.Index(rawRequest, []byte(CRLF))
	if separatorIndex == -1 {
		return nil, 0, nil
	}
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

	return parsedRequestLine, separatorIndex + len(CRLF), nil
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
