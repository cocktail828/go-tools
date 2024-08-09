package lumberjack

import (
	"io"
)

// flush writes any buffered data to the underlying io.Writer.
func (l *Logger) flush() error {
	if l.n == 0 {
		return nil
	}
	n, err := l.file.Write(l.buf[0:l.n])
	if n < l.n && err == nil {
		err = io.ErrShortWrite
	}
	if err != nil {
		if n > 0 && n < l.n {
			copy(l.buf[0:l.n-n], l.buf[n:l.n])
		}
		l.n -= n
		return err
	}
	l.n = 0
	return nil
}

// available returns how many bytes are unused in the buffer.
func (l *Logger) available() int { return len(l.buf) - l.n }

// buffered returns the number of bytes that have been written into the current buffer.
func (l *Logger) buffered() int { return l.n }

// Write writes the contents of p into the buffer.
// It returns the number of bytes written.
// If nn < len(p), it also returns an error explaining
// why the write is short.
func (l *Logger) writeBuffer(p []byte) (nn int, err error) {
	for len(p) > l.available() {
		var n int
		if l.buffered() == 0 {
			// Large write, empty buffer.
			// Write directly from p to avoid copy.
			n, err = l.file.Write(p)
		} else {
			n = copy(l.buf[l.n:], p)
			l.n += n
			err = l.flush()
		}
		nn += n
		p = p[n:]

		if err != nil {
			return nn, err
		}
	}
	n := copy(l.buf[l.n:], p)
	l.n += n
	nn += n
	return nn, nil
}
