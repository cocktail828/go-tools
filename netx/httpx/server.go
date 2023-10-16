package httpx

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"
)

type GracefulServer struct {
	mu      sync.Mutex
	server  *http.Server
	sigChan chan os.Signal
}

func (gs *GracefulServer) init(srv *http.Server) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	if gs.sigChan == nil {
		gs.sigChan = make(chan os.Signal, 1)
		signal.Notify(gs.sigChan, syscall.SIGTERM, syscall.SIGINT)
	}
	gs.server = srv
}

func (gs *GracefulServer) ListenAndServe(srv *http.Server) error {
	gs.init(srv)
	defer signal.Stop(gs.sigChan)
	if err := gs.server.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

func (gs *GracefulServer) WaitForSignal(timeout time.Duration) (os.Signal, error) {
	if gs.sigChan == nil || gs.server == nil {
		return nil, errors.Errorf("call ListenAndServe first")
	}

	sig := <-gs.sigChan
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return sig, gs.server.Shutdown(ctx)
}
