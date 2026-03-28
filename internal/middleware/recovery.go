// internal/middleware/recovery.go
package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/example/epay-go/pkg/response"
	"github.com/gin-gonic/gin"
)

// Recovery panic 恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC] %v\n%s", err, debug.Stack())
				c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
					Code: response.CodeServerError,
					Msg:  "服务器内部错误",
				})
			}
		}()
		c.Next()
	}
}
