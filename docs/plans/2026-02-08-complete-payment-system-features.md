# 完整支付系统功能细化实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**目标：** 参考 epay 项目，为 epay-go 补全企业级聚合支付系统的完整功能，包括管理员管理、订单退款、统计报表、系统配置、安全增强等核心模块。

**架构：** 采用三层架构（Handler-Service-Repository），前端使用 Vue 3 + TypeScript + Arco Design，后端使用 Go + Gin + GORM + PostgreSQL。

**技术栈：** Go 1.21+, Gin, GORM, PostgreSQL, Redis, Vue 3, TypeScript, Arco Design, Vite

---

## 功能优先级划分

### 🔴 P0 - 核心业务功能（必须实现）
1. 订单退款功能
2. 管理员管理模块
3. 商户余额调整
4. 系统配置管理

### 🟡 P1 - 重要运营功能（优先实现）
5. 操作日志记录
6. 统计报表与图表
7. 数据导出功能
8. 订单补单功能

### 🟢 P2 - 增强功能（后续实现）
9. 登录日志与安全增强
10. 消息通知系统
11. API 文档生成

---

## Task 1: 订单退款功能

**文件：**
- Modify: `internal/model/order.go`
- Create: `internal/model/refund.go`
- Create: `internal/repository/refund.go`
- Create: `internal/service/refund.go`
- Create: `internal/handler/admin/refund.go`
- Create: `internal/handler/merchant/refund.go`
- Modify: `internal/router/router.go`
- Create: `web/src/views/admin/Refunds.vue`
- Create: `web/src/views/merchant/Refunds.vue`
- Modify: `web/src/api/admin.ts`
- Modify: `web/src/api/merchant.ts`
- Modify: `web/src/api/types.ts`

### 步骤 1: 创建退款数据模型

在 `internal/model/refund.go` 创建退款模型：

```go
// internal/model/refund.go
package model

import (
	"time"
	"gorm.io/gorm"
)

// Refund 退款订单
type Refund struct {
	ID              int64          `gorm:"primaryKey" json:"id"`
	RefundNo        string         `gorm:"size:32;uniqueIndex;not null" json:"refund_no"`        // 退款单号
	TradeNo         string         `gorm:"size:32;index;not null" json:"trade_no"`               // 原订单号
	MerchantID      int64          `gorm:"index;not null" json:"merchant_id"`                    // 商户ID
	OrderID         int64          `gorm:"index;not null" json:"order_id"`                       // 原订单ID
	ChannelID       int64          `gorm:"not null" json:"channel_id"`                           // 支付通道ID
	Amount          string         `gorm:"type:decimal(10,2);not null" json:"amount"`            // 退款金额
	RefundFee       string         `gorm:"type:decimal(10,2);default:0.00" json:"refund_fee"`    // 退款手续费
	Reason          string         `gorm:"size:200" json:"reason"`                               // 退款原因
	Status          int8           `gorm:"default:0;index" json:"status"`                        // 状态：0待处理 1成功 2失败
	ApiRefundNo     string         `gorm:"size:64" json:"api_refund_no"`                         // 第三方退款单号
	FailReason      string         `gorm:"size:200" json:"fail_reason"`                          // 失败原因
	NotifyURL       string         `gorm:"size:255" json:"notify_url"`                           // 异步通知地址
	NotifyStatus    int8           `gorm:"default:0" json:"notify_status"`                       // 通知状态：0未通知 1已通知
	NotifyCount     int            `gorm:"default:0" json:"notify_count"`                        // 通知次数
	NextNotifyTime  *time.Time     `json:"next_notify_time"`                                     // 下次通知时间
	ProcessedAt     *time.Time     `json:"processed_at"`                                         // 处理完成时间
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// 退款状态常量
const (
	RefundStatusPending = 0 // 待处理
	RefundStatusSuccess = 1 // 成功
	RefundStatusFailed  = 2 // 失败
)

// TableName 指定表名
func (Refund) TableName() string {
	return "refunds"
}
```

### 步骤 2: 添加退款表迁移

在 `internal/database/migrate.go` 的 `Migrate()` 函数中添加：

```go
&model.Refund{},
```

