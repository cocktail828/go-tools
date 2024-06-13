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
	srvCtx      context.Context
	srvCancel   context.CancelFunc // called on server exit, should not be nil
	listener    net.Listener
}

func (srv *Server) normalize() { srv.srvCtx, srv.srvCancel = context.WithCancel(context.Background()) }

// get the net.Addr the server used
func (srv *Server) Port() net.Addr {
	srv.normalize()
	if srv.listener == nil {
		return nil
	}
	return srv.listener.Addr()
}

func (srv *Server) ListenAndServe() error {
	srv.normalize()
	defer srv.srvCancel()

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
	srv.normalize()
	defer srv.srvCancel()

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
	srv.normalize()
	defer srv.srvCancel()

	srv.listener = ln
	return srv.Server.Serve(ln)
}

func (srv *Server) ServeTLS(ln net.Listener, certFile string, keyFile string) error {
	srv.normalize()
	defer srv.srvCancel()

	srv.listener = ln
	return srv.Server.ServeTLS(ln, certFile, keyFile)
}

type Registry interface {
	Register()
	DeRegister()
}

// block until server quit, then wait gracelful time and exit
func (srv *Server) WaitForSignal(r Registry, tmo time.Duration, signals ...os.Signal) {
	srv.normalize()
	if len(signals) == 0 {
		signals = DefaultSignals
	}

	select {
	case <-srv.srvCtx.Done(): // already quit
		return
	default:
	}

	// register service
	if r != nil {
		r.Register()
	}

	// wait for signals or server quit
	sigctx, _ := signal.NotifyContext(srv.srvCtx, signals...)
	<-sigctx.Done()

	// deregister service
	if r != nil {
		r.DeRegister()
	}

	select {
	case <-time.After(tmo): // graceful time
		srv.Server.Shutdown(context.Background())
	case <-srv.srvCtx.Done(): // already quit, no need to be graceful
	}
}
