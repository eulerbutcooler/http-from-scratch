package request

import (
	"bytes"
	"fmt"
	"http/internal/headers"
	"io"
	"strconv"
)

type parserState string

const (
	StateInit    parserState = "init"
	StateHeaders parserState = "headers"
	StateDone    parserState = "done"
	StateBody    parserState = "body"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	state       parserState
	headers     *headers.Headers
	body        string
}

func getInt(headers *headers.Headers, name string, defaultValue int) int {
	valStr, exists := headers.Get(name)
	if !exists {
		return defaultValue
	}
	value, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func newRequest() *Request {
	return &Request{
		state:   StateInit,
		headers: headers.NewHeaders(),
		body:    "",
	}
}

var ERROR_MALFORMED_REQUESTLINE = fmt.Errorf("malformed request-line")
var ERROR_UNSUPPORTED_HTTP_VERSION = fmt.Errorf("unsupported http version")
var SEPARATOR = []byte("\r\n")

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, SEPARATOR)
	if idx == -1 {
		return nil, 0, nil
	}
	startLine := b[:idx]
	read := idx + len(SEPARATOR)
	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, ERROR_MALFORMED_REQUESTLINE
	}
	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, ERROR_MALFORMED_REQUESTLINE
	}
	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}

	return rl, read, nil

}

func (r *Request) parse(data []byte) (int, error) {
	read := 0
outer:
	for {
		currentData := data[read:]
		switch r.state {
		case StateInit:
			rl, n, err := parseRequestLine(currentData)
			if err != nil {
				return 0, err
			}
			if n == 0 {
				break outer
			}
			r.RequestLine = *rl
			read += n
			r.state = StateHeaders
		case StateHeaders:
			n, done, err := r.headers.Parse(currentData)
			if err != nil {
				return 0, err
			}
			if n == 0 {
				break outer
			}

			read += n
			if done {
				r.state = StateBody
			}
		case StateBody:
			//currentData = current chunk of raw bytes being processed
			//length = total expected body size
			length := getInt(r.headers, "content-length", 0)
			if length == 0 {
				r.state = StateDone
				break
			}
			remaining := length - len(r.body)
			// toRead = data left to be read
			toRead := min(remaining, len(currentData))
			// r.body == string that accumulates the body data as its parsed
			r.body += string(currentData[:toRead])
			// read = counter tracking how many bytes have been consumed from currentData
			read += toRead
			if len(r.body) == length {
				r.state = StateDone
			}
		case StateDone:
			break outer
		}
	}
	return read, nil

}

func (r *Request) done() bool {
	return r.state == StateDone
}

func (r *Request) Headers() *headers.Headers {
	return r.headers
}

func (r *Request) Body() string {
	return r.body
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()
	buf := make([]byte, 1024)
	bufLen := 0
	for !request.done() {
		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			return nil, err
		}
		bufLen += n
		readN, err := request.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}
		copy(buf, buf[readN:bufLen])
		bufLen -= readN

	}

	return request, nil
}
