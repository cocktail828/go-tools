package {{.PkgName}}

import (
	"context"
	"time"
	"net/http"

	"github.com/cocktail828/go-tools/xlog"
	"github.com/gin-gonic/gin"
	{{.ImportPackages}}
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
		
		l := {{.LogicName}}.New{{.LogicType}}(ctx, log)
		{{if .HasResp}}resp, {{end}}err := l.{{.Call}}({{if .HasRequest}}&req{{end}})
		if err != nil {
			c.AbortWithError(http.StatusBadGateway, err)
		} else {
			{{if .HasResp}}c.AbortWithStatusJSON(http.StatusOK, resp){{else}}c.AbortWithStatus(http.StatusOK){{end}}
		}
	}
}
