package router

import (
	pay "github.com/example/epay-go/internal/api/pay"
	"github.com/gin-gonic/gin"
)

func registerWebhookRoutes(r *gin.Engine) {
	webhookAPI := r.Group("/api/pay")
	{
		webhookAPI.POST("/notify/:channel", pay.HandleNotify)
		webhookAPI.GET("/return/:channel", pay.HandleReturn)
	}
}
