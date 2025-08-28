package main

import (
    "context"
    "log"
    "net/http"
    "os/signal"
    "syscall"

    "github.com/gin-gonic/gin"

    "{{ .project}}/handler"
    {{ if .has_interceptor }}"{{ .project}}/interceptor"{{ end }}
)

func main() {
    r := gin.Default()
    r.Use({{ range .interceptors }}
        interceptor.{{ . }}Incp(/* init meta */ nil),{{ end }}
    )

    {{ range $grp := .service.Groups }}{
        {{ .Prefix }}Group := r.Group("{{ .Prefix }}"{{ if $grp.Interceptors }}, {{ end }}{{ range $grp.Interceptors }}
            interceptor.{{ . }}Incp(/* init meta */ nil),{{ end }}
        )
        {{ range .Routes }}
        {{ $grp.Prefix }}Group.Handle(http.Method{{ .Method }}, "{{ .Path }}", handler.{{ .Request }}Handler){{ end }}
    }{{ end }}

    // demo server start at ':8080'
    srv := http.Server{Addr: ":8080"}
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    go func() {
        defer stop()
        srv.ListenAndServe()
    }()

    <-ctx.Done() // wait server exit or a quit signal
    log.Println("server exit with error: ", srv.Shutdown(context.Background()))
}