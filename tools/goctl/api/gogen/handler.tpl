package {{.PkgName}}

import (
	"context"
	"time"
	"net/http"

	"github.com/cocktail828/go-tools/xlog"
	"github.com/gin-gonic/gin"
	{{.imports}}
)

{{if .HasDoc}}{{.Doc}}{{end}}func {{.HandlerName}}Handler(tmo time.Duration, log xlog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		{{if .HasRequest}}var req model.{{.RequestType}}
		if err := c.ShouldBind(&req); err != nil {
			c.AbortWithError(http.StatusBadGateway, err)
			return
		}

		{{end}}ctx, cancel := context.WithTimeout(c.Request.Context(), tmo)
		defer cancel()

		{{if .HasResponse}}resp, {{end}}err := handle{{.HandlerName}}(ctx{{if .HasRequest}}, &req{{end}})
		if err != nil {
			c.AbortWithError(http.StatusBadGateway, err)
		} else {
			{{if .HasResponse}}c.AbortWithStatusJSON(http.StatusOK, resp){{else}}c.AbortWithStatus(http.StatusOK){{end}}
		}
	}
}

func handle{{.HandlerName}}(ctx context.Context{{if .HasRequest}}, req *model.{{.RequestType}}{{end}}) ({{if .HasResponse}}resp {{.ResponseType}}, {{end}}err error) {
	// TODO: add your logic here and delete this line

	return
}
