package service

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/cocktail828/go-tools/xlog"
	{{.imports}}
)

type Config struct {
	Addr    string `toml:"addr"`
	Timeout time.Duration
}

func Run(cfg Config, log xlog.Logger) error {
	gin.SetMode(gin.ReleaseMode)
	g := gin.Default()

	handler.RegisterHandlers(g, handler.Meta{
		Logger:       log,
		Timeout:      cfg.Timeout,
		Interceptors: {{.middlewares}},
	})

	return g.Run(cfg.Addr)
}
