package service

import (
	"context"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/cocktail828/go-tools/xlog"
	"github.com/cocktail828/go-tools/xlog/colorful"
	"github.com/cocktail828/go-tools/z/runnable"
	{{.imports}}
)

type Config struct {
	Addr    string        `toml:"addr"`
	Timeout time.Duration `toml:"timeout"`
}

func Run(cfg Config, log xlog.Logger) {
	gin.SetMode(gin.ReleaseMode)
	g := gin.Default()

	m := &{{.route}}.Meta{
		Logger:       log,
		Timeout:      cfg.Timeout,

		// Meta stores global application metadata and shared resources
		Meta:         &sync.Map{},
	}
	{{.route}}.RegisterHandlers(g, m)

	sigctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	defer cancel()

	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: g.Handler(),
	}

	gs := runnable.Graceful{
		Start: func() error {
			err := srv.ListenAndServe()
			if err != http.ErrServerClosed {
				colorful.Error(err)
			}
			return err
		},
		Stop: func() error {
			return srv.Shutdown(context.TODO())
		},
	}
	gs.Launch(sigctx)
}
