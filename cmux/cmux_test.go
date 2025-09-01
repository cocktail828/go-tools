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
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"errors"
	"fmt"
	"go/build"
	"io"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"
)

const (
	testHTTP1Resp = "http1"
	rpcVal        = 1234
)

func abortOnErr(t *testing.T, errCh <-chan error) {
	select {
	case err := <-errCh:
		t.Fatal(err)
	default:
	}
}

func safeServe(errCh chan<- error, muxl *CMux) {
	if err := muxl.Serve(); err != nil &&
		!slices.Contains([]error{context.Canceled, net.ErrClosed}, err) &&
		!strings.Contains(err.Error(), "use of closed") {
		errCh <- err
	}
}

func safeDial(t *testing.T, addr net.Addr) (*rpc.Client, func()) {
	c, err := rpc.Dial(addr.Network(), addr.String())
	if err != nil {
		t.Fatal(err)
	}
	return c, func() {
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

type chanListener struct {
	connCh chan net.Conn
}

func newChanListener() *chanListener {
	return &chanListener{connCh: make(chan net.Conn, 1)}
}

func (l *chanListener) Accept() (net.Conn, error) {
	if c, ok := <-l.connCh; ok {
		return c, nil
	}
	return nil, net.ErrClosed
}

func (l *chanListener) Addr() net.Addr {
	return &net.TCPAddr{}
}

func (l *chanListener) Close() error {
	return nil
}

func testListener(t *testing.T) (net.Listener, func()) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	return l, func() {
		if err := l.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

type testHTTP1Handler struct{}

func (h *testHTTP1Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, testHTTP1Resp)
}

func runTestHTTPServer(errCh chan<- error, l net.Listener) {
	var mu sync.Mutex
	conns := make(map[net.Conn]struct{})

	defer func() {
		mu.Lock()
		for c := range conns {
			if err := c.Close(); err != nil {
				errCh <- err
			}
		}
		mu.Unlock()
	}()

	s := &http.Server{
		Handler: &testHTTP1Handler{},
		ConnState: func(c net.Conn, state http.ConnState) {
			mu.Lock()
			switch state {
			case http.StateNew:
				conns[c] = struct{}{}
			case http.StateClosed:
				delete(conns, c)
			}
			mu.Unlock()
		},
	}
	if err := s.Serve(l); err != ErrServerClosed {
		errCh <- err
	}
}

func runTestTLSServer(errCh chan<- error, l net.Listener) {
	certificate, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		errCh <- err
		return
	}

	runTestHTTPServer(errCh,
		tls.NewListener(l, &tls.Config{
			Certificates: []tls.Certificate{certificate},
			Rand:         rand.Reader,
		}))
}

func runTestHTTP1Client(t *testing.T, addr net.Addr) {
	runTestHTTPClient(t, "http", addr)
}

func runTestTLSClient(t *testing.T, addr net.Addr) {
	runTestHTTPClient(t, "https", addr)
}

func runTestHTTPClient(t *testing.T, proto string, addr net.Addr) {
	client := http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	r, err := client.Get(proto + "://" + addr.String())
	if err != nil {
		t.Fatal(err)
	}
	defer r.Body.Close()

	b, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != testHTTP1Resp {
		t.Fatalf("invalid response: want=%s got=%s", testHTTP1Resp, b)
	}
}

type TestRPCRcvr struct{}

func (r TestRPCRcvr) Test(i int, j *int) error {
	*j = i
	return nil
}

func runTestRPCServer(errCh chan<- error, l net.Listener) {
	s := rpc.NewServer()
	if err := s.Register(TestRPCRcvr{}); err != nil {
		errCh <- err
	}
	for {
		c, err := l.Accept()
		if err != nil {
			if err != ErrServerClosed {
				errCh <- err
			}
			return
		}
		go s.ServeConn(c)
	}
}

func runTestRPCClient(t *testing.T, addr net.Addr) {
	c, cleanup := safeDial(t, addr)
	defer cleanup()

	var num int
	if err := c.Call("TestRPCRcvr.Test", rpcVal, &num); err != nil {
		t.Fatal(err)
	}

	if num != rpcVal {
		t.Errorf("wrong rpc response: want=%d got=%v", rpcVal, num)
	}
}

const (
	handleHTTP1Close   = 1
	handleHTTP1Request = 2
	handleAnyClose     = 3
	handleAnyRequest   = 4
)

func TestTimeout(t *testing.T) {
	defer leakCheck(t)()
	lis, closer := testListener(t)

	result := make(chan int, 5)
	testDuration := time.Millisecond * 100
	m := New(lis)
	m.SetReadTimeout(testDuration)
	http1 := m.Match(HTTP1Fast())
	anyl := m.Match(Any())
	go m.Serve()
	go func() {
		con, err := http1.Accept()
		if err != nil {
			result <- handleHTTP1Close
		} else {
			_, _ = con.Write([]byte("http1"))
			_ = con.Close()
			result <- handleHTTP1Request
		}
	}()
	go func() {
		con, err := anyl.Accept()
		if err != nil {
			result <- handleAnyClose
		} else {
			_, _ = con.Write([]byte("any"))
			_ = con.Close()
			result <- handleAnyRequest
		}
	}()
	time.Sleep(testDuration) // wait to prevent timeouts on slow test-runners
	client, err := net.Dial("tcp", lis.Addr().String())
	if err != nil {
		t.Fatal("testTimeout client failed: ", err)
	}
	defer client.Close()

	time.Sleep(testDuration / 2)
	if len(result) != 0 {
		t.Fatal("testTimeout failed: accepted(tcp) to fast: ", len(result))
	}
	_ = client.SetReadDeadline(time.Now().Add(testDuration * 3))
	buffer := make([]byte, 10)
	rl, err := client.Read(buffer)
	if err != nil {
		t.Fatal("testTimeout failed: client error: ", err, rl)
	}
	closer()
	if rl != 3 {
		t.Fatal("testTimeout failed: response from wrong service ", rl)
	}
	if string(buffer[0:3]) != "any" {
		t.Fatal("testTimeout failed: response from wrong service ")
	}
	time.Sleep(testDuration * 2)
	if len(result) != 2 {
		t.Fatal("testTimeout failed: accepted to less: ", len(result))
	}
	if a := <-result; a != handleAnyRequest {
		t.Fatal("testTimeout failed: any rule did not match")
	}
	if a := <-result; a != handleHTTP1Close {
		t.Fatal("testTimeout failed: no close an http rule")
	}
}

func TestRead(t *testing.T) {
	defer leakCheck(t)()
	errCh := make(chan error)
	defer abortOnErr(t, errCh)
	const payload = "hello world\r\n"
	const mult = 2

	writer, reader := net.Pipe()
	go func() {
		if _, err := io.WriteString(writer, strings.Repeat(payload, mult)); err != nil {
			panic(err)
		}
		if err := writer.Close(); err != nil {
			panic(err)
		}
	}()

	l := newChanListener()
	defer close(l.connCh)
	l.connCh <- reader
	muxl := New(l)
	// Register a bogus matcher to force buffering exactly the right amount.
	// Before this fix, this would trigger a bug where `Read` would incorrectly
	// report `io.EOF` when only the buffer had been consumed.
	muxl.Match(func(r io.Reader) bool {
		var b [len(payload)]byte
		_, _ = r.Read(b[:])
		return false
	})
	anyl := muxl.Match(Any())
	go safeServe(errCh, muxl)
	muxedConn, err := anyl.Accept()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < mult; i++ {
		var b [len(payload)]byte
		n, err := muxedConn.Read(b[:])
		if err != nil {
			t.Error(err)
			continue
		}
		if e := len(b); n != e {
			t.Errorf("expected to read %d bytes, but read %d bytes", e, n)
		}
	}
	var b [1]byte
	if _, err := muxedConn.Read(b[:]); err != io.EOF {
		t.Errorf("unexpected error %v, expected %v", err, io.EOF)
	}
}

func TestAny(t *testing.T) {
	defer leakCheck(t)()
	errCh := make(chan error)
	defer abortOnErr(t, errCh)
	l, closer := testListener(t)
	defer closer()

	muxl := New(l)
	httpl := muxl.Match(Any())

	go runTestHTTPServer(errCh, httpl)
	go safeServe(errCh, muxl)

	runTestHTTP1Client(t, l.Addr())
}

func TestTLS(t *testing.T) {
	err := exec.Command("go", "run", build.Default.GOROOT+"/src/crypto/tls/generate_cert.go", "--host", "*").Run()
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := errors.Join(os.Remove("cert.pem"), os.Remove("key.pem")); err != nil {
			t.Error(err)
		}
	}()

	defer leakCheck(t)()
	errCh := make(chan error)
	defer abortOnErr(t, errCh)
	l, closer := testListener(t)
	defer closer()

	muxl := New(l)
	tlsl := muxl.Match(TLS())
	httpl := muxl.Match(Any())

	go runTestTLSServer(errCh, tlsl)
	go runTestHTTPServer(errCh, httpl)
	go safeServe(errCh, muxl)

	runTestHTTP1Client(t, l.Addr())
	runTestTLSClient(t, l.Addr())
}

