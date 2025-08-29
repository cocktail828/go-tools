package handler

import ({{ if  .route.Response }}
    "net/http"{{ end }}{{ if .route.Request }}
    "github.com/pkg/errors"{{ end }}
    "github.com/gin-gonic/gin"
    {{ if or (len .route.Request) (len .route.Response) }}
    "{{ .project}}/model"{{ end }}
)

func {{ .route.HandlerName }}(c *gin.Context) { {{ if .route.Request }}
    req := model.{{ .route.Request }}{}
    if err := c.ShouldBindBodyWithJSON(&req); err != nil {
        c.AbortWithError(http.StatusBadRequest, errors.Errorf("ShouldBindBodyWithJSON fail for %v", err))
        return
    }{{ end }}
    {{ if .route.Response }}resp := {{ end }}__{{ .route.HandlerName }}({{ if .route.Request }}&req{{ end }})
    {{ if .route.Response }}c.AbortWithStatusJSON(http.StatusOK, resp){{ end }}
}

func __{{ .route.HandlerName }}({{ if .route.Request }}req *model.{{ .route.Request }}{{ end }}) {{ if .route.Response }}*model.{{ .route.Response }}{{ end }} {
    /* add some business code here */

    return{{ if .route.Response }} nil {{ end }}
}