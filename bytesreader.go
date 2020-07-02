package pgrest

import (
	"bytes"
	"errors"
	"io"
)

// BytesReader is reader implementation
type BytesReader struct {
	s []byte
	i int
}

// NewBytesReader creates new instance
func NewBytesReader(b []byte) *BytesReader {
	return &BytesReader{
		s: b,
	}
}

// Reset resets
func (r *BytesReader) Reset(b []byte) {
	r.s = b
	r.i = 0
}

// Buffered executes buffered
func (r *BytesReader) Buffered() int {
	return len(r.s) - r.i
}

// Bytes gets bytes
func (r *BytesReader) Bytes() []byte {
	return r.s[r.i:]
}

// Read reads
func (r *BytesReader) Read(b []byte) (n int, err error) {
	if r.i >= len(r.s) {
		return 0, io.EOF
	}
	n = copy(b, r.s[r.i:])
	r.i += n
	return
}

// ReadByte reads byte
func (r *BytesReader) ReadByte() (byte, error) {
	if r.i >= len(r.s) {
		return 0, io.EOF
	}
	b := r.s[r.i]
	r.i++
	return b, nil
}

// UnreadByte undreads byte
func (r *BytesReader) UnreadByte() error {
	if r.i <= 0 {
		return errors.New("UnreadByte: at beginning of slice")
	}
	r.i--
	return nil
}

// ReadSlice reads slice
func (r *BytesReader) ReadSlice(delim byte) ([]byte, error) {
	if i := bytes.IndexByte(r.s[r.i:], delim); i >= 0 {
		i++
		line := r.s[r.i : r.i+i]
		r.i += i
		return line, nil
	}

	line := r.s[r.i:]
	r.i = len(r.s)
	return line, io.EOF
}

// ReadBytes reads bytes
func (r *BytesReader) ReadBytes(fn func(byte) bool) ([]byte, error) {
	for i, c := range r.s[r.i:] {
		if !fn(c) {
			i++
			line := r.s[r.i : r.i+i]
			r.i += i
			return line, nil
		}
	}

	line := r.s[r.i:]
	r.i = len(r.s)
	return line, io.EOF
}

// Discard discards
func (r *BytesReader) Discard(n int) (int, error) {
	b, err := r.ReadN(n)
	return len(b), err
}

// ReadN reads int
func (r *BytesReader) ReadN(n int) ([]byte, error) {
	nn := n
	if nn > len(r.s) {
		nn = len(r.s)
	}

	b := r.s[r.i : r.i+nn]
	r.i += nn
	if n > nn {
		return b, io.EOF
	}
	return b, nil
}

// ReadFull reads full
func (r *BytesReader) ReadFull() ([]byte, error) {
	b := make([]byte, len(r.s)-r.i)
	copy(b, r.s[r.i:])
	r.i = len(r.s)
	return b, nil
}

// ReadFullTemp reads full temp
func (r *BytesReader) ReadFullTemp() ([]byte, error) {
	b := r.s[r.i:]
	r.i = len(r.s)
	return b, nil
}
