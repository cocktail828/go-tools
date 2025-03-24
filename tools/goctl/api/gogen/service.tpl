package service

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/cocktail828/go-tools/xlog"
	"github.com/cocktail828/go-tools/xlog/colorful"
	{{.imports}}
)

type Config struct {
	Addr    string        `toml:"addr"`
	Timeout time.Duration `toml:"timeout"`
}

func Run(cfg Config, log xlog.Logger) {
	gin.SetMode(gin.ReleaseMode)
	g := gin.Default()

	handler.RegisterHandlers(g, &handler.Meta{
		Logger:       log,
		Timeout:      cfg.Timeout,
		Interceptors: {{.middlewares}},
	})

	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: g.Handler(),
	}

	finishChan := make(chan struct{})
	sigctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	defer cancel()
	go func() {
		defer close(finishChan)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			colorful.Error(err)
		}
	}()

	<-sigctx.Done()              // wait for signal...
	srv.Shutdown(context.TODO()) // graceful shutdown
	<-finishChan
}
