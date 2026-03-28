// internal/handler/merchant/settlement.go
package merchant

import (
	"strconv"

	"github.com/example/epay-go/internal/middleware"
	"github.com/example/epay-go/internal/service"
	"github.com/example/epay-go/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)


// ListSettlements 结算列表
func ListSettlements(c *gin.Context) {
	merchantID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var status *int8
	if s := c.Query("status"); s != "" {
		v, _ := strconv.ParseInt(s, 10, 8)
		v8 := int8(v)
		status = &v8
	}

	settlementService := service.NewSettlementService()
	settlements, total, err := settlementService.List(page, pageSize, &merchantID, status)
	if err != nil {
		response.ServerError(c, "获取结算列表失败")
		return
	}

	response.SuccessPage(c, settlements, total, page, pageSize)
}

// ApplySettlementRequest 申请结算请求
type ApplySettlementRequest struct {
	Amount      string `json:"amount" binding:"required"`
	AccountType string `json:"account_type" binding:"required,oneof=alipay bank"`
	AccountNo   string `json:"account_no" binding:"required"`
	AccountName string `json:"account_name" binding:"required"`
}

// ApplySettlement 申请结算
func ApplySettlement(c *gin.Context) {
	merchantID := middleware.GetUserID(c)

	var req ApplySettlementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数错误: "+err.Error())
		return
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		response.ParamError(c, "金额格式错误")
		return
	}

	settlementService := service.NewSettlementService()
	settlement, err := settlementService.Apply(&service.ApplyRequest{
		MerchantID:  merchantID,
		Amount:      amount,
		AccountType: req.AccountType,
		AccountNo:   req.AccountNo,
		AccountName: req.AccountName,
	})
	if err != nil {
		response.Error(c, response.CodeParamError, err.Error())
		return
	}

	response.Success(c, settlement)
}
