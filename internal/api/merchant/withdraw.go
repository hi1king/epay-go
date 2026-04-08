// internal/api/merchant/withdraw.go
package merchant

import (
	"strconv"

	"github.com/example/epay-go/internal/middleware"
	"github.com/example/epay-go/internal/service"
	"github.com/example/epay-go/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// ListWithdraws 提现列表
func ListWithdraws(c *gin.Context) {
	merchantID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var status *int8
	if s := c.Query("status"); s != "" {
		v, _ := strconv.ParseInt(s, 10, 8)
		v8 := int8(v)
		status = &v8
	}

	withdrawService := service.NewWithdrawService()
	withdraws, total, err := withdrawService.List(page, pageSize, &merchantID, status)
	if err != nil {
		response.ServerError(c, "获取提现列表失败")
		return
	}

	response.SuccessPage(c, withdraws, total, page, pageSize)
}

// ApplyWithdrawRequest 申请提现请求
type ApplyWithdrawRequest struct {
	Amount      string `json:"amount" binding:"required"`
	AccountType string `json:"account_type" binding:"required,oneof=alipay bank"`
	AccountNo   string `json:"account_no" binding:"required"`
	AccountName string `json:"account_name" binding:"required"`
}

// ApplyWithdraw 申请提现
func ApplyWithdraw(c *gin.Context) {
	merchantID := middleware.GetUserID(c)

	var req ApplyWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数错误: "+err.Error())
		return
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		response.ParamError(c, "金额格式错误")
		return
	}

	withdrawService := service.NewWithdrawService()
	withdraw, err := withdrawService.Apply(&service.ApplyWithdrawRequest{
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

	response.Success(c, withdraw)
}
