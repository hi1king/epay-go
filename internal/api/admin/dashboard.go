// internal/handler/admin/dashboard.go
package admin

import (
	"github.com/example/epay-go/internal/service"
	"github.com/example/epay-go/pkg/response"
	"github.com/gin-gonic/gin"
)

// Dashboard 仪表盘数据
func Dashboard(c *gin.Context) {
	orderService := service.NewOrderService()
	merchantService := service.NewMerchantService()

	// 今日统计
	todayCount, todayAmount, err := orderService.GetTodayStats(nil)
	if err != nil {
		response.ServerError(c, "获取统计数据失败")
		return
	}

	// 商户数量
	merchants, totalMerchants, _ := merchantService.List(1, 1, nil)
	_ = merchants

	response.Success(c, gin.H{
		"today_order_count":  todayCount,
		"today_order_amount": todayAmount,
		"total_merchants":    totalMerchants,
	})
}
