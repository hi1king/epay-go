// internal/api/merchant/balance_log.go
package merchant

import (
	"strconv"

	"github.com/example/epay-go/internal/middleware"
	"github.com/example/epay-go/internal/repository"
	"github.com/example/epay-go/pkg/response"
	"github.com/gin-gonic/gin"
)

// ListBalanceLogs 资金记录列表
func ListBalanceLogs(c *gin.Context) {
	merchantID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	recordRepo := repository.NewMerchantBalanceLogRepository()
	records, total, err := recordRepo.List(page, pageSize, merchantID)
	if err != nil {
		response.ServerError(c, "获取资金记录失败")
		return
	}

	response.SuccessPage(c, records, total, page, pageSize)
}
