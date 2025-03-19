package main

import (
	"time"

	"github.com/cocktail828/go-tools/xlog/colorful"
	"github.com/cocktail828/go-tools/xlog"
	"github.com/cocktail828/go-tools/z"

	{{.imports}}
)

func main() {
	cfg := service.Config{
		Addr:    ":8080", // random port
		Timeout: time.Second,
	}

	colorful.Infof("Starting server at %s...", cfg.Addr)
	z.Must(service.Run(cfg, xlog.NoopLogger{}))
}
