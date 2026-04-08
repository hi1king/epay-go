package router

import "github.com/gin-gonic/gin"

// Setup 设置所有路由
func Setup(r *gin.Engine) {
	registerPayRoutes(r)
	registerWebhookRoutes(r)
	registerAdminRoutes(r)
	registerMerchantRoutes(r)
}
