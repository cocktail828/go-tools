package main

import (
    "context"
    "log"
    "net/http"
    "os/signal"
    "syscall"

    "github.com/gin-gonic/gin"
	"github.com/cocktail828/go-tools/configor"
	"github.com/cocktail828/go-tools/pkg/nacs"
	"github.com/cocktail828/go-tools/pkg/nacs/regular"

	"{{ .project}}/config"
    "{{ .project}}/handler"
    {{ if .service.HasInterceptor }}"{{ .project}}/interceptor"{{ end }}
)

func main() {
	/* load config from file */
	log.Printf("about to load config")
	cfgor, err := regular.NewFileConfigor( /* the config dir */ "xxx.toml")
	if err != nil {
		log.Fatalf("load config fail: %v", err)
	}

	payload, err := cfgor.Load(nacs.Config{ID: "server.toml"})
	if err != nil {
		log.Fatalf("regular.Load() config fail: %v", err)
	}

	cfg := config.Config{}
	if err := configor.Load(&cfg, payload); err != nil {
		log.Fatalf("parser config fail: %v", err)
	}

    /* manually change to other mode gin.DebugMode, gin.TestMode, gin.ReleaseMode */
    gin.SetMode(gin.DebugMode)
    r := gin.Default()
    {{ if .service.Interceptors }}
    /* register global interceptors */
    r.Use({{ range .service.Interceptors }}
        interceptor.{{ . }}( /* init via meta */ nil),{{ end }}
    ){{ end }}
    {{ range $group := .service.Groups }}{
        /* register group with group level interceptors */
        {{ if .Interceptors }}{{ .Name }}Group := r.Group("{{ .Prefix }}", {{ range .Interceptors }}
            interceptor.{{ . }}( /* init via meta */ nil),{{ end }}
        ){{ else }}{{ .Name }}Group := r.Group("{{ .Prefix }}"){{ end }}
        {{ range .Routes }}
        {{ $group.Name }}Group.Handle(http.Method{{ .Method }}, "{{ .Path }}", handler.{{ .HandlerName }}){{ end }}
    }{{ end }}

    /* demo server start at ':8080' */
    srv := http.Server{
        Addr:    ":8080",
        Handler: r,
    }

    log.Printf("about to start server at %q", srv.Addr)
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    go func() {
        defer stop()
        srv.ListenAndServe()
    }()

    <-ctx.Done() // wait server exit or a quit signal
    log.Println("server exit with error: ", srv.Shutdown(context.Background()))
}