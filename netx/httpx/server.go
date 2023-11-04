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

func (s *Server) WaitForSignal(timeout time.Duration, sigs ...os.Signal) (os.Signal, error) {
	sigChan := make(chan os.Signal, 5)
	signal.Notify(sigChan, sigs...)
	defer signal.Stop(sigChan)

	sig := <-sigChan
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return sig, ((*http.Server)(s)).Shutdown(ctx)
}
