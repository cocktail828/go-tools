// Copyright 2016 The CMux Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package cmux

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// Matcher matches a connection based on its content.
type Matcher func(io.Reader) bool

// MatchWriter is a match that can also write response (say to do handshake).
type MatchWriter func(io.Writer, io.Reader) bool

// ErrServerClosed is returned from muxListener.Accept when mux server is closed.
var ErrServerClosed = errors.New("mux: server closed")
var ErrMissMatch = errors.New("mux: no handler for the conn")

type CMux struct {
	ctx              context.Context
	cancel           context.CancelCauseFunc
	root             net.Listener
	children         []*muxListener
	readTimeout      time.Duration
	abortOnMissMatch bool
}

// New instantiates a new connection multiplexer.
func New(ln net.Listener) *CMux {
	ctx, cancel := context.WithCancelCause(context.Background())
	return &CMux{
		ctx:    ctx,
		cancel: cancel,
		root:   ln,
	}
}

func (m *CMux) Match(matchers ...Matcher) net.Listener {
	mws := make([]MatchWriter, 0, len(matchers))
	for _, m := range matchers {
		cm := m
		mws = append(mws, func(w io.Writer, r io.Reader) bool { return cm(r) })
	}
	return m.MatchWithWriters(mws...)
}

func (m *CMux) MatchWithWriters(matchers ...MatchWriter) net.Listener {
	ml := &muxListener{
		Listener: m.root,
		matchers: matchers,
		connc:    make(chan net.Conn, 1024),
	}
	m.children = append(m.children, ml)
	return ml
}

func (m *CMux) AbortOnMissMatch() {
	m.abortOnMissMatch = true
}

func (m *CMux) SetReadTimeout(t time.Duration) {
	m.readTimeout = t
}

func (m *CMux) Close() {
	m.cancel(nil)
}

func (m *CMux) Serve() error {
loop:
	for {
		select {
		case <-m.ctx.Done():
			break loop
		default:
			c, err := m.root.Accept()
			if err != nil {
				m.cancel(err)
				break loop
			}
			go m.serve(c)
		}
	}

	for _, sl := range m.children {
		sl.Close()
	}
	return context.Cause(m.ctx)
}

func (m *CMux) serve(c net.Conn) {
	muc := newMuxConn(c)
	if m.readTimeout > 0 {
		_ = c.SetReadDeadline(time.Now().Add(m.readTimeout))
	}

	for _, sl := range m.children {
		for _, s := range sl.matchers {
			matched := s(muc.Conn, muc.startSniffing())
			if matched {
				muc.doneSniffing()
				if m.readTimeout > 0 {
					_ = c.SetReadDeadline(time.Time{})
				}
				sl.onMatch(muc)
				return
			}
		}
	}

	_ = c.Close()
	if m.abortOnMissMatch {
		m.cancel(ErrMissMatch)
	}
}

type muxListener struct {
	net.Listener
	isClosed atomic.Bool
	mu       sync.RWMutex
	matchers []MatchWriter
	connc    chan net.Conn
}

func (l *muxListener) onMatch(muc *MuxConn) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.isClosed.Load() {
		_ = muc.Close()
		return
	}
	l.connc <- muc
}

// Accept waits for and returns the next connection to the listener.
func (l *muxListener) Accept() (net.Conn, error) {
	if c, ok := <-l.connc; ok {
		return c, nil
	}
	return nil, ErrServerClosed
}

// Close closes the muxListener.
// Any blocked Accept operations will be unblocked and return errors.
func (l *muxListener) Close() error {
	if l.isClosed.CompareAndSwap(false, true) {
		l.mu.Lock()
		close(l.connc)
		l.mu.Unlock()
	}
	return nil
}

// MuxConn wraps a net.Conn and provides transparent sniffing of connection data.
type MuxConn struct {
	net.Conn
	buf bufferedReader
}

func newMuxConn(c net.Conn) *MuxConn {
	return &MuxConn{
		Conn: c,
		buf:  bufferedReader{source: c},
	}
}

// From the io.Reader documentation:
//
// When Read encounters an error or end-of-file condition after
// successfully reading n > 0 bytes, it returns the number of
// bytes read.  It may return the (non-nil) error from the same call
// or return the error (and n == 0) from a subsequent call.
// An instance of this general case is that a Reader returning
// a non-zero number of bytes at the end of the input stream may
// return either err == EOF or err == nil.  The next Read should
// return 0, EOF.
func (m *MuxConn) Read(p []byte) (int, error) {
	return m.buf.Read(p)
}

func (m *MuxConn) startSniffing() io.Reader {
	m.buf.reset(true)
	return &m.buf
}

func (m *MuxConn) doneSniffing() {
	m.buf.reset(false)
}