### 步骤 3: 创建退款 Repository

在 `internal/repository/refund.go` 创建：

```go
// internal/repository/refund.go
package repository

import (
	"github.com/example/epay-go/internal/database"
	"github.com/example/epay-go/internal/model"
)

type RefundRepository struct {
	db *gorm.DB
}

func NewRefundRepository() *RefundRepository {
	return &RefundRepository{db: database.Get()}
}

// Create 创建退款单
func (r *RefundRepository) Create(refund *model.Refund) error {
	return r.db.Create(refund).Error
}

// GetByRefundNo 根据退款单号查询
func (r *RefundRepository) GetByRefundNo(refundNo string) (*model.Refund, error) {
	var refund model.Refund
	err := r.db.Where("refund_no = ?", refundNo).First(&refund).Error
	return &refund, err
}

// GetByTradeNo 根据订单号查询退款记录
func (r *RefundRepository) GetByTradeNo(tradeNo string) ([]*model.Refund, error) {
	var refunds []*model.Refund
	err := r.db.Where("trade_no = ?", tradeNo).Find(&refunds).Error
	return refunds, err
}

// List 分页查询退款列表
func (r *RefundRepository) List(page, pageSize int, merchantID *int64, status *int8) ([]*model.Refund, int64, error) {
	var refunds []*model.Refund
	var total int64

	query := r.db.Model(&model.Refund{})

	if merchantID != nil {
		query = query.Where("merchant_id = ?", *merchantID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Order("id DESC").Limit(pageSize).Offset(offset).Find(&refunds).Error
	return refunds, total, err
}

// Update 更新退款单
func (r *RefundRepository) Update(refund *model.Refund) error {
	return r.db.Save(refund).Error
}
```

### 步骤 4: 创建退款 Service

在 `internal/service/refund.go` 创建：

