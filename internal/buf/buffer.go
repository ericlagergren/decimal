package buf

import (
	"io"
	"unsafe"
)

type B struct {
	buf []byte
}

func New(size int) *B { return &B{buf: make([]byte, 0, size)} }

func (b *B) Bytes() []byte { return b.buf }

func (b *B) Len() int { return len(b.buf) }

func (b *B) Write(p []byte) (int, error) {
	b.buf = append(b.buf, p...)
	return len(p), nil
}

func (b *B) WriteByte(c byte) error {
	b.buf = append(b.buf, c)
	return nil
}

func (b *B) WriteString(s string) (int, error) {
	b.buf = append(b.buf, s...)
	return len(s), nil
}

func (b *B) WriteTo(w io.Writer) (int, error) { return w.Write(b.buf) }

func (b *B) String() string { return *(*string)(unsafe.Pointer(&b.buf)) }
