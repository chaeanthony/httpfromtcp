package request

import (
	"bytes"
	"errors"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
	"strings"
)

type requestState int

const (
	requestStateInitialized requestState = iota
	requestStateParsingHeaders
	requestStateParsingBody
	requestStateDone
)

type Request struct {
	RequestLine    RequestLine
	state          requestState
	Headers        headers.Headers
	Body           []byte
	bodyLengthRead int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const crlf = "\r\n"
const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0 // index of the last byte read into buf
	req := &Request{
		state:   requestStateInitialized,
		Headers: headers.NewHeaders(),
		Body:    make([]byte, 0),
	}
	for req.state != requestStateDone {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		nBytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if req.state != requestStateDone {
					return nil, fmt.Errorf("incomplete request, in state: %d, read n bytes on EOF: %d", req.state, nBytesRead)
				}
				break
			}
			return nil, err
		}

		readToIndex += nBytesRead

		nBytesParsed, err := req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[nBytesParsed:]) // remove data that was parsed successfully to keep buffer small and memory-efficient
		readToIndex -= nBytesParsed
	}

	return req, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.state != requestStateDone { // headers and body requires multiple parses
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		if n == 0 {
			break
		}
		totalBytesParsed += n
	}
	return totalBytesParsed, nil
}

// parseSingle parses a single line of the request from the data buffer depending on the current state of the request.
// It returns the number of bytes consumed and an error if any.
// It updates the state of the request as it parses.
func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case requestStateDone:
		return 0, fmt.Errorf("trying to read data in done state")
	case requestStateInitialized:
		requestLine, bytesConsumed, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if bytesConsumed == 0 {
			return 0, nil // need more data to parse request line
		}
		r.RequestLine = *requestLine
		r.state = requestStateParsingHeaders
		return bytesConsumed, nil
	case requestStateParsingHeaders:
		bytesConsumed, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.state = requestStateParsingBody
		}
		return bytesConsumed, nil
	case requestStateParsingBody:
		val, ok := r.Headers.Get("content-length")
		if !ok {
			r.state = requestStateDone
			return len(data), nil // no body to parse
		}
		contentLength, err := strconv.Atoi(val)
		if err != nil {
			return 0, fmt.Errorf("malformed Content-Length: %v", err)
		}
		r.Body = append(r.Body, data...)
		r.bodyLengthRead += len(data)
		if r.bodyLengthRead > contentLength {
			return 0, fmt.Errorf("body %d does not match content-length size %d bytes", len(r.Body), contentLength)
		}
		if r.bodyLengthRead == contentLength {
			r.state = requestStateDone
		}
		return len(data), nil
	default:
		return 0, fmt.Errorf("unknown state %d", r.state)
	}
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return nil, 0, nil
	}

	requestLineText := string(data[:idx])
	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, 0, err
	}

	return requestLine, idx + len([]byte(crlf)), nil
}

func requestLineFromString(line string) (*RequestLine, error) {
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid request line: %s", line)
	}

	method := parts[0]
	for _, char := range method {
		if char < 'A' || char > 'Z' {
			return nil, fmt.Errorf("invalid method: %s", method)
		}
	}

	if parts[2] != "HTTP/1.1" {
		return nil, fmt.Errorf("invalid HTTP version: %s", parts[2])
	}
	httpVersion := strings.TrimPrefix(parts[2], "HTTP/")

	return &RequestLine{
		Method:        method,
		RequestTarget: parts[1],
		HttpVersion:   httpVersion,
	}, nil
}