```go
// internal/service/refund.go
package service

import (
	"errors"
	"time"

	"github.com/example/epay-go/internal/model"
	"github.com/example/epay-go/internal/repository"
	"github.com/example/epay-go/pkg/utils"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type RefundService struct {
	refundRepo   *repository.RefundRepository
	orderRepo    *repository.OrderRepository
	merchantRepo *repository.MerchantRepository
	recordRepo   *repository.BalanceRecordRepository
}

func NewRefundService() *RefundService {
	return &RefundService{
		refundRepo:   repository.NewRefundRepository(),
		orderRepo:    repository.NewOrderRepository(),
		merchantRepo: repository.NewMerchantRepository(),
		recordRepo:   repository.NewBalanceRecordRepository(),
	}
}

// CreateRefundRequest 创建退款请求
type CreateRefundRequest struct {
	TradeNo   string `json:"trade_no" binding:"required"`
	Amount    string `json:"amount" binding:"required"`
	Reason    string `json:"reason"`
	NotifyURL string `json:"notify_url"`
}

// CreateRefund 创建退款单
func (s *RefundService) CreateRefund(merchantID int64, req *CreateRefundRequest) (*model.Refund, error) {
	// 查询原订单
	order, err := s.orderRepo.GetByTradeNo(req.TradeNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("订单不存在")
		}
		return nil, err
	}

	// 验证商户权限
	if order.MerchantID != merchantID {
		return nil, errors.New("无权操作此订单")
	}

	// 验证订单状态
	if order.Status != model.OrderStatusPaid {
		return nil, errors.New("订单未支付，无法退款")
	}

	// 验证退款金额
	refundAmount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, errors.New("退款金额格式错误")
	}
	orderAmount, _ := decimal.NewFromString(order.Amount)
	if refundAmount.GreaterThan(orderAmount) {
		return nil, errors.New("退款金额不能大于订单金额")
	}

	// 检查是否已有退款记录
	existingRefunds, err := s.refundRepo.GetByTradeNo(req.TradeNo)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 计算已退款总额
	totalRefunded := decimal.Zero
	for _, r := range existingRefunds {
		if r.Status == model.RefundStatusSuccess {
			amt, _ := decimal.NewFromString(r.Amount)
			totalRefunded = totalRefunded.Add(amt)
		}
	}

	if totalRefunded.Add(refundAmount).GreaterThan(orderAmount) {
		return nil, errors.New("退款总额超过订单金额")
	}

	// 创建退款单
	refund := &model.Refund{
		RefundNo:    utils.GenerateOrderNo("R"),
		TradeNo:     order.TradeNo,
		MerchantID:  order.MerchantID,
		OrderID:     order.ID,
		ChannelID:   order.ChannelID,
		Amount:      req.Amount,
		RefundFee:   "0.00",
		Reason:      req.Reason,
		Status:      model.RefundStatusPending,
		NotifyURL:   req.NotifyURL,
	}

	if err := s.refundRepo.Create(refund); err != nil {
		return nil, err
	}

	return refund, nil
}

// ProcessRefund 处理退款（管理员审核）
func (s *RefundService) ProcessRefund(refundNo string, success bool, failReason string) error {
	refund, err := s.refundRepo.GetByRefundNo(refundNo)
	if err != nil {
		return err
	}

	if refund.Status != model.RefundStatusPending {
		return errors.New("退款单已处理")
	}

	now := time.Now()
	if success {
		refund.Status = model.RefundStatusSuccess
		refund.ProcessedAt = &now

		// 退款到商户余额
		amount, _ := decimal.NewFromString(refund.Amount)
		merchant, err := s.merchantRepo.GetByID(refund.MerchantID)
		if err != nil {
			return err
		}

		beforeBalance, _ := decimal.NewFromString(merchant.Balance)
		afterBalance := beforeBalance.Add(amount)

		if err := s.merchantRepo.UpdateBalance(refund.MerchantID, afterBalance.String()); err != nil {
			return err
		}

		// 记录资金变动
		record := &model.BalanceRecord{
			MerchantID:    refund.MerchantID,
			Action:        1, // 增加
			Amount:        refund.Amount,
			BeforeBalance: beforeBalance.String(),
			AfterBalance:  afterBalance.String(),
			Type:          "refund",
			TradeNo:       refund.RefundNo,
		}
		s.recordRepo.Create(record)
	} else {
		refund.Status = model.RefundStatusFailed
		refund.FailReason = failReason
		refund.ProcessedAt = &now
	}

	return s.refundRepo.Update(refund)
}

// GetRefundByNo 根据退款单号查询
func (s *RefundService) GetRefundByNo(refundNo string) (*model.Refund, error) {
	return s.refundRepo.GetByRefundNo(refundNo)
}

// List 退款列表
func (s *RefundService) List(page, pageSize int, merchantID *int64, status *int8) ([]*model.Refund, int64, error) {
	return s.refundRepo.List(page, pageSize, merchantID, status)
}
```

### 步骤 5: 创建管理后台退款 Handler

在 `internal/handler/admin/refund.go` 创建：

```go
// internal/handler/admin/refund.go
package admin

import (
	"strconv"

	"github.com/example/epay-go/internal/service"
	"github.com/example/epay-go/pkg/response"
	"github.com/gin-gonic/gin"
)

// ListRefunds 退款列表
func ListRefunds(c *gin.Context) {
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

	refundService := service.NewRefundService()
	refunds, total, err := refundService.List(page, pageSize, merchantID, status)
	if err != nil {
		response.ServerError(c, "获取退款列表失败")
		return
	}

	response.SuccessPage(c, refunds, total, page, pageSize)
}

// ProcessRefundRequest 处理退款请求
type ProcessRefundRequest struct {
	Success    bool   `json:"success" binding:"required"`
	FailReason string `json:"fail_reason"`
}

// ProcessRefund 处理退款
func ProcessRefund(c *gin.Context) {
	refundNo := c.Param("refund_no")

	var req ProcessRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数错误")
		return
	}

	refundService := service.NewRefundService()
	if err := refundService.ProcessRefund(refundNo, req.Success, req.FailReason); err != nil {
		response.Error(c, response.CodeServerError, err.Error())
		return
	}

	response.Success(c, nil)
}
```

### 步骤 6: 创建商户端退款 Handler

在 `internal/handler/merchant/refund.go` 创建：

```go
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
```

### 步骤 7: 注册路由

在 `internal/router/router.go` 添加：

