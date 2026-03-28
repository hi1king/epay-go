# EPay Go 重构 - 阶段三：API Handler 层

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 实现管理后台 API、商户端 API 和对外支付 API，完成后端所有接口开发。

**Architecture:** Handler 层负责参数校验、调用 Service、返回响应。使用 Gin 路由分组，JWT 中间件保护需要认证的接口。

**Tech Stack:** Go, Gin, JWT, Validator

**前置条件:** 阶段一、阶段二已完成。

---

## Task 1: 创建路由注册入口

**Files:**
- Create: `epay-go/internal/router/router.go`

**Step 1: 创建路由注册文件**

```go
// internal/router/router.go
package router

import (
	"github.com/example/epay-go/internal/handler/admin"
	"github.com/example/epay-go/internal/handler/merchant"
	"github.com/example/epay-go/internal/handler/payment"
	"github.com/example/epay-go/internal/middleware"
	"github.com/example/epay-go/pkg/jwt"
	"github.com/gin-gonic/gin"
)

// Setup 设置所有路由
func Setup(r *gin.Engine) {
	// 对外支付 API（无需登录）
	payAPI := r.Group("/api/pay")
	{
		payAPI.POST("/create", payment.CreateOrder)
		payAPI.GET("/query", payment.QueryOrder)
		payAPI.POST("/notify/:channel", payment.HandleNotify)
		payAPI.GET("/return/:channel", payment.HandleReturn)
	}

	// 管理后台 API
	adminAPI := r.Group("/admin")
	{
		// 无需认证
		adminAPI.POST("/auth/login", admin.Login)

		// 需要认证
		adminAuth := adminAPI.Group("")
		adminAuth.Use(middleware.JWTAuth(jwt.TokenTypeAdmin))
		{
			adminAuth.POST("/auth/logout", admin.Logout)
			adminAuth.GET("/dashboard", admin.Dashboard)

			// 商户管理
			adminAuth.GET("/merchants", admin.ListMerchants)
			adminAuth.GET("/merchants/:id", admin.GetMerchant)
			adminAuth.PUT("/merchants/:id", admin.UpdateMerchant)
			adminAuth.PATCH("/merchants/:id/status", admin.UpdateMerchantStatus)

			// 订单管理
			adminAuth.GET("/orders", admin.ListOrders)
			adminAuth.GET("/orders/:trade_no", admin.GetOrder)
			adminAuth.POST("/orders/:trade_no/renotify", admin.RenotifyOrder)

			// 通道管理
			adminAuth.GET("/channels", admin.ListChannels)
			adminAuth.POST("/channels", admin.CreateChannel)
			adminAuth.GET("/channels/:id", admin.GetChannel)
			adminAuth.PUT("/channels/:id", admin.UpdateChannel)
			adminAuth.DELETE("/channels/:id", admin.DeleteChannel)

			// 结算管理
			adminAuth.GET("/settlements", admin.ListSettlements)
			adminAuth.PATCH("/settlements/:id/approve", admin.ApproveSettlement)
			adminAuth.PATCH("/settlements/:id/reject", admin.RejectSettlement)
		}
	}

	// 商户端 API
	merchantAPI := r.Group("/merchant")
	{
		// 无需认证
		merchantAPI.POST("/auth/login", merchant.Login)
		merchantAPI.POST("/auth/register", merchant.Register)

		// 需要认证
		merchantAuth := merchantAPI.Group("")
		merchantAuth.Use(middleware.JWTAuth(jwt.TokenTypeMerchant))
		{
			merchantAuth.POST("/auth/logout", merchant.Logout)
			merchantAuth.GET("/dashboard", merchant.Dashboard)

			// 个人信息
			merchantAuth.GET("/profile", merchant.GetProfile)
			merchantAuth.PUT("/profile", merchant.UpdateProfile)
			merchantAuth.PUT("/profile/password", merchant.UpdatePassword)
			merchantAuth.POST("/profile/reset-key", merchant.ResetAPIKey)

			// 订单
			merchantAuth.GET("/orders", merchant.ListOrders)
			merchantAuth.GET("/orders/:trade_no", merchant.GetOrder)

			// 结算
			merchantAuth.GET("/settlements", merchant.ListSettlements)
			merchantAuth.POST("/settlements", merchant.ApplySettlement)

			// 资金记录
			merchantAuth.GET("/records", merchant.ListRecords)
		}
	}
}
```

