/*
	ioutils

*/
package ioutils

import (
	"bufio"
)

func Pop(br *bufio.Reader, n int) (bs []byte, err error) {
	bs, err = br.Peek(n)
	if err == nil {
		_, err = br.Discard(n)
	}
	return
}
