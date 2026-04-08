package router

import (
	"github.com/example/epay-go/internal/api/merchant"
	"github.com/example/epay-go/internal/middleware"
	"github.com/example/epay-go/pkg/jwt"
	"github.com/gin-gonic/gin"
)

func registerMerchantRoutes(r *gin.Engine) {
	merchantAPI := r.Group("/api/merchant")
	{
		merchantAPI.POST("/auth/login", merchant.Login)
		merchantAPI.POST("/auth/register", merchant.Register)

		merchantAuth := merchantAPI.Group("")
		merchantAuth.Use(middleware.JWTAuth(jwt.TokenTypeMerchant))
		{
			merchantAuth.POST("/auth/logout", merchant.Logout)
			merchantAuth.GET("/dashboard", merchant.Dashboard)

			merchantAuth.GET("/profile", merchant.GetProfile)
			merchantAuth.PUT("/profile", merchant.UpdateProfile)
			merchantAuth.PUT("/profile/password", merchant.UpdatePassword)
			merchantAuth.POST("/profile/reset-key", merchant.ResetAPIKey)

			merchantAuth.GET("/orders", merchant.ListOrders)
			merchantAuth.GET("/orders/:trade_no", merchant.GetOrder)

			merchantAuth.GET("/withdrawals", merchant.ListWithdraws)
			merchantAuth.POST("/withdrawals", merchant.ApplyWithdraw)

			merchantAuth.GET("/balance-logs", merchant.ListBalanceLogs)

			merchantAuth.POST("/refunds", merchant.CreateRefund)
			merchantAuth.GET("/refunds", merchant.ListRefunds)

			merchantAuth.POST("/test-payment", merchant.TestPayment)
		}
	}
}