**Step 2: 提交**

```bash
cd d:/project/payment/epay-go
git add internal/router/router.go
git commit -m "feat: add router setup with all api routes"
```

---

## Task 2: 创建管理员认证 Handler

**Files:**
- Create: `epay-go/internal/handler/admin/auth.go`

**Step 1: 创建管理员认证 Handler**

```go
// internal/handler/admin/auth.go
package admin

import (
	"github.com/example/epay-go/internal/service"
	"github.com/example/epay-go/pkg/jwt"
	"github.com/example/epay-go/pkg/response"
	"github.com/gin-gonic/gin"
)

var adminService = service.NewAdminService()

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 管理员登录
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数错误")
		return
	}

	admin, err := adminService.Login(&service.AdminLoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		response.Error(c, response.CodeUnauthorized, err.Error())
		return
	}

	// 生成 Token
	token, err := jwt.GenerateToken(admin.ID, admin.Username, jwt.TokenTypeAdmin)
	if err != nil {
		response.ServerError(c, "生成Token失败")
		return
	}

	response.Success(c, gin.H{
		"token": token,
		"admin": gin.H{
			"id":       admin.ID,
			"username": admin.Username,
			"role":     admin.Role,
		},
	})
}

// Logout 退出登录
func Logout(c *gin.Context) {
	// JWT 无状态，客户端删除 token 即可
	response.Success(c, nil)
}
```

**Step 2: 提交**

```bash
git add internal/handler/admin/auth.go
git commit -m "feat: add admin auth handler (login/logout)"
```

---

## Task 3: 创建管理后台仪表盘和商户管理 Handler

**Files:**
- Create: `epay-go/internal/handler/admin/dashboard.go`
- Create: `epay-go/internal/handler/admin/merchant.go`

**Step 1: 创建仪表盘 Handler**

```go
// internal/handler/admin/dashboard.go
package admin

import (
	"github.com/example/epay-go/internal/service"
	"github.com/example/epay-go/pkg/response"
	"github.com/gin-gonic/gin"
)

var orderService = service.NewOrderService()
var merchantService = service.NewMerchantService()

// Dashboard 仪表盘数据
func Dashboard(c *gin.Context) {
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
```

**Step 2: 创建商户管理 Handler**

```go
// internal/handler/admin/merchant.go
package admin

import (
	"strconv"

	"github.com/example/epay-go/pkg/response"
	"github.com/gin-gonic/gin"
)

// ListMerchants 商户列表
func ListMerchants(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var status *int8
	if s := c.Query("status"); s != "" {
		v, _ := strconv.ParseInt(s, 10, 8)
		v8 := int8(v)
		status = &v8
	}

	merchants, total, err := merchantService.List(page, pageSize, status)
	if err != nil {
		response.ServerError(c, "获取商户列表失败")
		return
	}

	response.SuccessPage(c, merchants, total, page, pageSize)
}

// GetMerchant 获取商户详情
func GetMerchant(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ParamError(c, "无效的商户ID")
		return
	}

	merchant, err := merchantService.GetByID(id)
	if err != nil {
		response.NotFound(c, "商户不存在")
		return
	}

	response.Success(c, merchant)
}

// UpdateMerchantRequest 更新商户请求
type UpdateMerchantRequest struct {
	Email string `json:"email"`
	Phone string `json:"phone"`
}

// UpdateMerchant 更新商户信息
func UpdateMerchant(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ParamError(c, "无效的商户ID")
		return
	}

	var req UpdateMerchantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数错误")
		return
	}

	merchant, err := merchantService.GetByID(id)
	if err != nil {
		response.NotFound(c, "商户不存在")
		return
	}

	merchant.Email = req.Email
	merchant.Phone = req.Phone

	// TODO: 调用 service 更新

	response.Success(c, merchant)
}

// UpdateMerchantStatusRequest 更新状态请求
type UpdateMerchantStatusRequest struct {
	Status int8 `json:"status" binding:"oneof=0 1"`
}

// UpdateMerchantStatus 更新商户状态
func UpdateMerchantStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ParamError(c, "无效的商户ID")
		return
	}

	var req UpdateMerchantStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数错误")
		return
	}

	if err := merchantService.UpdateStatus(id, req.Status); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}
```

