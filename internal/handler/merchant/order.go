// internal/handler/merchant/order.go
package merchant

import (
	"strconv"

	"github.com/example/epay-go/internal/middleware"
	"github.com/example/epay-go/internal/service"
	"github.com/example/epay-go/pkg/response"
	"github.com/gin-gonic/gin"
)

// ListOrders 订单列表
func ListOrders(c *gin.Context) {
	merchantID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var status *int8
	if s := c.Query("status"); s != "" {
		v, _ := strconv.ParseInt(s, 10, 8)
		v8 := int8(v)
		status = &v8
	}

	orderService := service.NewOrderService()
	orders, total, err := orderService.List(page, pageSize, &merchantID, status)
	if err != nil {
		response.ServerError(c, "获取订单列表失败")
		return
	}

	response.SuccessPage(c, orders, total, page, pageSize)
}

// GetOrder 获取订单详情
func GetOrder(c *gin.Context) {
	merchantID := middleware.GetUserID(c)
	tradeNo := c.Param("trade_no")

	orderService := service.NewOrderService()
	order, err := orderService.GetByTradeNo(tradeNo)
	if err != nil {
		response.NotFound(c, "订单不存在")
		return
	}

	// 验证订单归属
	if order.MerchantID != merchantID {
		response.Forbidden(c, "无权查看此订单")
		return
	}

	response.Success(c, order)
}
