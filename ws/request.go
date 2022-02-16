package ws

import (
	"bufio"
	"bytes"
	"errors"
	"net/http"
)

type Request struct {
	Method     string
	RequestURI string
	Proto      string
	Host       string
	Header     http.Header

	Reader *bufio.Reader
}

func (r *Request) readLine() ([]byte, error) {
	var line []byte
	for {
		l, more, err := r.Reader.ReadLine()
		if err != nil {
			return nil, err
		}
		if line == nil && !more {
			return l, nil
		}
		line = append(line, l...)
		if !more {
			break
		}
	}
	return line, nil
}

func (r *Request) parseRequestLine() (method, requestURI, proto string, ok bool) {
	var (
		err  error
		line []byte
	)
	if line, err = r.readLine(); err != nil {
		return
	}
	s1 := bytes.IndexByte(line, ' ')
	s2 := bytes.IndexByte(line[s1+1:], ' ')
	if s1 < 0 || s2 < 0 {
		return
	}
	s2 += s1 + 1
	return string(line[:s1]), string(line[s1+1 : s2]), string(line[s2+1:]), true
}

func (r *Request) readMIMEHeader() (hdr http.Header, err error) {
	var (
		line []byte
		i    int
		k, v string
	)
	hdr = make(http.Header, 16)
	for {
		if line, err = r.readLine(); err != nil {
			return
		}
		// line = trim(line)
		if len(line) == 0 {
			return
		}
		if i = bytes.IndexByte(line, ':'); i <= 0 {
			return hdr, errors.New("malformed MIME header line")
		}
		k = string(line[:i])
		// skip initial spaces in value
		i++
		for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
			i++
		}
		v = string(line[i:])
		hdr.Add(k, v)
	}
}

func ReadRequest(br *bufio.Reader) (r *Request, err error) {
	var (
		ok bool
	)
	r = &Request{
		Reader: br,
	}
	if r.Method, r.RequestURI, r.Proto, ok = r.parseRequestLine(); !ok {
		return nil, errors.New("malformed HTTP request")
	}
	if r.Header, err = r.readMIMEHeader(); err != nil {
		return
	}
}
