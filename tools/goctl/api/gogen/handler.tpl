package {{.PkgName}}

import (
	"context"
	"time"
	"net/http"

	"github.com/cocktail828/go-tools/xlog"
	"github.com/gin-gonic/gin"
	{{.imports}}
)

{{if .HasDoc}}{{.Doc}}{{end}}
func {{.HandlerName}}(tmo time.Duration, log xlog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		{{if .HasRequest}}var req model.{{.RequestType}}
		if err := c.ShouldBind(&req); err != nil {
			c.AbortWithError(http.StatusBadGateway, err)
			return
		}

		{{end}}ctx, cancel := context.WithTimeout(c.Request.Context(), tmo)
		defer cancel()

		{{if .HasResp}}resp, {{end}}err := {{.HandlerName}}Handler(ctx, {{if .HasRequest}}&req{{end}})
		if err != nil {
			c.AbortWithError(http.StatusBadGateway, err)
		} else {
			{{if .HasResp}}c.AbortWithStatusJSON(http.StatusOK, resp){{else}}c.AbortWithStatus(http.StatusOK){{end}}
		}
	}
}

func {{.HandlerName}}Handler(ctx context.Context, {{.Request}}) {{.ResponseType}} {
	// TODO: add your logic here and delete this line

	{{.ReturnString}}
}
