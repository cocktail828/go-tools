package httpx

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	DefaultSignals = []os.Signal{syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT}
)

type Server struct {
	http.Server // canonical go http server
	Context     context.Context
	Cancel      context.CancelFunc // called on server exit, should not be nil
	listener    net.Listener
}

// go will server on a random port if Addr is ":0", we can get the port via listener
func (srv *Server) Port() net.Addr {
	if srv.listener == nil {
		return nil
	}
	return srv.listener.Addr()
}

func (srv *Server) ListenAndServe() error {
	defer srv.Cancel()

	addr := srv.Server.Addr
	if addr == "" {
		addr = ":http"
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	defer ln.Close()
	srv.listener = ln
	return srv.Server.Serve(srv.listener)
}

func (srv *Server) ListenAndServeTLS(certFile string, keyFile string) error {
	defer srv.Cancel()

	addr := srv.Addr
	if addr == "" {
		addr = ":https"
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	defer ln.Close()
	srv.listener = ln
	return srv.Server.ServeTLS(ln, certFile, keyFile)
}

func (srv *Server) Serve(ln net.Listener) error {
	defer srv.Cancel()

	srv.listener = ln
	return srv.Server.Serve(ln)
}

func (srv *Server) ServeTLS(ln net.Listener, certFile string, keyFile string) error {
	defer srv.Cancel()

	srv.listener = ln
	return srv.Server.ServeTLS(ln, certFile, keyFile)
}

type Registry interface {
	Register()
	DeRegister()
}

// block until server quit, then wait gracelful time and exit
func (srv *Server) WaitForSignal(r Registry, tmo time.Duration, signals ...os.Signal) {
	if len(signals) == 0 {
		signals = DefaultSignals
	}

	// register service
	if r != nil {
		r.Register()
	}

	// wait for signals or server quit
	sigctx, _ := signal.NotifyContext(srv.Context, signals...)
	<-sigctx.Done()

	// deregister service
	if r != nil {
		r.DeRegister()
	}

	select {
	case <-time.After(tmo): // graceful time
		srv.Server.Shutdown(context.Background())
	case <-srv.Context.Done(): // already quit, no need to be graceful
	}
}