func TestHTTP2(t *testing.T) {
	defer leakCheck(t)()
	errCh := make(chan error)
	defer abortOnErr(t, errCh)
	writer, reader := net.Pipe()
	go func() {
		if _, err := io.WriteString(writer, http2.ClientPreface); err != nil {
			panic(err)
		}
		if err := writer.Close(); err != nil {
			panic(err)
		}
	}()

	l := newChanListener()
	l.connCh <- reader
	muxl := New(l)
	// Register a bogus matcher that only reads one byte.
	muxl.Match(func(r io.Reader) bool {
		var b [1]byte
		_, _ = r.Read(b[:])
		return false
	})
	h2l := muxl.Match(HTTP2())
	go safeServe(errCh, muxl)
	muxedConn, err := h2l.Accept()
	close(l.connCh)
	if err != nil {
		t.Fatal(err)
	}
	var b [len(http2.ClientPreface)]byte
	var n int
	// We have the sniffed buffer first...
	if n, err = muxedConn.Read(b[:]); err == io.EOF {
		t.Fatal(err)
	}
	// and then we read from the source.
	if _, err = muxedConn.Read(b[n:]); err != io.EOF {
		t.Fatal(err)
	}
	if string(b[:]) != http2.ClientPreface {
		t.Errorf("got unexpected read %s, expected %s", b, http2.ClientPreface)
	}
}