**Step 3: 提交**

```bash
git add internal/handler/admin/
git commit -m "feat: add admin dashboard and merchant management handlers"
```

---

## Task 4: 创建管理后台订单和通道管理 Handler

**Files:**
- Create: `epay-go/internal/handler/admin/order.go`
- Create: `epay-go/internal/handler/admin/channel.go`

**Step 1: 创建订单管理 Handler**

```go
// internal/handler/admin/order.go
package admin

import (
	"strconv"

	"github.com/example/epay-go/internal/service"
	"github.com/example/epay-go/pkg/response"
	"github.com/gin-gonic/gin"
)

var notifyService = service.NewNotifyService()

// ListOrders 订单列表
func ListOrders(c *gin.Context) {
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

	orders, total, err := orderService.List(page, pageSize, merchantID, status)
	if err != nil {
		response.ServerError(c, "获取订单列表失败")
		return
	}

	response.SuccessPage(c, orders, total, page, pageSize)
}

// GetOrder 获取订单详情
func GetOrder(c *gin.Context) {
	tradeNo := c.Param("trade_no")

	order, err := orderService.GetByTradeNo(tradeNo)
	if err != nil {
		response.NotFound(c, "订单不存在")
		return
	}

	response.Success(c, order)
}

// RenotifyOrder 重新发送通知
func RenotifyOrder(c *gin.Context) {
	tradeNo := c.Param("trade_no")

	order, err := orderService.GetByTradeNo(tradeNo)
	if err != nil {
		response.NotFound(c, "订单不存在")
		return
	}

	if order.Status != 1 {
		response.Error(c, response.CodeParamError, "订单未支付，无法发送通知")
		return
	}

	if err := notifyService.SendNotify(order); err != nil {
		response.ServerError(c, "发送通知失败: "+err.Error())
		return
	}

	response.Success(c, nil)
}
```

**Step 2: 创建通道管理 Handler**

```go
// internal/handler/admin/channel.go
package admin

import (
	"strconv"

	"github.com/example/epay-go/internal/service"
	"github.com/example/epay-go/pkg/response"
	"github.com/gin-gonic/gin"
)

var channelService = service.NewChannelService()

// ListChannels 通道列表
func ListChannels(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	channels, total, err := channelService.List(page, pageSize)
	if err != nil {
		response.ServerError(c, "获取通道列表失败")
		return
	}

	response.SuccessPage(c, channels, total, page, pageSize)
}

// GetChannel 获取通道详情
func GetChannel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ParamError(c, "无效的通道ID")
		return
	}

	channel, err := channelService.GetByID(id)
	if err != nil {
		response.NotFound(c, "通道不存在")
		return
	}

	response.Success(c, channel)
}

// CreateChannel 创建通道
func CreateChannel(c *gin.Context) {
	var req service.CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数错误: "+err.Error())
		return
	}

	channel, err := channelService.Create(&req)
	if err != nil {
		response.ServerError(c, "创建通道失败: "+err.Error())
		return
	}

	response.Success(c, channel)
}

// UpdateChannel 更新通道
func UpdateChannel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ParamError(c, "无效的通道ID")
		return
	}

	var req service.UpdateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数错误: "+err.Error())
		return
	}

	if err := channelService.Update(id, &req); err != nil {
		response.ServerError(c, "更新通道失败: "+err.Error())
		return
	}

	response.Success(c, nil)
}

// DeleteChannel 删除通道
func DeleteChannel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ParamError(c, "无效的通道ID")
		return
	}

	if err := channelService.Delete(id); err != nil {
		response.ServerError(c, "删除通道失败: "+err.Error())
		return
	}

	response.Success(c, nil)
}
```

**Step 3: 提交**

```bash
git add internal/handler/admin/
git commit -m "feat: add admin order and channel management handlers"
```

---

## Task 5: 创建管理后台结算管理 Handler

**Files:**
- Create: `epay-go/internal/handler/admin/settlement.go`

**Step 1: 创建结算管理 Handler**

