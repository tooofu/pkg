package ws

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
)

const (
	continuationFrame        = 0
	continuationFrameMaxRead = 100
)

type Conn struct {
	rwc io.ReadWriteCloser

	Reader *bufio.Reader
	Writer *bufio.Writer

	maskKey []byte
}

func newConn(rwc io.ReadWriteCloser, br *bufio.Reader, bw *bufio.Writer) *Conn {
	return &Conn{
		rwc:     rwc,
		Reader:  br,
		Writer:  bw,
		maskKey: make([]byte, 4),
	}
}

func (c *Conn) WriteMessage(msgType int, msg []byte) (err error) {
	if err = c.WriteHeader(msgType, len(msg)); err != nil {
		return
	}
	err = c.WriteBody(msg)
	return
}

func (c *Conn) WriteHeader(msgType int, length int) (err error) {
	hdr := make([]byte, 10)
	// 1. First byte =  | FIN | RSV1 | RSV2 | RSV3 | OpCode(4bits) |
	hdr[0] = 0
	hdr[0] |= finBit | byte(msgType)
	// 2. Second byte = | Mask | Payload len(7bits) |
	hdr[1] = 0
	switch {
	case length <= 125:
		// 7 bits
		hdr[1] |= byte(length)
		c.Writer.Write(hdr[:2])
	case length <= 65536:
		// 16 bits
		hdr[1] |= 126
		binary.BigEndian.PutUint16(hdr[2:], uint16(length))
		c.Writer.Write(hdr[:4])
	default:
		// 64bits
		hdr[1] |= 127
		binary.BigEndian.PutUint64(hdr[2:], uint64(length))
		c.Writer.Write(hdr)
	}
	return
}

func (c *Conn) WriteBody(msg []byte) (err error) {
	if len(msg) > 0 {
		_, err = c.Writer.Write(msg)
	}
	return
}

func (c *Conn) ReadMessage() (op int, payload []byte, err error) {
	var (
	// fin         bool
	// finOp, n    int
	// partPayload []byte
	)

	// for {
	//     // read frame
	//     if fin, op, partPayload, err = c.readFrame(); err != nil {

	//     }
	// }
	return
}

func (c *Conn) readFrame() (fin bool, op int, payload []byte, err error) {
	var (
		b byte
		// p          []byte
		// mask       bool
		// maskKey    []byte
		// payloadLen int64
	)
	// 1. First byte
	if b, err = c.Reader.ReadByte(); err != nil {
		return
	}
	// final frame
	fin = (b & finBit) != 0
	// rsv MUST be 0
	if rsv := b & (rsv1Bit | rsv2Bit | rsv3Bit); rsv != 0 {
		return false, 0, nil, errors.New("unexpected reserved bits")
	}
	// op code
	return
}
