// Code generated by goctl. DO NOT EDIT.
// goctl {{.version}}

package handler

import (
    "time"
	"net/http"
	
	"github.com/cocktail828/go-tools/xlog"
	"github.com/gin-gonic/gin"
	{{.imports}}
)

type Meta struct {
	xlog.Logger
	Timeout      time.Duration
	Interceptors []gin.HandlerFunc
}

func RegisterHandlers(g *gin.Engine, meta Meta) {
	g.Use(meta.Interceptors...)

	{{.routes}}
}