```go
// 管理后台 - 退款管理
adminGroup.GET("/refunds", admin.ListRefunds)
adminGroup.POST("/refunds/:refund_no/process", admin.ProcessRefund)

// 商户中心 - 退款管理
merchantGroup.POST("/refunds", merchant.CreateRefund)
merchantGroup.GET("/refunds", merchant.ListRefunds)
```

### 步骤 8: 添加前端类型定义

在 `web/src/api/types.ts` 添加：

```typescript
export interface Refund {
  id: number
  refund_no: string
  trade_no: string
  merchant_id: number
  order_id: number
  channel_id: number
  amount: string
  refund_fee: string
  reason: string
  status: number // 0待处理 1成功 2失败
  api_refund_no: string
  fail_reason: string
  notify_url: string
  notify_status: number
  notify_count: number
  next_notify_time: string | null
  processed_at: string | null
  created_at: string
}
```

### 步骤 9: 添加管理后台 API

在 `web/src/api/admin.ts` 添加：

```typescript
// 退款管理
export const getRefunds = (params: any) =>
  request.get<{ list: Refund[]; total: number }>('/admin/refunds', { params })

export const processRefund = (refundNo: string, data: { success: boolean; fail_reason?: string }) =>
  request.post(`/admin/refunds/${refundNo}/process`, data)
```

### 步骤 10: 添加商户端 API

在 `web/src/api/merchant.ts` 添加：

```typescript
// 退款管理
export const createRefund = (data: {
  trade_no: string
  amount: string
  reason?: string
  notify_url?: string
}) => request.post('/merchant/refunds', data)

export const getRefunds = (params: any) =>
  request.get<{ list: Refund[]; total: number }>('/merchant/refunds', { params })
```

### 步骤 11: 创建管理后台退款页面

创建 `web/src/views/admin/Refunds.vue`

### 步骤 12: 创建商户端退款页面

创建 `web/src/views/merchant/Refunds.vue`

### 步骤 13: 测试退款功能

```bash
# 重新构建后端
docker-compose build backend
docker-compose up -d backend

# 重启前端
cd web && npm run dev
```

测试步骤：
1. 商户端创建退款申请
2. 管理后台审核退款
3. 验证商户余额变动
4. 验证资金记录

### 步骤 14: 提交代码

