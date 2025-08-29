package interceptor

import (
    "github.com/gin-gonic/gin"
)

func {{ .name }}(in any) gin.HandlerFunc {
    /* add some init code here */

    return func(*gin.Context) {
        /* add some business code here */
    }
}
