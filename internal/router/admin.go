package router

import (
	"github.com/example/epay-go/internal/api/admin"
	"github.com/example/epay-go/internal/middleware"
	"github.com/example/epay-go/pkg/jwt"
	"github.com/gin-gonic/gin"
)

func registerAdminRoutes(r *gin.Engine) {
	adminAPI := r.Group("/api/admin")
	{
		adminAPI.POST("/auth/login", admin.Login)

		adminAuth := adminAPI.Group("")
		adminAuth.Use(middleware.JWTAuth(jwt.TokenTypeAdmin))
		{
			adminAuth.POST("/auth/logout", admin.Logout)
			adminAuth.PUT("/profile/password", admin.UpdatePassword)
			adminAuth.GET("/dashboard", admin.Dashboard)

			adminAuth.GET("/merchants", admin.ListMerchants)
			adminAuth.GET("/merchants/:id", admin.GetMerchant)
			adminAuth.PUT("/merchants/:id", admin.UpdateMerchant)
			adminAuth.PATCH("/merchants/:id/status", admin.UpdateMerchantStatus)

			adminAuth.GET("/orders", admin.ListOrders)
			adminAuth.GET("/orders/:trade_no", admin.GetOrder)
			adminAuth.POST("/orders/:trade_no/renotify", admin.RenotifyOrder)

			adminAuth.GET("/channels", admin.ListChannels)
			adminAuth.POST("/channels", admin.CreateChannel)
			adminAuth.GET("/channels/:id", admin.GetChannel)
			adminAuth.PUT("/channels/:id", admin.UpdateChannel)
			adminAuth.DELETE("/channels/:id", admin.DeleteChannel)

			adminAuth.GET("/plugins", admin.GetAllPluginConfigs)
			adminAuth.GET("/plugins/:plugin/config", admin.GetPluginConfig)

			adminAuth.POST("/test-payment", admin.TestPayment)

			adminAuth.GET("/settlements", admin.ListSettlements)
			adminAuth.PATCH("/settlements/:id/approve", admin.ApproveSettlement)
			adminAuth.PATCH("/settlements/:id/reject", admin.RejectSettlement)

			adminAuth.POST("/refunds", admin.CreateRefund)
			adminAuth.GET("/refunds", admin.ListRefunds)
			adminAuth.POST("/refunds/:refund_no/process", admin.ProcessRefund)
		}
	}
}
