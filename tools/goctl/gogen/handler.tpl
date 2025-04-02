package {{.pkgName}}

import (
	"context"
	"time"
	"net/http"
	"sync"

	"github.com/cocktail828/go-tools/xlog"
	"github.com/gin-gonic/gin"
	{{.imports}}
)

{{if .doc}}{{.doc}}{{end}}// Meta stores global application metadata and shared resources
func {{.handler}}Handler(tmo time.Duration, log xlog.Logger, m *sync.Map) gin.HandlerFunc {
	return func(c *gin.Context) {
		{{if .requestType}}var req model.{{.requestType}}
		if err := c.ShouldBind(&req); err != nil {
			c.AbortWithError(http.StatusBadGateway, err)
			return
		}

		{{end}}ctx, cancel := context.WithTimeout(c.Request.Context(), tmo)
		defer cancel()

		{{if .responseType}}resp, {{end}}err := process{{.handler}}(ctx{{if .requestType}}, &req{{end}})
		if err != nil {
			c.AbortWithError(http.StatusBadGateway, err)
		} else {
			{{if .responseType}}c.AbortWithStatusJSON(http.StatusOK, resp){{else}}c.AbortWithStatus(http.StatusOK){{end}}
		}
	}
}

func process{{.handler}}(ctx context.Context{{if .requestType}}, req *model.{{.requestType}}{{end}}) ({{if .responseType}}resp {{.responseType}}, {{end}}err error) {
	// TODO: add your logic here and delete this line

	return
}
