package {{.PkgName}}

import (
	"context"
	"testing"
	"time"
	
	{{if .HasResponse}}
	"github.com/stretchr/testify/assert"{{end}}
	"github.com/stretchr/testify/require"
	{{.imports}}
)
{{if .HasDoc}}{{.Doc}}
{{end}}
func Test{{.HandlerName}}Handler(t *testing.T) {
	tmo := time.Second       // TODO: alter tmo as expect

	tests := []struct {
		name       string
		{{if .HasRequest}}req    model.{{.RequestType}}
		{{end}}{{if .HasResponse}}expect   {{.ResponseType}}{{end}}
	}{
		{
			name: "handler error",
			{{if .HasRequest}}// TODO: add argument here
			{{end}}{{if .HasResponse}}// TODO: add expect result here
		{{end}}},
		{
			name: "handler successful",
			{{if .HasRequest}}// TODO: add argument here
			{{end}}{{if .HasResponse}}// TODO: add expect result here
		{{end}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tmo)
			{{if .HasResponse}}ret, {{end}}err := handle{{.HandlerName}}(ctx{{if .HasRequest}}, &tt.req{{end}})
			cancel()
			require.NoError(t, err){{if .HasResponse}}
			assert.Equal(t, tt.expect, ret){{end}}
		})
	}
}
