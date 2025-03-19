package {{.PkgName}}

import (
	"bytes"
	"time"
	{{if .HasRequest}}"encoding/json"{{end}}
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cocktail828/go-tools/xlog"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	{{.imports}}
)

{{if .HasDoc}}{{.Doc}}{{end}}
func Test{{.HandlerName}}(t *testing.T) {
	log := xlog.NoopLogger{} // mock
	tmo := time.Second       // mock

	tests := []struct {
		name       string
		reqBody    interface{}
		wantStatus int
		wantResp   string
		setupMocks func()
	}{
		{
			name:    "invalid request body",
			reqBody: "invalid",
			wantStatus: http.StatusBadRequest,
			wantResp:   "unsupported type", // Adjust based on actual error response
			setupMocks: func() {
				// No setup needed for this test case
			},
		},
		{
			name: "handler error",
			{{if .HasRequest}}reqBody: model.{{.RequestType}}{
				//TODO: add fields here
			},
			{{end}}wantStatus: http.StatusBadRequest,
			wantResp:  "error", // Adjust based on actual error response
			setupMocks: func() {
				// Mock login logic to return an error
			},
		},
		{
			name: "handler successful",
			{{if .HasRequest}}reqBody: model.{{.RequestType}}{
				//TODO: add fields here
			},
			{{end}}wantStatus: http.StatusOK,
			wantResp:   `{"code":0,"msg":"success","data":{}}`, // Adjust based on actual success response
			setupMocks: func() {
				// Mock login logic to return success
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			var reqBody []byte
			{{if .HasRequest}}var err error
			reqBody, err = json.Marshal(tt.reqBody)
			require.NoError(t, err){{end}}
			req, err := http.NewRequest("POST", "/ut", bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rr)
			c.Request = req
			{{.HandlerName}}(tmo, log)(c)
			 
			t.Log(rr.Body.String())
			assert.Equal(t, tt.wantStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.wantResp)
		})
	}
}
