package handler

import (
	"github.com/gin-gonic/gin"
)

func {{ .name }}Handler(in any) gin.HandlerFunc {
	/* add some init code here */

	return func(*gin.Context) {
	}
}
