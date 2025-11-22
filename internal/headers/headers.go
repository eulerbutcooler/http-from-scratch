package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers struct {
	headers map[string]string
}

var rn = []byte("\r\n")

func NewHeaders() *Headers {
	return &Headers{
		headers: map[string]string{},
	}
}

func (h *Headers) Get(name string) string {
	return h.headers[strings.ToLower(name)]
}

func (h *Headers) Set(name, value string) {
	h.headers[strings.ToLower(name)] = value
}

func isToken(name string) bool {
	for _, ch := range name {
		found := false
		if ch > 'A' && ch < 'Z' || ch > 'a' && ch < 'z' || ch > '0' && ch < '9' {
			found = true
		}
		switch ch {
		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
			found = true
		}
		if !found {
			return false
		}
	}
	return true
}

func parseHeader(fieldLine []byte) (string, string, error) {
	name, val, found := bytes.Cut(fieldLine, []byte(":"))
	if found == true {
		val = bytes.TrimSpace(val)
		if !isToken(string(name)) {
			return "", "", fmt.Errorf("malformed header name")
		}
		if bytes.HasSuffix(name, []byte(" ")) {
			return "", "", fmt.Errorf("malformed field name")
		}
		return string(name), string(val), nil
	} else {
		return "", "", fmt.Errorf("malformed field line")
	}
}

func (h Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	done := false
	for {
		idx := bytes.Index(data[read:], rn)
		if idx == -1 {
			break
		}
		//Empty header
		if idx == 0 {
			done = true
			read += len(rn)
			break
		}
		name, value, err := parseHeader(data[read : read+idx])
		if err != nil {
			return 0, false, err
		}
		read += idx + len(rn)
		h.Set(name, value)
	}
	return read, done, nil

}