```go
// internal/handler/admin/settlement.go
package admin

import (
	"strconv"

	"github.com/example/epay-go/internal/service"
	"github.com/example/epay-go/pkg/response"
	"github.com/gin-gonic/gin"
)

var settlementService = service.NewSettlementService()

// ListSettlements 结算列表
func ListSettlements(c *gin.Context) {
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

	settlements, total, err := settlementService.List(page, pageSize, merchantID, status)
	if err != nil {
		response.ServerError(c, "获取结算列表失败")
		return
	}

	response.SuccessPage(c, settlements, total, page, pageSize)
}

// ApproveSettlement 审核通过
func ApproveSettlement(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ParamError(c, "无效的结算ID")
		return
	}

	if err := settlementService.Approve(id); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// RejectSettlementRequest 驳回请求
type RejectSettlementRequest struct {
	Remark string `json:"remark" binding:"required"`
}

// RejectSettlement 驳回结算
func RejectSettlement(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ParamError(c, "无效的结算ID")
		return
	}

	var req RejectSettlementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "请填写驳回原因")
		return
	}

	if err := settlementService.Reject(id, req.Remark); err != nil {
		response.ServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}
```

**Step 2: 提交**

```bash
git add internal/handler/admin/settlement.go
git commit -m "feat: add admin settlement management handler"
```

---

## Task 6: 创建商户端认证 Handler

**Files:**
- Create: `epay-go/internal/handler/merchant/auth.go`

**Step 1: 创建商户认证 Handler**

```go
// internal/handler/merchant/auth.go
package merchant

import (
	"github.com/example/epay-go/internal/service"
	"github.com/example/epay-go/pkg/jwt"
	"github.com/example/epay-go/pkg/response"
	"github.com/gin-gonic/gin"
)

var merchantService = service.NewMerchantService()

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=4,max=32"`
	Password string `json:"password" binding:"required,min=6,max=32"`
	Email    string `json:"email" binding:"omitempty,email"`
	Phone    string `json:"phone" binding:"omitempty"`
}

// Register 商户注册
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数错误: "+err.Error())
		return
	}

	merchant, err := merchantService.Register(&service.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
		Phone:    req.Phone,
	})
	if err != nil {
		response.Error(c, response.CodeParamError, err.Error())
		return
	}

	// 生成 Token
	token, err := jwt.GenerateToken(merchant.ID, merchant.Username, jwt.TokenTypeMerchant)
	if err != nil {
		response.ServerError(c, "生成Token失败")
		return
	}

	response.Success(c, gin.H{
		"token": token,
		"merchant": gin.H{
			"id":       merchant.ID,
			"username": merchant.Username,
			"api_key":  merchant.ApiKey,
		},
	})
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 商户登录
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数错误")
		return
	}

	merchant, err := merchantService.Login(&service.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		response.Error(c, response.CodeUnauthorized, err.Error())
		return
	}

	// 生成 Token
	token, err := jwt.GenerateToken(merchant.ID, merchant.Username, jwt.TokenTypeMerchant)
	if err != nil {
		response.ServerError(c, "生成Token失败")
		return
	}

	response.Success(c, gin.H{
		"token": token,
		"merchant": gin.H{
			"id":       merchant.ID,
			"username": merchant.Username,
			"balance":  merchant.Balance,
		},
	})
}

// Logout 退出登录
func Logout(c *gin.Context) {
	response.Success(c, nil)
}
```

**Step 2: 提交**

```bash
git add internal/handler/merchant/auth.go
git commit -m "feat: add merchant auth handler (register/login/logout)"
```

---

## Task 7: 创建商户端个人信息和仪表盘 Handler

**Files:**
- Create: `epay-go/internal/handler/merchant/profile.go`
- Create: `epay-go/internal/handler/merchant/dashboard.go`

**Step 1: 创建个人信息 Handler**

```go
// internal/handler/merchant/profile.go
package merchant

import (
	"github.com/example/epay-go/internal/middleware"
	"github.com/example/epay-go/pkg/response"
	"github.com/gin-gonic/gin"
)

// GetProfile 获取个人信息
func GetProfile(c *gin.Context) {
	merchantID := middleware.GetUserID(c)

	merchant, err := merchantService.GetByID(merchantID)
	if err != nil {
		response.NotFound(c, "商户不存在")
		return
	}

	response.Success(c, gin.H{
		"id":             merchant.ID,
		"username":       merchant.Username,
		"email":          merchant.Email,
		"phone":          merchant.Phone,
		"api_key":        merchant.ApiKey,
		"balance":        merchant.Balance,
		"frozen_balance": merchant.FrozenBalance,
		"status":         merchant.Status,
		"created_at":     merchant.CreatedAt,
	})
}

