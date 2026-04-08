// internal/api/merchant/stat.go
package merchant

import (
	"github.com/example/epay-go/internal/middleware"
	"github.com/example/epay-go/internal/service"
	"github.com/example/epay-go/pkg/response"
	"github.com/gin-gonic/gin"
)

// Dashboard 商户仪表盘
func Dashboard(c *gin.Context) {
	merchantID := middleware.GetUserID(c)

	merchantService := service.NewMerchantService()
	orderService := service.NewOrderService()

	// 获取商户信息
	merchant, err := merchantService.GetByID(merchantID)
	if err != nil {
		response.NotFound(c, "商户不存在")
		return
	}

	// 今日统计
	todayCount, todayAmount, err := orderService.GetTodayStats(&merchantID)
	if err != nil {
		response.ServerError(c, "获取统计数据失败")
		return
	}

	response.Success(c, gin.H{
		"balance":            merchant.Balance,
		"frozen_balance":     merchant.FrozenBalance,
		"today_order_count":  todayCount,
		"today_order_amount": todayAmount,
	})
}
