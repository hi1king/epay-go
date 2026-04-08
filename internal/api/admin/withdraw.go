// internal/api/admin/withdraw.go
package admin

import (
	"strconv"

	"github.com/example/epay-go/internal/service"
	"github.com/example/epay-go/pkg/response"
	"github.com/gin-gonic/gin"
)

// ListWithdraws 提现列表
func ListWithdraws(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var merchantID *int64
	if mid := c.Query("merchant_id"); mid != "" {
		v, _ := strconv.ParseInt(mid, 10, 64)
		merchantID = &v
	}

	var status *int8
	if s := c.Query("status"); s != "" {
		v, _ := strconv.ParseInt(s, 10, 8)
		v8 := int8(v)
		status = &v8
	}

	withdrawService := service.NewWithdrawService()
	withdraws, total, err := withdrawService.List(page, pageSize, merchantID, status)
	if err != nil {
		response.ServerError(c, "获取提现列表失败")
		return
	}

	response.SuccessPage(c, withdraws, total, page, pageSize)
}

// ApproveWithdraw 审核通过
func ApproveWithdraw(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ParamError(c, "无效的提现ID")
		return
	}

	withdrawService := service.NewWithdrawService()
	if err := withdrawService.Approve(id); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// RejectWithdrawRequest 驳回请求
type RejectWithdrawRequest struct {
	Remark string `json:"remark" binding:"required"`
}

// RejectWithdraw 驳回提现
func RejectWithdraw(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ParamError(c, "无效的提现ID")
		return
	}

	var req RejectWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "请填写驳回原因")
		return
	}

	withdrawService := service.NewWithdrawService()
	if err := withdrawService.Reject(id, req.Remark); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}