// UpdateProfileRequest 更新信息请求
type UpdateProfileRequest struct {
	Email string `json:"email" binding:"omitempty,email"`
	Phone string `json:"phone" binding:"omitempty"`
}

// UpdateProfile 更新个人信息
func UpdateProfile(c *gin.Context) {
	merchantID := middleware.GetUserID(c)

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数错误: "+err.Error())
		return
	}

	merchant, err := merchantService.GetByID(merchantID)
	if err != nil {
		response.NotFound(c, "商户不存在")
		return
	}

	merchant.Email = req.Email
	merchant.Phone = req.Phone

	// TODO: 保存更新

	response.Success(c, nil)
}

// UpdatePasswordRequest 修改密码请求
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=32"`
}

// UpdatePassword 修改密码
func UpdatePassword(c *gin.Context) {
	merchantID := middleware.GetUserID(c)

	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数错误: "+err.Error())
		return
	}

	if err := merchantService.UpdatePassword(merchantID, req.OldPassword, req.NewPassword); err != nil {
		response.Error(c, response.CodeParamError, err.Error())
		return
	}

	response.Success(c, nil)
}

// ResetAPIKey 重置API密钥
func ResetAPIKey(c *gin.Context) {
	merchantID := middleware.GetUserID(c)

	newKey, err := merchantService.ResetAPIKey(merchantID)
	if err != nil {
		response.ServerError(c, "重置密钥失败")
		return
	}

	response.Success(c, gin.H{
		"api_key": newKey,
	})
}
```

**Step 2: 创建仪表盘 Handler**

```go
// internal/handler/merchant/dashboard.go
package merchant

import (
	"github.com/example/epay-go/internal/middleware"
	"github.com/example/epay-go/internal/service"
	"github.com/example/epay-go/pkg/response"
	"github.com/gin-gonic/gin"
)

var orderService = service.NewOrderService()

