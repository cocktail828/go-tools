package httpx

import (
	"context"
	"errors"
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

type Server interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

type GracefulServer struct {
	Server                // canonical go http server
	Signals []os.Signal   // should not be nil
	Timeout time.Duration // graceful time
}

func (srv GracefulServer) ListenAndServe() error {
	var srverr error
	signals := srv.Signals
	if len(srv.Signals) == 0 {
		signals = DefaultSignals
	}

	timeout := srv.Timeout
	if srv.Timeout == 0 {
		timeout = time.Second * 3
	}

	ctx, cancel := signal.NotifyContext(context.Background(), signals...)
	go func() {
		defer cancel()
		srverr = srv.Server.ListenAndServe()
	}()
	<-ctx.Done()

	tmoctx, _ := context.WithTimeout(context.Background(), timeout)
	return errors.Join(srverr, srv.Server.Shutdown(tmoctx))
}

// 针对随机端口做处理
type GoHTTPServer struct {
	http.Server
	listener net.Listener
}

func (srv *GoHTTPServer) Port() int {
	if srv.listener == nil {
		return 0
	}
	return srv.listener.Addr().(*net.TCPAddr).Port
}

func (srv *GoHTTPServer) ListenAndServe() error {
	addr := srv.Server.Addr
	if addr == "" {
		addr = ":http"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	srv.listener = ln
	return srv.Server.Serve(ln)
}
