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
	"fmt"
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

// ErrorHandler handles an error and returns whether
// the mux should continue serving the listener.
type ErrorHandler func(error) bool

var _ net.Error = ErrNotMatched{}

// ErrNotMatched is returned whenever a connection is not matched by any of
// the matchers registered in the multiplexer.
type ErrNotMatched struct {
	c net.Conn
}

func (e ErrNotMatched) Error() string {
	return fmt.Sprintf("mux: connection %v not matched by an matcher",
		e.c.RemoteAddr())
}

// Temporary implements the net.Error interface.
func (e ErrNotMatched) Temporary() bool { return true }

// Timeout implements the net.Error interface.
func (e ErrNotMatched) Timeout() bool { return false }

// ErrServerClosed is returned from muxListener.Accept when mux server is closed.
var ErrServerClosed = errors.New("mux: server closed")

// for readability of readTimeout
var noTimeout time.Duration

// New instantiates a new connection multiplexer.
func New(l net.Listener) CMux {
	ctx, cancel := context.WithCancel(context.Background())
	return &cMux{
		runningCtx:  ctx,
		cancel:      cancel,
		root:        l,
		errh:        func(_ error) bool { return true },
		readTimeout: noTimeout,
	}
}

// CMux is a multiplexer for network connections.
type CMux interface {
	// Match returns a net.Listener that sees (i.e., accepts) only
	// the connections matched by at least one of the matcher.
	//
	// The order used to call Match determines the priority of matchers.
	Match(...Matcher) net.Listener
	// MatchWithWriters returns a net.Listener that accepts only the
	// connections that matched by at least of the matcher writers.
	//
	// Prefer Matchers over MatchWriters, since the latter can write on the
	// connection before the actual handler.
	//
	// The order used to call Match determines the priority of matchers.
	MatchWithWriters(...MatchWriter) net.Listener
	// Serve starts multiplexing the listener. Serve blocks and perhaps
	// should be invoked concurrently within a go routine.
	Serve() error
	// Closes cmux server and stops accepting any connections on listener
	Close()
	// HandleError registers an error handler that handles listener errors.
	HandleError(ErrorHandler)
	// sets a timeout for the read of matchers
	SetReadTimeout(time.Duration)
}

type cMux struct {
	runningCtx  context.Context
	cancel      context.CancelFunc
	root        net.Listener
	errh        ErrorHandler
	children    []*muxListener
	readTimeout time.Duration
}

func matchersToMatchWriters(matchers []Matcher) []MatchWriter {
	mws := make([]MatchWriter, 0, len(matchers))
	for _, m := range matchers {
		cm := m
		mws = append(mws, func(w io.Writer, r io.Reader) bool {
			return cm(r)
		})
	}
	return mws
}

func (m *cMux) Match(matchers ...Matcher) net.Listener {
	mws := matchersToMatchWriters(matchers)
	return m.MatchWithWriters(mws...)
}

func (m *cMux) MatchWithWriters(matchers ...MatchWriter) net.Listener {
	subCtx, cancel := context.WithCancel(m.runningCtx)
	ml := &muxListener{
		Listener:   m.root,
		matchers:   matchers,
		runningCtx: subCtx,
		cancel:     cancel,
		connc:      make(chan net.Conn, 1024),
	}
	m.children = append(m.children, ml)
	return ml
}

func (m *cMux) SetReadTimeout(t time.Duration) {
	m.readTimeout = t
}

func (m *cMux) Serve() error {
	var wg sync.WaitGroup

	defer func() {
		m.cancel()
		wg.Wait()
		for _, sl := range m.children {
			sl.Close()
		}
	}()

	for {
		c, err := m.root.Accept()
		if err != nil {
			if !m.handleErr(err) {
				return err
			}
			continue
		}

		wg.Add(1)
		go m.serve(c, &wg)
	}
}

func (m *cMux) serve(c net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	muc := newMuxConn(c)
	if m.readTimeout > noTimeout {
		_ = c.SetReadDeadline(time.Now().Add(m.readTimeout))
	}

	for _, sl := range m.children {
		if sl.isClosed.Load() {
			continue
		}
		for _, s := range sl.matchers {
			matched := s(muc.Conn, muc.startSniffing())
			if matched {
				muc.doneSniffing()
				if m.readTimeout > noTimeout {
					_ = c.SetReadDeadline(time.Time{})
				}
				sl.onMatch(muc)
				return
			}
		}
	}

	_ = c.Close()
	if !m.handleErr(ErrNotMatched{c: c}) {
		_ = m.root.Close()
	}
}

func (m *cMux) Close() {
	m.cancel()
}

func (m *cMux) HandleError(h ErrorHandler) {
	m.errh = h
}

func (m *cMux) handleErr(err error) bool {
	if !m.errh(err) {
		return false
	}

	if ne, ok := err.(net.Error); ok {
		return ne.Temporary()
	}

	return false
}

type muxListener struct {
	net.Listener
	isClosed   atomic.Bool
	matchers   []MatchWriter
	runningCtx context.Context
	cancel     context.CancelFunc
	connc      chan net.Conn
}

func (l *muxListener) onMatch(muc *MuxConn) {
	if l.isClosed.Load() {
		_ = muc.Close()
		return
	}

	select {
	case l.connc <- muc:
	case <-l.runningCtx.Done():
		_ = muc.Close()
	}
}

// Accept waits for and returns the next connection to the listener.
func (l *muxListener) Accept() (net.Conn, error) {
	select {
	case c := <-l.connc:
		return c, nil
	case <-l.runningCtx.Done():
		l.isClosed.Store(true)
		return nil, ErrServerClosed
	}
}

// Close closes the root.
// Any blocked Accept operations will be unblocked and return errors.
func (l *muxListener) Close() error {
	l.cancel()
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