// Dashboard 商户仪表盘
func Dashboard(c *gin.Context) {
	merchantID := middleware.GetUserID(c)

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
```

**Step 3: 提交**

```bash
git add internal/handler/merchant/
git commit -m "feat: add merchant profile and dashboard handlers"
```

---

## Task 8: 创建商户端订单、结算、资金记录 Handler

**Files:**
- Create: `epay-go/internal/handler/merchant/order.go`
- Create: `epay-go/internal/handler/merchant/settlement.go`
- Create: `epay-go/internal/handler/merchant/record.go`

**Step 1: 创建订单 Handler**

```go
// internal/handler/merchant/order.go
package merchant

import (
	"strconv"

	"github.com/example/epay-go/internal/middleware"
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
```

**Step 2: 创建结算 Handler**

```go
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

var settlementService = service.NewSettlementService()

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
```

**Step 3: 创建资金记录 Handler**

```go
// internal/handler/merchant/record.go
package merchant

import (
	"strconv"

	"github.com/example/epay-go/internal/middleware"
	"github.com/example/epay-go/internal/repository"
	"github.com/example/epay-go/pkg/response"
	"github.com/gin-gonic/gin"
)

var recordRepo = repository.NewBalanceRecordRepository()

// ListRecords 资金记录列表
func ListRecords(c *gin.Context) {
	merchantID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	records, total, err := recordRepo.List(page, pageSize, merchantID)
	if err != nil {
		response.ServerError(c, "获取资金记录失败")
		return
	}

	response.SuccessPage(c, records, total, page, pageSize)
}
```

**Step 4: 提交**

```bash
git add internal/handler/merchant/
git commit -m "feat: add merchant order, settlement, record handlers"
```

---

## Task 9: 创建对外支付 API Handler

**Files:**
- Create: `epay-go/internal/handler/payment/create.go`
- Create: `epay-go/internal/handler/payment/query.go`
- Create: `epay-go/internal/handler/payment/notify.go`

**Step 1: 创建签名验证工具**

```go
// pkg/sign/sign.go
package sign

import (
	"crypto/md5"
	"encoding/hex"
	"net/url"
	"sort"
	"strings"
)

// VerifyMD5Sign 验证MD5签名（与原epay兼容）
func VerifyMD5Sign(params url.Values, key, sign string) bool {
	// 按key排序
	var keys []string
	for k := range params {
		if k != "sign" && k != "sign_type" && params.Get(k) != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// 拼接字符串
	var buf strings.Builder
	for i, k := range keys {
		if i > 0 {
			buf.WriteString("&")
		}
		buf.WriteString(k)
		buf.WriteString("=")
		buf.WriteString(params.Get(k))
	}
	buf.WriteString(key)

	// MD5
	hash := md5.Sum([]byte(buf.String()))
	expected := hex.EncodeToString(hash[:])

	return strings.EqualFold(expected, sign)
}

// GenerateMD5Sign 生成MD5签名
func GenerateMD5Sign(params url.Values, key string) string {
	var keys []string
	for k := range params {
		if k != "sign" && k != "sign_type" && params.Get(k) != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var buf strings.Builder
	for i, k := range keys {
		if i > 0 {
			buf.WriteString("&")
		}
		buf.WriteString(k)
		buf.WriteString("=")
		buf.WriteString(params.Get(k))
	}
	buf.WriteString(key)

	hash := md5.Sum([]byte(buf.String()))
	return hex.EncodeToString(hash[:])
}
```

**Step 2: 创建订单创建 Handler**

```go
// internal/handler/payment/create.go
package payment

import (
	"context"
	"net/url"

	"github.com/example/epay-go/internal/service"
	"github.com/example/epay-go/pkg/response"
	"github.com/example/epay-go/pkg/sign"
	"github.com/example/epay-go/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

var merchantService = service.NewMerchantService()
var orderService = service.NewOrderService()

// CreateOrderRequest 创建订单请求（兼容原epay）
type CreateOrderRequest struct {
	Pid        string `form:"pid" binding:"required"`         // 商户ID
	Type       string `form:"type" binding:"required"`        // 支付类型
	OutTradeNo string `form:"out_trade_no" binding:"required"`// 商户订单号
	NotifyURL  string `form:"notify_url" binding:"required"`  // 异步通知地址
	ReturnURL  string `form:"return_url"`                     // 同步跳转地址
	Name       string `form:"name" binding:"required"`        // 商品名称
	Money      string `form:"money" binding:"required"`       // 金额
	Sign       string `form:"sign" binding:"required"`        // 签名
	SignType   string `form:"sign_type"`                      // 签名类型
}

// CreateOrder 创建支付订单
func CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBind(&req); err != nil {
		response.ParamError(c, "参数错误: "+err.Error())
		return
	}

	// 获取商户信息
	merchant, err := merchantService.GetByAPIKey(req.Pid)
	if err != nil {
		response.Error(c, response.CodeParamError, "商户不存在")
		return
	}

	if merchant.Status != 1 {
		response.Error(c, response.CodeForbidden, "商户已被禁用")
		return
	}

	// 验证签名
	params := url.Values{}
	params.Set("pid", req.Pid)
	params.Set("type", req.Type)
	params.Set("out_trade_no", req.OutTradeNo)
	params.Set("notify_url", req.NotifyURL)
	params.Set("name", req.Name)
	params.Set("money", req.Money)
	if req.ReturnURL != "" {
		params.Set("return_url", req.ReturnURL)
	}

	if !sign.VerifyMD5Sign(params, merchant.ApiKey, req.Sign) {
		response.Error(c, response.CodeParamError, "签名验证失败")
		return
	}

	// 解析金额
	amount, err := decimal.NewFromString(req.Money)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		response.ParamError(c, "金额格式错误")
		return
	}

	// 创建订单
	orderReq := &service.CreateOrderRequest{
		MerchantID: merchant.ID,
		OutTradeNo: req.OutTradeNo,
		Amount:     amount,
		Name:       req.Name,
		PayType:    req.Type,
		NotifyURL:  req.NotifyURL,
		ReturnURL:  req.ReturnURL,
		ClientIP:   utils.GetClientIP(c),
		PayMethod:  c.DefaultQuery("pay_method", "scan"),
	}

	orderResp, err := orderService.Create(context.Background(), orderReq)
	if err != nil {
		response.Error(c, response.CodeServerError, err.Error())
		return
	}

	response.Success(c, orderResp)
}
```

**Step 3: 创建订单查询 Handler**

```go
// internal/handler/payment/query.go
package payment

import (
	"net/url"

	"github.com/example/epay-go/pkg/response"
	"github.com/example/epay-go/pkg/sign"
	"github.com/gin-gonic/gin"
)

// QueryOrderRequest 查询订单请求
type QueryOrderRequest struct {
	Pid        string `form:"pid" binding:"required"`
	TradeNo    string `form:"trade_no"`
	OutTradeNo string `form:"out_trade_no"`
	Sign       string `form:"sign" binding:"required"`
}

// QueryOrder 查询订单
func QueryOrder(c *gin.Context) {
	var req QueryOrderRequest
	if err := c.ShouldBind(&req); err != nil {
		response.ParamError(c, "参数错误")
		return
	}

	if req.TradeNo == "" && req.OutTradeNo == "" {
		response.ParamError(c, "trade_no 和 out_trade_no 不能同时为空")
		return
	}

	// 获取商户
	merchant, err := merchantService.GetByAPIKey(req.Pid)
	if err != nil {
		response.Error(c, response.CodeParamError, "商户不存在")
		return
	}

	// 验证签名
	params := url.Values{}
	params.Set("pid", req.Pid)
	if req.TradeNo != "" {
		params.Set("trade_no", req.TradeNo)
	}
	if req.OutTradeNo != "" {
		params.Set("out_trade_no", req.OutTradeNo)
	}

	if !sign.VerifyMD5Sign(params, merchant.ApiKey, req.Sign) {
		response.Error(c, response.CodeParamError, "签名验证失败")
		return
	}

	// 查询订单
	var order interface{}
	if req.TradeNo != "" {
		order, err = orderService.GetByTradeNo(req.TradeNo)
	} else {
		// TODO: 按商户订单号查询
		order, err = orderService.GetByTradeNo(req.OutTradeNo)
	}

	if err != nil {
		response.NotFound(c, "订单不存在")
		return
	}

	response.Success(c, order)
}
```

**Step 4: 创建回调处理 Handler**

```go
// internal/handler/payment/notify.go
package payment

import (
	"io"
	"log"
	"net/http"

	"github.com/example/epay-go/internal/model"
	intPayment "github.com/example/epay-go/internal/payment"
	"github.com/example/epay-go/internal/repository"
	"github.com/example/epay-go/internal/service"
	"github.com/gin-gonic/gin"
)

var channelRepo = repository.NewChannelRepository()
var notifyService = service.NewNotifyService()

// HandleNotify 处理支付回调
func HandleNotify(c *gin.Context) {
	channelPlugin := c.Param("channel")

	// 获取通道配置
	channel, err := channelRepo.GetByPluginAndPayType(channelPlugin, "")
	if err != nil {
		log.Printf("Channel not found: %s", channelPlugin)
		c.String(http.StatusOK, "fail")
		return
	}

	// 创建适配器
	adapter, err := intPayment.NewAdapter(channel.Plugin, channel.Config)
	if err != nil {
		log.Printf("Create adapter failed: %v", err)
		c.String(http.StatusOK, "fail")
		return
	}

	// 解析回调
	result, err := adapter.ParseNotify(c.Request.Context(), c.Request)
	if err != nil {
		log.Printf("Parse notify failed: %v", err)
		c.String(http.StatusOK, "fail")
		return
	}

	// 处理支付结果
	if result.Status == "success" {
		if err := orderService.ProcessPayNotify(result.TradeNo, result.ApiTradeNo, result.Buyer, result.Amount); err != nil {
			log.Printf("Process notify failed: %v", err)
			c.String(http.StatusOK, "fail")
			return
		}

		// 发送商户通知
		order, _ := orderService.GetByTradeNo(result.TradeNo)
		if order != nil && order.Status == model.OrderStatusPaid {
			go notifyService.SendNotify(order)
		}
	}

	// 返回成功响应
	c.String(http.StatusOK, adapter.NotifySuccess())
}

// HandleReturn 处理同步跳转
func HandleReturn(c *gin.Context) {
	// 从参数获取订单号
	tradeNo := c.Query("out_trade_no")
	if tradeNo == "" {
		// 尝试从 body 读取
		body, _ := io.ReadAll(c.Request.Body)
		log.Printf("Return body: %s", string(body))
	}

	order, err := orderService.GetByTradeNo(tradeNo)
	if err != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}

	// 跳转到商户 return_url
	if order.ReturnURL != "" {
		c.Redirect(http.StatusFound, order.ReturnURL)
		return
	}

	c.String(http.StatusOK, "支付完成")
}
```

**Step 5: 提交**

```bash
git add pkg/sign/sign.go internal/handler/payment/
git commit -m "feat: add payment api handlers (create/query/notify)"
```

---

## Task 10: 更新主程序入口，注册路由

**Files:**
- Modify: `epay-go/cmd/server/main.go`

**Step 1: 更新主程序**

```go
// cmd/server/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/example/epay-go/internal/cache"
	"github.com/example/epay-go/internal/config"
	"github.com/example/epay-go/internal/database"
	"github.com/example/epay-go/internal/middleware"
	"github.com/example/epay-go/internal/router"
	"github.com/example/epay-go/internal/service"
	"github.com/gin-gonic/gin"

	// 注册支付适配器
	_ "github.com/example/epay-go/internal/payment"
)

func main() {
	// 加载配置
	if err := config.Load("config.yaml"); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	cfg := config.Get()

	// 初始化数据库
	if err := database.Init(); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}
	defer database.Close()

	// 数据库迁移
	if err := database.Migrate(); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 初始化 Redis
	if err := cache.Init(); err != nil {
		log.Fatalf("Failed to init redis: %v", err)
	}
	defer cache.Close()

	// 初始化默认管理员
	adminService := service.NewAdminService()
	if err := adminService.InitDefaultAdmin(); err != nil {
		log.Printf("Failed to init default admin: %v", err)
	}

	// 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

	// 创建 Gin 引擎
	r := gin.New()

	// 全局中间件
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.Cors())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 注册所有路由
	router.Setup(r)

	// 启动异步通知工作协程
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	notifyService := service.NewNotifyService()
	go notifyService.StartNotifyWorker(ctx)

	// 优雅关闭
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("Shutting down server...")
		cancel()
	}()

	// 启动服务
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
```

**Step 2: 验证编译**

Run: `cd d:/project/payment/epay-go && go build -o epay-server ./cmd/server`
Expected: 编译成功

**Step 3: 提交**

```bash
git add cmd/server/main.go
git commit -m "feat: register all routes in main entry"
```

---

## 阶段三完成检查清单

- [ ] 路由注册入口 (`internal/router/router.go`)
- [ ] 管理员认证 Handler (login/logout)
- [ ] 管理后台仪表盘和商户管理 Handler
- [ ] 管理后台订单和通道管理 Handler
- [ ] 管理后台结算管理 Handler
- [ ] 商户端认证 Handler (register/login/logout)
- [ ] 商户端个人信息和仪表盘 Handler
- [ ] 商户端订单、结算、资金记录 Handler
- [ ] 对外支付 API Handler (create/query/notify)
- [ ] 签名验证工具 (`pkg/sign/sign.go`)
- [ ] 主程序集成路由

---

## API 接口汇总

### 对外支付 API
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/pay/create` | 创建支付订单 |
| GET | `/api/pay/query` | 查询订单状态 |
| POST | `/api/pay/notify/:channel` | 支付回调 |
| GET | `/api/pay/return/:channel` | 同步跳转 |

### 管理后台 API
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/admin/auth/login` | 管理员登录 |
| GET | `/admin/dashboard` | 仪表盘 |
| GET | `/admin/merchants` | 商户列表 |
| GET/PUT | `/admin/merchants/:id` | 商户详情/更新 |
| GET | `/admin/orders` | 订单列表 |
| GET/POST | `/admin/channels` | 通道列表/创建 |
| GET | `/admin/settlements` | 结算列表 |
| PATCH | `/admin/settlements/:id/approve` | 审核通过 |

### 商户端 API
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/merchant/auth/register` | 商户注册 |
| POST | `/merchant/auth/login` | 商户登录 |
| GET | `/merchant/dashboard` | 仪表盘 |
| GET/PUT | `/merchant/profile` | 个人信息 |
| GET | `/merchant/orders` | 订单列表 |
| GET/POST | `/merchant/settlements` | 结算列表/申请 |
| GET | `/merchant/records` | 资金记录 |

---

**下一阶段：** 阶段四将实现 Vue3 前端（管理后台、商户中心、收银台）。
