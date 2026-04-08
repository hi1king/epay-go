// internal/api/pay/debug_notify.go
package payment

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// TestNotify 用于测试单的商户通知地址，避免通知重试刷日志
func TestNotify(c *gin.Context) {
	c.String(http.StatusOK, "success")
}