func TestHTTP2MatchHeaderField(t *testing.T) {
	testHTTP2MatchHeaderField(t, HTTP2HeaderField, "value", "value", "anothervalue")
}

func TestHTTP2MatchHeaderFieldPrefix(t *testing.T) {
	testHTTP2MatchHeaderField(t, HTTP2HeaderFieldPrefix, "application/grpc+proto", "application/grpc", "application/json")
}

func testHTTP2MatchHeaderField(
	t *testing.T,
	matcherConstructor func(string, string) Matcher,
	headerValue string,
	matchValue string,
	notMatchValue string,
) {
	defer leakCheck(t)()
	errCh := make(chan error)
	defer abortOnErr(t, errCh)
	name := "name"
	writer, reader := net.Pipe()
	go func() {
		if _, err := io.WriteString(writer, http2.ClientPreface); err != nil {
			panic(err)
		}
		var buf bytes.Buffer
		enc := hpack.NewEncoder(&buf)
		if err := enc.WriteField(hpack.HeaderField{Name: name, Value: headerValue}); err != nil {
			panic(err)
		}
		framer := http2.NewFramer(writer, nil)
		err := framer.WriteHeaders(http2.HeadersFrameParam{
			StreamID:      1,
			BlockFragment: buf.Bytes(),
			EndStream:     true,
			EndHeaders:    true,
		})
		if err != nil {
			panic(err)
		}
		if err := writer.Close(); err != nil {
			panic(err)
		}
	}()

	l := newChanListener()
	l.connCh <- reader
	muxl := New(l)
	// Register a bogus matcher that only reads one byte.
	muxl.Match(func(r io.Reader) bool {
		var b [1]byte
		_, _ = r.Read(b[:])
		return false
	})
	// Create a matcher that cannot match the response.
	muxl.Match(matcherConstructor(name, notMatchValue))
	// Then match with the expected field.
	h2l := muxl.Match(matcherConstructor(name, matchValue))
	go safeServe(errCh, muxl)
	muxedConn, err := h2l.Accept()
	close(l.connCh)
	if err != nil {
		t.Fatal(err)
	}
	var b [len(http2.ClientPreface)]byte
	// We have the sniffed buffer first...
	if _, err := muxedConn.Read(b[:]); err == io.EOF {
		t.Fatal(err)
	}
	if string(b[:]) != http2.ClientPreface {
		t.Errorf("got unexpected read %s, expected %s", b, http2.ClientPreface)
	}
}