```bash
git add .
git commit -m "feat: add refund functionality

- Add Refund model and database migration
- Implement refund repository, service, and handlers
- Add admin refund approval workflow
- Add merchant refund application
- Refund amount validation and balance refund
- Add frontend refund management pages

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 2: 管理员管理模块

**文件：**
- Modify: `internal/model/admin.go`
- Create: `internal/service/admin_manage.go`
- Create: `internal/handler/admin/admin_manage.go`
- Modify: `internal/router/router.go`
- Create: `web/src/views/admin/Admins.vue`

### 步骤 1: 扩展管理员模型

在 `internal/model/admin.go` 添加字段：

```go
Email       string         `gorm:"size:100" json:"email"`
Status      int8           `gorm:"default:1" json:"status"` // 状态：0禁用 1启用
```

### 步骤 2: 创建管理员管理 Service

在 `internal/service/admin_manage.go` 实现管理员的增删改查功能。

### 步骤 3: 创建管理员管理 Handler

### 步骤 4: 实现前端页面

### 步骤 5: 测试并提交

---

## Task 3: 商户余额调整

**文件：**
- Create: `internal/handler/admin/balance.go`
- Modify: `web/src/views/admin/Merchants.vue`

### 步骤 1: 创建余额调整 Handler

实现后台手动增加/减少商户余额的功能。

### 步骤 2: 添加资金变动记录

记录所有余额变动操作。

### 步骤 3: 实现前端界面

在商户管理页面添加余额调整按钮和对话框。

### 步骤 4: 测试并提交

---

## Task 4: 系统配置管理

**文件：**
- Create: `internal/service/config.go`
- Create: `internal/handler/admin/config.go`
- Create: `web/src/views/admin/Settings.vue`

### 步骤 1: 实现配置 Service

支持系统参数的读取和更新。

### 步骤 2: 定义配置项

- 站点名称
- 结算手续费率
- 异步通知重试次数
- 异步通知间隔时间
- 商户注册开关
- 测试模式开关

### 步骤 3: 实现前端配置页面

### 步骤 4: 测试并提交

---

## Task 5: 操作日志记录

**文件：**
- Create: `internal/model/log.go`
- Create: `internal/repository/log.go`
- Create: `internal/service/log.go`
- Create: `internal/middleware/logger.go`
- Create: `internal/handler/admin/log.go`
- Create: `web/src/views/admin/Logs.vue`

### 步骤 1: 创建日志模型

### 步骤 2: 实现日志中间件

自动记录管理员和商户的重要操作。

### 步骤 3: 实现日志查询界面

### 步骤 4: 测试并提交

---

## Task 6: 统计报表与图表

**文件：**
- Create: `internal/service/statistics.go`
- Create: `internal/handler/admin/statistics.go`
- Modify: `web/src/views/admin/Dashboard.vue`
- Modify: `web/src/views/merchant/Dashboard.vue`

### 步骤 1: 实现统计 Service

- 按日期统计订单
- 按通道统计订单
- 按商户统计订单
- 收入趋势统计

### 步骤 2: 安装图表库

```bash
cd web
npm install @arco-design/web-vue echarts
```

### 步骤 3: 实现仪表盘图表

- 订单趋势折线图
- 收入统计柱状图
- 通道占比饼图
- 商户排行榜

### 步骤 4: 测试并提交

---

## Task 7: 数据导出功能

**文件：**
- Create: `internal/handler/admin/export.go`
- Modify: `web/src/views/admin/Orders.vue`

### 步骤 1: 安装 Excel 库

```bash
go get github.com/xuri/excelize/v2
```

### 步骤 2: 实现订单导出

支持导出 Excel 格式的订单数据。

### 步骤 3: 添加导出按钮

### 步骤 4: 测试并提交

---

## Task 8: 订单补单功能

**文件：**
- Create: `internal/handler/admin/order_repair.go`
- Modify: `web/src/views/admin/Orders.vue`

### 步骤 1: 实现补单逻辑

手动标记订单为已支付，并触发异步通知。

### 步骤 2: 添加权限验证

只有超级管理员可以补单。

### 步骤 3: 实现前端界面

### 步骤 4: 测试并提交

---

## Task 9: 登录日志与安全增强

**文件：**
- Create: `internal/model/login_log.go`
- Create: `internal/middleware/rate_limit.go`
- Create: `internal/middleware/ip_whitelist.go`

### 步骤 1: 记录登录日志

记录登录时间、IP、设备信息、登录结果。

### 步骤 2: 实现 IP 白名单

### 步骤 3: 实现频率限制

防止暴力破解。

### 步骤 4: 测试并提交

---

## Task 10: 消息通知系统

**文件：**
- Create: `internal/service/notification.go`
- Create: `pkg/mail/sender.go`
- Create: `pkg/sms/sender.go`

### 步骤 1: 实现邮件通知

使用 SMTP 发送邮件。

### 步骤 2: 实现短信通知

集成短信服务商 API。

### 步骤 3: 定义通知场景

- 订单支付成功
- 结算审核通过
- 余额不足提醒

### 步骤 4: 测试并提交

---

## Task 11: API 文档生成

**文件：**
- 安装 Swagger
- 添加 API 注释

### 步骤 1: 安装 Swagger

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

### 步骤 2: 添加 Swagger 注释

为所有 Handler 添加注释。

### 步骤 3: 生成文档

```bash
swag init -g cmd/server/main.go
```

### 步骤 4: 访问文档

http://localhost:8099/swagger/index.html

### 步骤 5: 提交代码

---

## 总结

本计划涵盖了 11 个核心功能模块，完整参考了 epay 项目的功能设计，将使 epay-go 成为一个功能完善的企业级聚合支付系统。

**实施建议：**
1. 按优先级顺序实施（P0 → P1 → P2）
2. 每个 Task 完成后进行充分测试
3. 保持代码质量和架构一致性
4. 及时更新文档

**预计工作量：**
- P0 功能：2-3 周
- P1 功能：2 周
- P2 功能：1-2 周

总计：5-7 周可完成全部功能细化。
