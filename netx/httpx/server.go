package httpx

import (
	"context"
	"errors"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server interface {
	Close() error
	ListenAndServe() error
	ListenAndServeTLS(certFile string, keyFile string) error
	RegisterOnShutdown(f func())
	Serve(l net.Listener) error
	ServeTLS(l net.Listener, certFile string, keyFile string) error
	SetKeepAlivesEnabled(v bool)
	Shutdown(ctx context.Context) error
}

var _ Server = &GracefulServer{}

type GracefulServer struct {
	Server  // canonical go http server
	Signals []os.Signal
	Timeout time.Duration // graceful time
}

func (srv *GracefulServer) normalize() {
	if len(srv.Signals) == 0 {
		srv.Signals = append(srv.Signals, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	}

	if srv.Timeout == 0 {
		srv.Timeout = time.Second * 3
	}
}

func (srv *GracefulServer) ListenAndServe() error {
	srv.normalize()
	var srverr error
	ctx, cancel := signal.NotifyContext(context.Background(), srv.Signals...)
	go func() {
		defer cancel()
		srverr = srv.Server.ListenAndServe()
	}()
	<-ctx.Done()

	tmoctx, _ := context.WithTimeout(context.Background(), srv.Timeout)
	return errors.Join(srverr, srv.Server.Shutdown(tmoctx))
}
