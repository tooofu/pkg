/*
	ref:
		github.com/moby/moby/pkg/pools/pools.go


	1. BufioReaderPool & BufioWriterPool
	TODO: 2. NewReaderCloserWrapper & NewWriterCloserWrapper
	3. []byte pool
	TODO: 4. reader/writer copy by buffer
*/

package pools

import (
	"bufio"
	"io"
	"sync"
)

const buffer4K = 4 * 1024

var (
	Buffer4KPool      = newBufferPoolWithSize(buffer4K)
	BufioReader4KPool = newBufioReaderPoolWithSize(buffer4K)
	BufioWriter4KPool = newBufioWriterPoolWithSize(buffer4K)
)

type BufferPool struct {
	pool sync.Pool
}

func newBufferPoolWithSize(size int) *BufferPool {
	return &BufferPool{
		pool: sync.Pool{
			New: func() interface{} { s := make([]byte, size); return &s },
		},
	}

}

func (bp *BufferPool) Get() *[]byte {
	return bp.pool.Get().(*[]byte)

}

func (bp *BufferPool) Put(bs *[]byte) {
	bp.pool.Put(bs)

}

type BufioReaderPool struct {
	pool sync.Pool
}

func newBufioReaderPoolWithSize(size int) *BufioReaderPool {
	return &BufioReaderPool{
		pool: sync.Pool{
			New: func() interface{} { return bufio.NewReaderSize(nil, size) },
		},
	}

}

func (brp *BufioReaderPool) Get(r io.Reader) *bufio.Reader {
	buf := brp.pool.Get().(*bufio.Reader)
	buf.Reset(r)
	return buf

}

func (brp *BufioReaderPool) Put(br *bufio.Reader) {
	br.Reset(nil)
	brp.pool.Put(br)

}

type BufioWriterPool struct {
	pool sync.Pool
}

func newBufioWriterPoolWithSize(size int) *BufioWriterPool {
	return &BufioWriterPool{
		pool: sync.Pool{
			New: func() interface{} { return bufio.NewWriterSize(nil, size) },
		},
	}

}

func (bwp *BufioWriterPool) Get(w io.Writer) *bufio.Writer {
	buf := bwp.pool.Get().(*bufio.Writer)
	buf.Reset(w)
	return buf

}

func (bwp *BufioWriterPool) Put(bw *bufio.Writer) {
	bw.Reset(nil)
	bwp.pool.Put(bw)

}
