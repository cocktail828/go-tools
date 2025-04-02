package {{.pkgName}}

import (
    "time"
	"net/http"
	"sync"
	
	"github.com/cocktail828/go-tools/xlog"
	"github.com/gin-gonic/gin"
	{{.imports}}
)

type Meta struct {
	xlog.Logger
	Timeout      time.Duration

	// Meta stores global application metadata and shared resources
	Meta         *sync.Map
}

func RegisterHandlers(g *gin.Engine, m *Meta) {
	{{if .middleware}}g.Use({{.middleware}}){{end}}

	{{.routes}}
}