func TestHTTPGoRPC(t *testing.T) {
	defer leakCheck(t)()
	errCh := make(chan error)
	defer abortOnErr(t, errCh)
	l, closer := testListener(t)
	defer closer()

	muxl := New(l)
	httpl := muxl.Match(HTTP2(), HTTP1Fast())
	rpcl := muxl.Match(Any())

	go runTestHTTPServer(errCh, httpl)
	go runTestRPCServer(errCh, rpcl)
	go safeServe(errCh, muxl)

	runTestHTTP1Client(t, l.Addr())
	runTestRPCClient(t, l.Addr())
}

func TestMultipleMatchers(t *testing.T) {
	defer leakCheck(t)()
	errCh := make(chan error)
	defer abortOnErr(t, errCh)
	l, closer := testListener(t)
	defer closer()

	muxl := New(l)
	unmatcher := func(r io.Reader) bool { return false }

	go safeServe(errCh, muxl)
	go runTestHTTPServer(errCh, muxl.Match(unmatcher, func(r io.Reader) bool {
		return true
	}, unmatcher))
	runTestHTTP1Client(t, l.Addr())
}

func TestListenerClose(t *testing.T) {
	defer leakCheck(t)()
	errCh := make(chan error)
	defer abortOnErr(t, errCh)
	l := newChanListener()

	c1, c2 := net.Pipe()
	muxl := New(l)
	anyl := muxl.Match(Any())

	go safeServe(errCh, muxl)

	l.connCh <- c1
	l.connCh <- c2

	// First connection goes through.
	if _, err := anyl.Accept(); err != nil {
		t.Fatal(err)
	}

	// Listener is closed.
	close(l.connCh)

	// Second connection either goes through or it is closed.
	if _, err := anyl.Accept(); err != nil {
		if err != ErrServerClosed {
			t.Fatal(err)
		}
		c2.Close()
		// The error is either io.ErrClosedPipe or net.OpError wrapping
		// a net.pipeError depending on the go version.
		if _, err := c2.Read([]byte{}); !strings.Contains(err.Error(), "closed") {
			t.Fatalf("connection is not closed and is leaked: %v", err)
		}
	}
}

func TestClose(t *testing.T) {
	defer leakCheck(t)()
	errCh := make(chan error)
	defer abortOnErr(t, errCh)
	l, closer := testListener(t)
	defer closer()

	muxl := New(l)
	anyl := muxl.Match(Any())

	go safeServe(errCh, muxl)

	muxl.Close()
	if _, err := anyl.Accept(); err != ErrServerClosed {
		t.Fatal(err)
	}
}

// Cribbed from google.golang.org/grpc/test/end2end_test.go.

// interestingGoroutines returns all goroutines we care about for the purpose
// of leak checking. It excludes testing or runtime ones.
func interestingGoroutines() (gs []string) {
	buf := make([]byte, 2<<20)
	buf = buf[:runtime.Stack(buf, true)]
	for _, g := range strings.Split(string(buf), "\n\n") {
		sl := strings.SplitN(g, "\n", 2)
		if len(sl) != 2 {
			continue
		}
		stack := strings.TrimSpace(sl[1])
		if strings.HasPrefix(stack, "testing.RunTests") {
			continue
		}

		if stack == "" ||
			strings.Contains(stack, "main.main()") ||
			strings.Contains(stack, "testing.Main(") ||
			strings.Contains(stack, "runtime.goexit") ||
			strings.Contains(stack, "created by runtime.gc") ||
			strings.Contains(stack, "interestingGoroutines") ||
			strings.Contains(stack, "runtime.MHeap_Scavenger") {
			continue
		}
		gs = append(gs, g)
	}
	sort.Strings(gs)
	return
}

// leakCheck snapshots the currently-running goroutines and returns a
// function to be run at the end of tests to see whether any
// goroutines leaked.
func leakCheck(t testing.TB) func() {
	orig := map[string]bool{}
	for _, g := range interestingGoroutines() {
		orig[g] = true
	}
	return func() {
		// Loop, waiting for goroutines to shut down.
		// Wait up to 5 seconds, but finish as quickly as possible.
		deadline := time.Now().Add(5 * time.Second)
		for {
			var leaked []string
			for _, g := range interestingGoroutines() {
				if !orig[g] {
					leaked = append(leaked, g)
				}
			}
			if len(leaked) == 0 {
				return
			}
			if time.Now().Before(deadline) {
				time.Sleep(50 * time.Millisecond)
				continue
			}
			for _, g := range leaked {
				t.Errorf("Leaked goroutine: %v", g)
			}
			return
		}
	}
}
