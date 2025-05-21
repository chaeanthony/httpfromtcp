package headers

import (
	"bytes"
	"fmt"
	"strings"
)

const (
	crlf = "\r\n"
)

type Headers map[string]string

func NewHeaders() Headers {
	return map[string]string{}
}

func (h Headers) Get(key string) (string, bool) {
	key = strings.ToLower(key)
	val, ok := h[key]
	return val, ok
}

func (h Headers) Set(key, value string) {
	key = strings.ToLower(key)
	v, ok := h[key]
	if ok {
		value = v + ", " + value
	}
	h[key] = value
}

func (h Headers) Override(key, value string) {
	key = strings.ToLower(key)
	h[key] = value
}

func (h Headers) Remove(key string) {
	key = strings.ToLower(key)
	delete(h, key)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return 0, false, nil // not enough data to parse
	}
	if idx == 0 {
		return len([]byte(crlf)), true, nil // end of headers
	}

	headerParts := bytes.SplitN(data[:idx], []byte(":"), 2)
	if len(headerParts) != 2 {
		return 0, false, fmt.Errorf("invalid header format: %s", string(data[:idx]))
	}

	keyBytes := headerParts[0]
	if bytes.HasSuffix(keyBytes, []byte(" ")) {
		return 0, false, fmt.Errorf("invalid header format: key has trailing space")
	}

	key := string(bytes.ToLower(bytes.TrimSpace(keyBytes)))
	if !validFieldName(key) {
		return 0, false, fmt.Errorf("invalid header key %s: ", key)
	}
	val := string(bytes.TrimSpace(headerParts[1]))

	if _, ok := h[key]; ok {
		h[key] = h[key] + ", " + val
	} else {
		h[key] = val
	}
	return idx + len([]byte(crlf)), false, nil
}

func validFieldName(name string) bool {
	if len(name) == 0 {
		return false
	}
	for _, c := range name {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '!' || c == '#' || c == '$' || c == '%' ||
			c == '&' || c == '\'' || c == '*' || c == '+' || c == '-' || c == '.' || c == '^' || c == '_' || c == '`' || c == '|' || c == '~') {
			return false
		}
	}
	return true
}
