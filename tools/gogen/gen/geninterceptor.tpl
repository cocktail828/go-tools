package interceptor

import (
	"github.com/gin-gonic/gin"
)

func {{ .name }}Incp(in any) gin.HandlerFunc {
	/* add some init code here */

	return func(*gin.Context) {
	}
}
