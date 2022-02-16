package ws

import (
	"bufio"
	"errors"
	"io"
	"strings"
)

const (
	WsResponse = "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n"
	WsAccept   = "Sec-WebSocket-Accept: %s\r\n\r\n"
)

func Upgrade(rwc io.ReadWriteCloser, br *bufio.Reader, bw *bufio.Writer, r *Request) (conn *Conn, err error) {
	if r.Method != "GET" {
		return nil, errors.New("bad method")
	}
	if r.Header.Get("Sec-Websocket-Version") != "13" {
		return nil, errors.New("missing or bad ws version")
	}
	if strings.ToLower(r.Header.Get("Upgrade")) != "websocket" {
		return nil, errors.New("not ws protocol")
	}

	challengeKey := r.Header.Get("Sec-Websocket-Key")
	if challengeKey == "" {
		return nil, errors.New("mismatch challenge")
	}

	_, _ = bw.WriteString(WsResponse)
	_, _ = bw.WriteString(WsAccept)
	if err = bw.Flush(); err != nil {
		return
	}
	return newConn(rwc, br, bw), nil
}
