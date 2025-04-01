package {{.pkgName}}

import (
	"context"
	"testing"
	"time"
	
	{{if .responseType}}
	"github.com/stretchr/testify/assert"{{end}}
	"github.com/stretchr/testify/require"
	{{.imports}}
)

{{if .doc}}{{.doc}}{{end}}func Test_{{.handler}}Handler(t *testing.T) {
	tmo := time.Second       // TODO: alter tmo as expect

	tests := []struct {
		name       string
		{{if .requestType}}req    model.{{.requestType}}
		{{end}}{{if .responseType}}expect   {{.responseType}}{{end}}
	}{
		{
			name: "handler error",
			{{if .requestType}}// TODO: add argument here
			{{end}}{{if .responseType}}// TODO: add expect result here
		{{end}}},
		{
			name: "handler successful",
			{{if .requestType}}// TODO: add argument here
			{{end}}{{if .responseType}}// TODO: add expect result here
		{{end}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tmo)
			{{if .responseType}}ret, {{end}}err := process{{.handler}}(ctx{{if .requestType}}, &tt.req{{end}})
			cancel()
			require.NoError(t, err){{if .responseType}}
			assert.Equal(t, tt.expect, ret){{end}}
		})
	}
}
