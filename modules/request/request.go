package request

import (
	"bufio"
	"errors"
	"net"
	"strconv"
	"strings"
)

type Request struct {
	Method        string
	Path          string
	Headers       map[string]string
	Body          string
	ContentLength uint
}

func New() Request {
	return Request{
		Headers: make(map[string]string),
	}
}

func From(conn net.Conn) (Request, error) {
	req := New()
	reader := bufio.NewReader(conn)

	// read the request line
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return req, errors.New("failed to read request line")
	}
	requestLine = strings.TrimSpace(requestLine)
	parts := strings.Split(requestLine, " ")

	if len(parts) != 3 {
		return req, errors.New("invalid HTTP request line")
	}

	req.Method = parts[0]

	// our server will only support these three methods
	if req.Method != "GET" && req.Method != "POST" && req.Method != "DELETE" {
		return req, errors.New("405 Method Not Allowed")
	}

	req.Path = parts[1]
	if !strings.HasPrefix(req.Path, "/") {
		return req, errors.New("404 Not Found")
	}

	// do not allow through any other HTTP versions
	if parts[2] != "HTTP/1.1" {
		return req, errors.New("505 HTTP Version Not Supported")
	}

	// read headers
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return req, errors.New("failed to read header line")
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}

		headerParts := strings.SplitN(line, ": ", 2)
		if len(headerParts) == 2 {
			req.Headers[headerParts[0]] = headerParts[1]
		}
	}

	// read Content-Length if it is present
	if contentLength, ok := req.Headers["Content-Length"]; ok {
		length, err := strconv.Atoi(contentLength)
		if err != nil {
			return req, errors.New("invalid Content-Length value")
		}
		req.ContentLength = uint(length)

		// read body
		body := make([]byte, req.ContentLength)
		_, err = reader.Read(body)
		if err != nil {
			return req, errors.New("failed to read body")
		}
		req.Body = string(body)
	}

	return req, nil
}

// utility method to get cookies of a request as a map
func (r Request) GetCookies() (map[string]string, error) {
	cookies := map[string]string{}

	if cookiesHeader, ok := r.Headers["Cookie"]; ok {
		for _, cookie := range strings.Split(cookiesHeader, ";") {
			if key, value, flag := strings.Cut(cookie, "="); flag {
				cookies[key] = value
			}
		}
	} else {
		return nil, errors.New("cookies not found")
	}

	return cookies, nil
}
