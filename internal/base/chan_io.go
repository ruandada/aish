package base

import (
	"bytes"
	"io"
	"sync"
)

type ChanIO struct {
	mu   sync.Mutex
	ch   chan []byte
	rest []byte
	buf  *bytes.Buffer
}

// Close implements io.ReadWriteCloser.
func (cio *ChanIO) Close() error {
	close(cio.ch)
	return nil
}

// Write implements io.Writer.
func (cio *ChanIO) Write(p []byte) (n int, err error) {
	cio.ch <- p
	cio.buf.Write(p)
	return len(p), nil
}

// Read implements io.Reader.
func (cio *ChanIO) Read(p []byte) (n int, err error) {
	cio.mu.Lock()
	if len(cio.rest) > 0 {
		n = copy(p, cio.rest)
		cio.rest = cio.rest[n:]

		cio.mu.Unlock()
		return n, nil
	}
	cio.mu.Unlock()

	b, ok := <-cio.ch
	if !ok {
		return 0, io.EOF
	}

	cio.mu.Lock()
	n = copy(p, b)
	if n < len(b) {
		cio.rest = append(cio.rest, b[n:]...)
	}
	cio.mu.Unlock()

	return n, nil
}

func (cio *ChanIO) Bytes() []byte {
	return cio.buf.Bytes()
}

func NewChanIO(chanBufSize int) *ChanIO {
	return &ChanIO{
		ch:   make(chan []byte, chanBufSize),
		rest: nil,
		buf:  bytes.NewBuffer(nil),
	}
}

var _ io.ReadWriteCloser = (*ChanIO)(nil)
