package router

import (
	pay "github.com/example/epay-go/internal/api/pay"
	"github.com/gin-gonic/gin"
)

func registerPayRoutes(r *gin.Engine) {
	// 兼容旧版易支付接口
	r.GET("/submit.php", pay.LegacySubmit)
	r.POST("/submit.php", pay.LegacySubmit)
	r.POST("/mapi.php", pay.LegacyCreateOrder)
	r.GET("/api.php", pay.LegacyAPI)

	payAPI := r.Group("/api/pay")
	{
		payAPI.POST("/create", pay.CreateOrder)
		payAPI.GET("/query", pay.QueryOrder)
		payAPI.GET("/status/:trade_no", pay.PublicOrderStatus)
		payAPI.GET("/test-notify", pay.TestNotify)
	}
}
