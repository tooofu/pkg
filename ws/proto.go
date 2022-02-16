package ws

const (
	// RFC 6455, Section 5.2
	// https://datatracker.ietf.org/doc/html/rfc6455#section-5.2
	// Frame header byte 0 bits
	finBit  = 1 << 7
	rsv1Bit = 1 << 6
	rsv2Bit = 1 << 5
	rsv3Bit = 1 << 4
	opBit   = 0x0f
	// Frame header byte 1 bits
	maskBit = 1 << 7
	lenBit  = 0x7f
)

/*
0               7               15                              31
+-+-+-+-+-------+-+-------------+-------------------------------+
|F|R|R|R| opcode|M| Payload len |    Extended payload length    |
|I|S|S|S|  (4)  |A|     (7)     |             (16/64)           |
|N|V|V|V|       |S|             |   (if payload len==126/127)   |
| |1|2|3|       |K|             |                               |
+-+-+-+-+-------+-+-------------+ - - - - - - - - - - - - - - - +
|     Extended payload length continued, if payload len == 127  |
+ - - - - - - - - - - - - - - - +-------------------------------+
|                               |Masking-key, if MASK set to 1  |
+-------------------------------+-------------------------------+
| Masking-key (continued)       |          Payload Data         |
+-------------------------------- - - - - - - - - - - - - - - - +
:                     Payload Data continued ...                :
+ - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - +
|                     Payload Data continued ...                |
+---------------------------------------------------------------+
*/
type Proto struct {
	FinRsvOpcode int8
	MaskSize     int8
	ExtSize      uint16
	Data         []byte
}
