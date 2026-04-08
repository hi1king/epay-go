// internal/handler/merchant/refund.go
package merchant

import (
	"strconv"

	"github.com/example/epay-go/internal/middleware"
	"github.com/example/epay-go/internal/service"
	"github.com/example/epay-go/pkg/response"
	"github.com/gin-gonic/gin"
)

// CreateRefund 创建退款申请
func CreateRefund(c *gin.Context) {
	merchantID := middleware.GetUserID(c)

	var req service.CreateRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数错误: "+err.Error())
		return
	}

	refundService := service.NewRefundService()
	refund, err := refundService.CreateRefund(merchantID, &req)
	if err != nil {
		response.Error(c, response.CodeServerError, err.Error())
		return
	}

	response.Success(c, refund)
}

// ListRefunds 退款列表
func ListRefunds(c *gin.Context) {
	merchantID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var status *int8
	if s := c.Query("status"); s != "" {
		v, _ := strconv.ParseInt(s, 10, 8)
		v8 := int8(v)
		status = &v8
	}

	refundService := service.NewRefundService()
	refunds, total, err := refundService.List(page, pageSize, &merchantID, status)
	if err != nil {
		response.ServerError(c, "获取退款列表失败")
		return
	}

	response.SuccessPage(c, refunds, total, page, pageSize)
}
