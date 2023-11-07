package httpx

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type Server http.Server

func (s *Server) ListenAndServe() error {
	return ((*http.Server)(s)).ListenAndServe()
}

func (s *Server) WaitForSignal(timeout time.Duration, sigs ...os.Signal) error {
	func() {
		ctx, cancel := signal.NotifyContext(context.Background(), sigs...)
		defer cancel()
		<-ctx.Done()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return ((*http.Server)(s)).Shutdown(ctx)
}
