# EPay Go 重构 - 阶段二：后端核心业务功能

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 实现商户模块、支付通道管理、订单模块、支付适配器层（支付宝、微信）、异步回调处理和结算模块。

**Architecture:** 采用 handler → service → repository 三层架构，支付渠道使用适配器模式实现统一接口，使用 gopay SDK 对接支付宝和微信支付。

**Tech Stack:** Go, Gin, GORM, gopay, bcrypt, UUID

**前置条件:** 阶段一已完成，项目基础框架已搭建。

---

## Task 1: 创建通用工具函数

**Files:**
- Create: `epay-go/pkg/utils/utils.go`
- Create: `epay-go/pkg/utils/password.go`
- Create: `epay-go/pkg/utils/trade.go`

**Step 1: 创建通用工具函数**

```go
// pkg/utils/utils.go
package utils

import (
	"crypto/md5"
	"encoding/hex"
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetClientIP 获取客户端真实IP
func GetClientIP(c *gin.Context) string {
	// 优先从 X-Forwarded-For 获取
	xff := c.GetHeader("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	// 其次从 X-Real-IP 获取
	xri := c.GetHeader("X-Real-IP")
	if xri != "" && net.ParseIP(xri) != nil {
		return xri
	}

	// 最后使用 RemoteAddr
	ip, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
	return ip
}

// MD5 计算MD5哈希
func MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// ContainsString 检查字符串切片是否包含指定字符串
func ContainsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
```

**Step 2: 创建密码工具**

```go
// pkg/utils/password.go
package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword 使用bcrypt加密密码
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 验证密码
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
```

**Step 3: 创建订单号生成工具**

```go
// pkg/utils/trade.go
package utils

import (
	"crypto/rand"
	"fmt"
	"time"
)

// GenerateTradeNo 生成订单号 (时间戳 + 随机数，共24位)
func GenerateTradeNo() string {
	timestamp := time.Now().Format("20060102150405")
	random := make([]byte, 5)
	rand.Read(random)
	return fmt.Sprintf("%s%x", timestamp, random)
}

// GenerateSettleNo 生成结算单号
func GenerateSettleNo() string {
	timestamp := time.Now().Format("20060102150405")
	random := make([]byte, 4)
	rand.Read(random)
	return fmt.Sprintf("S%s%x", timestamp, random)
}

// GenerateAPIKey 生成商户API密钥 (32位)
func GenerateAPIKey() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
```

**Step 4: 提交**

```bash
cd d:/project/payment/epay-go
git add pkg/utils/
git commit -m "feat: add utility functions for password, trade no generation"
```

---

## Task 2: 创建 Repository 层 - 商户

**Files:**
- Create: `epay-go/internal/repository/merchant.go`

**Step 1: 创建商户 Repository**

```go
// internal/repository/merchant.go
package repository

import (
	"github.com/example/epay-go/internal/database"
	"github.com/example/epay-go/internal/model"
	"gorm.io/gorm"
)

type MerchantRepository struct {
	db *gorm.DB
}

func NewMerchantRepository() *MerchantRepository {
	return &MerchantRepository{db: database.Get()}
}

// Create 创建商户
func (r *MerchantRepository) Create(merchant *model.Merchant) error {
	return r.db.Create(merchant).Error
}

// GetByID 根据ID获取商户
func (r *MerchantRepository) GetByID(id int64) (*model.Merchant, error) {
	var merchant model.Merchant
	err := r.db.First(&merchant, id).Error
	if err != nil {
		return nil, err
	}
	return &merchant, nil
}

// GetByUsername 根据用户名获取商户
func (r *MerchantRepository) GetByUsername(username string) (*model.Merchant, error) {
	var merchant model.Merchant
	err := r.db.Where("username = ?", username).First(&merchant).Error
	if err != nil {
		return nil, err
	}
	return &merchant, nil
}

// GetByAPIKey 根据API Key获取商户
func (r *MerchantRepository) GetByAPIKey(apiKey string) (*model.Merchant, error) {
	var merchant model.Merchant
	err := r.db.Where("api_key = ?", apiKey).First(&merchant).Error
	if err != nil {
		return nil, err
	}
	return &merchant, nil
}

// Update 更新商户
func (r *MerchantRepository) Update(merchant *model.Merchant) error {
	return r.db.Save(merchant).Error
}

// UpdateBalance 更新余额 (使用事务)
func (r *MerchantRepository) UpdateBalance(tx *gorm.DB, id int64, amount float64) error {
	return tx.Model(&model.Merchant{}).Where("id = ?", id).
		Update("balance", gorm.Expr("balance + ?", amount)).Error
}

// List 分页查询商户列表
func (r *MerchantRepository) List(page, pageSize int, status *int8) ([]model.Merchant, int64, error) {
	var merchants []model.Merchant
	var total int64

	query := r.db.Model(&model.Merchant{})
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&merchants).Error
	if err != nil {
		return nil, 0, err
	}

	return merchants, total, nil
}

// ExistsByUsername 检查用户名是否存在
func (r *MerchantRepository) ExistsByUsername(username string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Merchant{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}
```

**Step 2: 提交**

```bash
git add internal/repository/merchant.go
git commit -m "feat: add merchant repository"
```

---

## Task 3: 创建 Repository 层 - 管理员、通道、订单

**Files:**
- Create: `epay-go/internal/repository/admin.go`
- Create: `epay-go/internal/repository/channel.go`
- Create: `epay-go/internal/repository/order.go`

**Step 1: 创建管理员 Repository**

```go
// internal/repository/admin.go
package repository

import (
	"time"

	"github.com/example/epay-go/internal/database"
	"github.com/example/epay-go/internal/model"
	"gorm.io/gorm"
)

type AdminRepository struct {
	db *gorm.DB
}

func NewAdminRepository() *AdminRepository {
	return &AdminRepository{db: database.Get()}
}

// Create 创建管理员
func (r *AdminRepository) Create(admin *model.Admin) error {
	return r.db.Create(admin).Error
}

// GetByID 根据ID获取管理员
func (r *AdminRepository) GetByID(id int64) (*model.Admin, error) {
	var admin model.Admin
	err := r.db.First(&admin, id).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

// GetByUsername 根据用户名获取管理员
func (r *AdminRepository) GetByUsername(username string) (*model.Admin, error) {
	var admin model.Admin
	err := r.db.Where("username = ?", username).First(&admin).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

// UpdateLastLogin 更新最后登录时间
func (r *AdminRepository) UpdateLastLogin(id int64) error {
	now := time.Now()
	return r.db.Model(&model.Admin{}).Where("id = ?", id).Update("last_login_at", &now).Error
}

// ExistsByUsername 检查用户名是否存在
func (r *AdminRepository) ExistsByUsername(username string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Admin{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

// Count 统计管理员数量
func (r *AdminRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.Admin{}).Count(&count).Error
	return count, err
}
```

**Step 2: 创建通道 Repository**

```go
// internal/repository/channel.go
package repository

import (
	"github.com/example/epay-go/internal/database"
	"github.com/example/epay-go/internal/model"
	"gorm.io/gorm"
)

type ChannelRepository struct {
	db *gorm.DB
}

func NewChannelRepository() *ChannelRepository {
	return &ChannelRepository{db: database.Get()}
}

// Create 创建通道
func (r *ChannelRepository) Create(channel *model.Channel) error {
	return r.db.Create(channel).Error
}

// GetByID 根据ID获取通道
func (r *ChannelRepository) GetByID(id int64) (*model.Channel, error) {
	var channel model.Channel
	err := r.db.First(&channel, id).Error
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

// Update 更新通道
func (r *ChannelRepository) Update(channel *model.Channel) error {
	return r.db.Save(channel).Error
}

// Delete 删除通道
func (r *ChannelRepository) Delete(id int64) error {
	return r.db.Delete(&model.Channel{}, id).Error
}

// List 分页查询通道列表
func (r *ChannelRepository) List(page, pageSize int) ([]model.Channel, int64, error) {
	var channels []model.Channel
	var total int64

	err := r.db.Model(&model.Channel{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = r.db.Offset(offset).Limit(pageSize).Order("sort ASC, id ASC").Find(&channels).Error
	if err != nil {
		return nil, 0, err
	}

	return channels, total, nil
}

// ListEnabled 获取所有启用的通道
func (r *ChannelRepository) ListEnabled() ([]model.Channel, error) {
	var channels []model.Channel
	err := r.db.Where("status = ?", 1).Order("sort ASC, id ASC").Find(&channels).Error
	return channels, err
}

// GetByPluginAndPayType 根据插件和支付类型获取可用通道
func (r *ChannelRepository) GetByPluginAndPayType(plugin, payType string) (*model.Channel, error) {
	var channel model.Channel
	err := r.db.Where("plugin = ? AND status = 1 AND pay_types LIKE ?", plugin, "%"+payType+"%").
		Order("sort ASC").First(&channel).Error
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

// GetAvailableByPayType 根据支付类型获取可用通道
func (r *ChannelRepository) GetAvailableByPayType(payType string) (*model.Channel, error) {
	var channel model.Channel
	err := r.db.Where("status = 1 AND pay_types LIKE ?", "%"+payType+"%").
		Order("sort ASC").First(&channel).Error
	if err != nil {
		return nil, err
	}
	return &channel, nil
}
```

**Step 3: 创建订单 Repository**

```go
// internal/repository/order.go
package repository

import (
	"time"

	"github.com/example/epay-go/internal/database"
	"github.com/example/epay-go/internal/model"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{db: database.Get()}
}

// Create 创建订单
func (r *OrderRepository) Create(order *model.Order) error {
	return r.db.Create(order).Error
}

// GetByID 根据ID获取订单
func (r *OrderRepository) GetByID(id int64) (*model.Order, error) {
	var order model.Order
	err := r.db.Preload("Merchant").Preload("Channel").First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// GetByTradeNo 根据系统订单号获取订单
func (r *OrderRepository) GetByTradeNo(tradeNo string) (*model.Order, error) {
	var order model.Order
	err := r.db.Preload("Merchant").Preload("Channel").
		Where("trade_no = ?", tradeNo).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// GetByOutTradeNo 根据商户订单号获取订单
func (r *OrderRepository) GetByOutTradeNo(merchantID int64, outTradeNo string) (*model.Order, error) {
	var order model.Order
	err := r.db.Where("merchant_id = ? AND out_trade_no = ?", merchantID, outTradeNo).
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// Update 更新订单
func (r *OrderRepository) Update(order *model.Order) error {
	return r.db.Save(order).Error
}

// UpdateStatus 更新订单状态
func (r *OrderRepository) UpdateStatus(tradeNo string, status int8) error {
	updates := map[string]interface{}{"status": status}
	if status == model.OrderStatusPaid {
		now := time.Now()
		updates["paid_at"] = &now
	}
	return r.db.Model(&model.Order{}).Where("trade_no = ?", tradeNo).Updates(updates).Error
}

// UpdateNotifyStatus 更新通知状态
func (r *OrderRepository) UpdateNotifyStatus(tradeNo string, status int8, nextNotifyAt *time.Time) error {
	updates := map[string]interface{}{
		"notify_status": status,
		"notify_count":  gorm.Expr("notify_count + 1"),
	}
	if nextNotifyAt != nil {
		updates["next_notify_at"] = nextNotifyAt
	}
	return r.db.Model(&model.Order{}).Where("trade_no = ?", tradeNo).Updates(updates).Error
}

// UpdatePayInfo 更新支付信息
func (r *OrderRepository) UpdatePayInfo(tradeNo, apiTradeNo, buyer string) error {
	now := time.Now()
	return r.db.Model(&model.Order{}).Where("trade_no = ?", tradeNo).Updates(map[string]interface{}{
		"api_trade_no": apiTradeNo,
		"buyer":        buyer,
		"status":       model.OrderStatusPaid,
		"paid_at":      &now,
	}).Error
}

// List 分页查询订单列表
func (r *OrderRepository) List(page, pageSize int, merchantID *int64, status *int8) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64

	query := r.db.Model(&model.Order{})
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
	err = query.Preload("Merchant").Preload("Channel").
		Offset(offset).Limit(pageSize).Order("id DESC").Find(&orders).Error
	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// GetPendingNotifyOrders 获取待通知的订单
func (r *OrderRepository) GetPendingNotifyOrders(limit int) ([]model.Order, error) {
	var orders []model.Order
	err := r.db.Where("status = ? AND notify_status < ? AND (next_notify_at IS NULL OR next_notify_at <= ?)",
		model.OrderStatusPaid, model.NotifyStatusSuccess, time.Now()).
		Limit(limit).Find(&orders).Error
	return orders, err
}

// GetTodayStats 获取今日统计
func (r *OrderRepository) GetTodayStats(merchantID *int64) (int64, decimal.Decimal, error) {
	var count int64
	var amount decimal.Decimal

	today := time.Now().Format("2006-01-02")
	query := r.db.Model(&model.Order{}).
		Where("status = ? AND DATE(created_at) = ?", model.OrderStatusPaid, today)

	if merchantID != nil {
		query = query.Where("merchant_id = ?", *merchantID)
	}

	err := query.Count(&count).Error
	if err != nil {
		return 0, decimal.Zero, err
	}

	var result struct {
		Total decimal.Decimal
	}
	err = query.Select("COALESCE(SUM(amount), 0) as total").Scan(&result).Error
	if err != nil {
		return 0, decimal.Zero, err
	}

	return count, result.Total, nil
}
```

**Step 4: 提交**

```bash
git add internal/repository/
git commit -m "feat: add admin, channel, order repositories"
```

---

## Task 4: 创建 Repository 层 - 结算、资金记录

**Files:**
- Create: `epay-go/internal/repository/settlement.go`
- Create: `epay-go/internal/repository/record.go`

**Step 1: 创建结算 Repository**

```go
// internal/repository/settlement.go
package repository

import (
	"github.com/example/epay-go/internal/database"
	"github.com/example/epay-go/internal/model"
	"gorm.io/gorm"
)

type SettlementRepository struct {
	db *gorm.DB
}

func NewSettlementRepository() *SettlementRepository {
	return &SettlementRepository{db: database.Get()}
}

// Create 创建结算记录
func (r *SettlementRepository) Create(settlement *model.Settlement) error {
	return r.db.Create(settlement).Error
}

// GetByID 根据ID获取结算记录
func (r *SettlementRepository) GetByID(id int64) (*model.Settlement, error) {
	var settlement model.Settlement
	err := r.db.Preload("Merchant").First(&settlement, id).Error
	if err != nil {
		return nil, err
	}
	return &settlement, nil
}

// GetBySettleNo 根据结算单号获取记录
func (r *SettlementRepository) GetBySettleNo(settleNo string) (*model.Settlement, error) {
	var settlement model.Settlement
	err := r.db.Where("settle_no = ?", settleNo).First(&settlement).Error
	if err != nil {
		return nil, err
	}
	return &settlement, nil
}

// Update 更新结算记录
func (r *SettlementRepository) Update(settlement *model.Settlement) error {
	return r.db.Save(settlement).Error
}

// UpdateStatus 更新结算状态
func (r *SettlementRepository) UpdateStatus(id int64, status int8, remark string) error {
	updates := map[string]interface{}{
		"status": status,
		"remark": remark,
	}
	return r.db.Model(&model.Settlement{}).Where("id = ?", id).Updates(updates).Error
}

// List 分页查询结算列表
func (r *SettlementRepository) List(page, pageSize int, merchantID *int64, status *int8) ([]model.Settlement, int64, error) {
	var settlements []model.Settlement
	var total int64

	query := r.db.Model(&model.Settlement{})
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
	err = query.Preload("Merchant").Offset(offset).Limit(pageSize).Order("id DESC").Find(&settlements).Error
	if err != nil {
		return nil, 0, err
	}

	return settlements, total, nil
}

// HasPendingSettlement 检查是否有待处理的结算
func (r *SettlementRepository) HasPendingSettlement(merchantID int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.Settlement{}).
		Where("merchant_id = ? AND status IN ?", merchantID, []int8{model.SettleStatusPending, model.SettleStatusProcessing}).
		Count(&count).Error
	return count > 0, err
}
```

**Step 2: 创建资金记录 Repository**

```go
// internal/repository/record.go
package repository

import (
	"github.com/example/epay-go/internal/database"
	"github.com/example/epay-go/internal/model"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type BalanceRecordRepository struct {
	db *gorm.DB
}

func NewBalanceRecordRepository() *BalanceRecordRepository {
	return &BalanceRecordRepository{db: database.Get()}
}

// Create 创建资金记录
func (r *BalanceRecordRepository) Create(record *model.BalanceRecord) error {
	return r.db.Create(record).Error
}

// CreateWithTx 在事务中创建资金记录
func (r *BalanceRecordRepository) CreateWithTx(tx *gorm.DB, record *model.BalanceRecord) error {
	return tx.Create(record).Error
}

// List 分页查询资金记录
func (r *BalanceRecordRepository) List(page, pageSize int, merchantID int64) ([]model.BalanceRecord, int64, error) {
	var records []model.BalanceRecord
	var total int64

	query := r.db.Model(&model.BalanceRecord{}).Where("merchant_id = ?", merchantID)

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// AddBalanceRecord 添加资金变动记录 (带事务)
func AddBalanceRecord(tx *gorm.DB, merchantID int64, action int8, amount decimal.Decimal, beforeBalance, afterBalance decimal.Decimal, recordType, tradeNo string) error {
	record := &model.BalanceRecord{
		MerchantID:    merchantID,
		Action:        action,
		Amount:        amount,
		BeforeBalance: beforeBalance,
		AfterBalance:  afterBalance,
		Type:          recordType,
		TradeNo:       tradeNo,
	}
	return tx.Create(record).Error
}
```

**Step 3: 提交**

```bash
git add internal/repository/
git commit -m "feat: add settlement and balance record repositories"
```

---

## Task 5: 创建 Service 层 - 商户服务

**Files:**
- Create: `epay-go/internal/service/merchant.go`

**Step 1: 创建商户服务**

```go
// internal/service/merchant.go
package service

import (
	"errors"

	"github.com/example/epay-go/internal/model"
	"github.com/example/epay-go/internal/repository"
	"github.com/example/epay-go/pkg/utils"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type MerchantService struct {
	repo *repository.MerchantRepository
}

func NewMerchantService() *MerchantService {
	return &MerchantService{
		repo: repository.NewMerchantRepository(),
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=4,max=32"`
	Password string `json:"password" binding:"required,min=6,max=32"`
	Email    string `json:"email" binding:"omitempty,email"`
	Phone    string `json:"phone" binding:"omitempty"`
}

// Register 商户注册
func (s *MerchantService) Register(req *RegisterRequest) (*model.Merchant, error) {
	// 检查用户名是否存在
	exists, err := s.repo.ExistsByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("用户名已存在")
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// 创建商户
	merchant := &model.Merchant{
		Username: req.Username,
		Password: hashedPassword,
		Email:    req.Email,
		Phone:    req.Phone,
		ApiKey:   utils.GenerateAPIKey(),
		Balance:  decimal.Zero,
		Status:   1,
	}

	if err := s.repo.Create(merchant); err != nil {
		return nil, err
	}

	return merchant, nil
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 商户登录
func (s *MerchantService) Login(req *LoginRequest) (*model.Merchant, error) {
	merchant, err := s.repo.GetByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, err
	}

	if merchant.Status != 1 {
		return nil, errors.New("账号已被禁用")
	}

	if !utils.CheckPassword(req.Password, merchant.Password) {
		return nil, errors.New("用户名或密码错误")
	}

	return merchant, nil
}

// GetByID 根据ID获取商户
func (s *MerchantService) GetByID(id int64) (*model.Merchant, error) {
	return s.repo.GetByID(id)
}

// GetByAPIKey 根据API Key获取商户
func (s *MerchantService) GetByAPIKey(apiKey string) (*model.Merchant, error) {
	return s.repo.GetByAPIKey(apiKey)
}

// List 分页查询商户列表
func (s *MerchantService) List(page, pageSize int, status *int8) ([]model.Merchant, int64, error) {
	return s.repo.List(page, pageSize, status)
}

// UpdateStatus 更新商户状态
func (s *MerchantService) UpdateStatus(id int64, status int8) error {
	merchant, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	merchant.Status = status
	return s.repo.Update(merchant)
}

// ResetAPIKey 重置API密钥
func (s *MerchantService) ResetAPIKey(id int64) (string, error) {
	merchant, err := s.repo.GetByID(id)
	if err != nil {
		return "", err
	}
	merchant.ApiKey = utils.GenerateAPIKey()
	if err := s.repo.Update(merchant); err != nil {
		return "", err
	}
	return merchant.ApiKey, nil
}

// UpdatePassword 更新密码
func (s *MerchantService) UpdatePassword(id int64, oldPassword, newPassword string) error {
	merchant, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if !utils.CheckPassword(oldPassword, merchant.Password) {
		return errors.New("原密码错误")
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	merchant.Password = hashedPassword
	return s.repo.Update(merchant)
}
```

**Step 2: 提交**

```bash
git add internal/service/merchant.go
git commit -m "feat: add merchant service with register, login, profile management"
```

---

## Task 6: 创建 Service 层 - 管理员服务

**Files:**
- Create: `epay-go/internal/service/admin.go`

**Step 1: 创建管理员服务**

```go
// internal/service/admin.go
package service

import (
	"errors"

	"github.com/example/epay-go/internal/model"
	"github.com/example/epay-go/internal/repository"
	"github.com/example/epay-go/pkg/utils"
	"gorm.io/gorm"
)

type AdminService struct {
	repo *repository.AdminRepository
}

func NewAdminService() *AdminService {
	return &AdminService{
		repo: repository.NewAdminRepository(),
	}
}

// InitDefaultAdmin 初始化默认管理员（如果不存在）
func (s *AdminService) InitDefaultAdmin() error {
	count, err := s.repo.Count()
	if err != nil {
		return err
	}

	if count > 0 {
		return nil // 已有管理员，跳过
	}

	// 创建默认管理员
	hashedPassword, err := utils.HashPassword("admin123")
	if err != nil {
		return err
	}

	admin := &model.Admin{
		Username: "admin",
		Password: hashedPassword,
		Role:     "super",
	}

	return s.repo.Create(admin)
}

// AdminLoginRequest 管理员登录请求
type AdminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 管理员登录
func (s *AdminService) Login(req *AdminLoginRequest) (*model.Admin, error) {
	admin, err := s.repo.GetByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, err
	}

	if !utils.CheckPassword(req.Password, admin.Password) {
		return nil, errors.New("用户名或密码错误")
	}

	// 更新最后登录时间
	_ = s.repo.UpdateLastLogin(admin.ID)

	return admin, nil
}

// GetByID 根据ID获取管理员
func (s *AdminService) GetByID(id int64) (*model.Admin, error) {
	return s.repo.GetByID(id)
}

// UpdatePassword 更新密码
func (s *AdminService) UpdatePassword(id int64, oldPassword, newPassword string) error {
	admin, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if !utils.CheckPassword(oldPassword, admin.Password) {
		return errors.New("原密码错误")
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	admin.Password = hashedPassword
	return s.repo.Create(admin) // 使用 Create 更新会触发 Save
}
```

**Step 2: 提交**

```bash
git add internal/service/admin.go
git commit -m "feat: add admin service with login and password management"
```

---

## Task 7: 创建 Service 层 - 通道服务

**Files:**
- Create: `epay-go/internal/service/channel.go`

**Step 1: 创建通道服务**

```go
// internal/service/channel.go
package service

import (
	"encoding/json"

	"github.com/example/epay-go/internal/model"
	"github.com/example/epay-go/internal/repository"
)

type ChannelService struct {
	repo *repository.ChannelRepository
}

func NewChannelService() *ChannelService {
	return &ChannelService{
		repo: repository.NewChannelRepository(),
	}
}

// CreateChannelRequest 创建通道请求
type CreateChannelRequest struct {
	Name       string                 `json:"name" binding:"required"`
	Plugin     string                 `json:"plugin" binding:"required"`
	PayTypes   string                 `json:"pay_types"`
	Config     map[string]interface{} `json:"config"`
	Rate       float64                `json:"rate"`
	DailyLimit float64                `json:"daily_limit"`
	Status     int8                   `json:"status"`
	Sort       int                    `json:"sort"`
}

// Create 创建通道
func (s *ChannelService) Create(req *CreateChannelRequest) (*model.Channel, error) {
	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		return nil, err
	}

	channel := &model.Channel{
		Name:       req.Name,
		Plugin:     req.Plugin,
		PayTypes:   req.PayTypes,
		Config:     configJSON,
		Rate:       decimal.NewFromFloat(req.Rate),
		DailyLimit: decimal.NewFromFloat(req.DailyLimit),
		Status:     req.Status,
		Sort:       req.Sort,
	}

	if err := s.repo.Create(channel); err != nil {
		return nil, err
	}

	return channel, nil
}

// GetByID 根据ID获取通道
func (s *ChannelService) GetByID(id int64) (*model.Channel, error) {
	return s.repo.GetByID(id)
}

// UpdateChannelRequest 更新通道请求
type UpdateChannelRequest struct {
	Name       string                 `json:"name"`
	PayTypes   string                 `json:"pay_types"`
	Config     map[string]interface{} `json:"config"`
	Rate       float64                `json:"rate"`
	DailyLimit float64                `json:"daily_limit"`
	Status     int8                   `json:"status"`
	Sort       int                    `json:"sort"`
}

// Update 更新通道
func (s *ChannelService) Update(id int64, req *UpdateChannelRequest) error {
	channel, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if req.Name != "" {
		channel.Name = req.Name
	}
	if req.PayTypes != "" {
		channel.PayTypes = req.PayTypes
	}
	if req.Config != nil {
		configJSON, err := json.Marshal(req.Config)
		if err != nil {
			return err
		}
		channel.Config = configJSON
	}
	channel.Rate = decimal.NewFromFloat(req.Rate)
	channel.DailyLimit = decimal.NewFromFloat(req.DailyLimit)
	channel.Status = req.Status
	channel.Sort = req.Sort

	return s.repo.Update(channel)
}

// Delete 删除通道
func (s *ChannelService) Delete(id int64) error {
	return s.repo.Delete(id)
}

// List 分页查询通道列表
func (s *ChannelService) List(page, pageSize int) ([]model.Channel, int64, error) {
	return s.repo.List(page, pageSize)
}

// ListEnabled 获取所有启用的通道
func (s *ChannelService) ListEnabled() ([]model.Channel, error) {
	return s.repo.ListEnabled()
}

// GetAvailableChannel 根据支付类型获取可用通道
func (s *ChannelService) GetAvailableChannel(payType string) (*model.Channel, error) {
	return s.repo.GetAvailableByPayType(payType)
}
```

需要在文件顶部添加 decimal 导入：

```go
import (
	"encoding/json"

	"github.com/example/epay-go/internal/model"
	"github.com/example/epay-go/internal/repository"
	"github.com/shopspring/decimal"
)
```

**Step 2: 提交**

```bash
git add internal/service/channel.go
git commit -m "feat: add channel service"
```

---

## Task 8: 创建支付适配器接口

**Files:**
- Create: `epay-go/internal/payment/adapter.go`
- Create: `epay-go/internal/payment/factory.go`

**Step 1: 创建支付适配器接口**

```go
// internal/payment/adapter.go
package payment

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/shopspring/decimal"
)

// PaymentAdapter 统一支付适配器接口
type PaymentAdapter interface {
	// CreateOrder 创建支付订单，返回支付参数
	CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error)

	// QueryOrder 查询订单状态
	QueryOrder(ctx context.Context, tradeNo string) (*QueryOrderResponse, error)

	// Refund 退款
	Refund(ctx context.Context, req *RefundRequest) (*RefundResponse, error)

	// ParseNotify 解析异步回调通知
	ParseNotify(ctx context.Context, r *http.Request) (*NotifyResult, error)

	// NotifySuccess 返回回调成功响应
	NotifySuccess() string
}

// CreateOrderRequest 统一下单请求
type CreateOrderRequest struct {
	TradeNo   string            `json:"trade_no"`   // 系统订单号
	Amount    decimal.Decimal   `json:"amount"`     // 金额（元）
	Subject   string            `json:"subject"`    // 商品名称
	ClientIP  string            `json:"client_ip"`  // 客户端IP
	NotifyURL string            `json:"notify_url"` // 异步通知地址
	ReturnURL string            `json:"return_url"` // 同步跳转地址
	PayMethod string            `json:"pay_method"` // 支付方式: scan/h5/jsapi/app/web
	Extra     map[string]string `json:"extra"`      // 扩展参数
}

// CreateOrderResponse 统一下单响应
type CreateOrderResponse struct {
	PayType   string `json:"pay_type"`   // redirect(跳转) / qrcode(二维码) / jsapi(JS调起)
	PayURL    string `json:"pay_url"`    // 支付链接或二维码内容
	PayParams string `json:"pay_params"` // JSAPI 支付参数（JSON）
}

// QueryOrderResponse 查询订单响应
type QueryOrderResponse struct {
	TradeNo    string          `json:"trade_no"`
	ApiTradeNo string          `json:"api_trade_no"`
	Amount     decimal.Decimal `json:"amount"`
	Status     string          `json:"status"` // pending/paid/refunded/closed
	PaidAt     string          `json:"paid_at"`
}

// RefundRequest 退款请求
type RefundRequest struct {
	TradeNo    string          `json:"trade_no"`
	RefundNo   string          `json:"refund_no"`
	Amount     decimal.Decimal `json:"amount"`
	RefundDesc string          `json:"refund_desc"`
}

// RefundResponse 退款响应
type RefundResponse struct {
	RefundNo     string `json:"refund_no"`
	ApiRefundNo  string `json:"api_refund_no"`
	Status       string `json:"status"` // success/processing/failed
	ErrorMessage string `json:"error_message,omitempty"`
}

// NotifyResult 统一回调结果
type NotifyResult struct {
	TradeNo    string          `json:"trade_no"`     // 系统订单号
	ApiTradeNo string          `json:"api_trade_no"` // 上游订单号
	Amount     decimal.Decimal `json:"amount"`       // 支付金额
	Buyer      string          `json:"buyer"`        // 买家标识
	Status     string          `json:"status"`       // success / fail
}

// ChannelConfig 通道配置解析
type ChannelConfig struct {
	Raw json.RawMessage
}

func (c *ChannelConfig) Unmarshal(v interface{}) error {
	return json.Unmarshal(c.Raw, v)
}
```

**Step 2: 创建适配器工厂**

```go
// internal/payment/factory.go
package payment

import (
	"encoding/json"
	"fmt"
)

// AdapterFactory 适配器工厂函数类型
type AdapterFactory func(config json.RawMessage) (PaymentAdapter, error)

// 注册的适配器工厂
var adapters = make(map[string]AdapterFactory)

// Register 注册适配器
func Register(plugin string, factory AdapterFactory) {
	adapters[plugin] = factory
}

// NewAdapter 创建适配器实例
func NewAdapter(plugin string, config json.RawMessage) (PaymentAdapter, error) {
	factory, ok := adapters[plugin]
	if !ok {
		return nil, fmt.Errorf("unsupported payment plugin: %s", plugin)
	}
	return factory(config)
}

// GetSupportedPlugins 获取支持的插件列表
func GetSupportedPlugins() []string {
	plugins := make([]string, 0, len(adapters))
	for plugin := range adapters {
		plugins = append(plugins, plugin)
	}
	return plugins
}
```

**Step 3: 提交**

```bash
git add internal/payment/
git commit -m "feat: add payment adapter interface and factory"
```

---

## Task 9: 创建支付宝适配器

**Files:**
- Create: `epay-go/internal/payment/alipay.go`

**Step 1: 创建支付宝适配器**

```go
// internal/payment/alipay.go
package payment

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/alipay"
	"github.com/shopspring/decimal"
)

// AlipayConfig 支付宝配置
type AlipayConfig struct {
	AppID        string `json:"app_id"`
	PrivateKey   string `json:"private_key"`
	PublicKey    string `json:"public_key"`       // 支付宝公钥
	AppPublicKey string `json:"app_public_key"`   // 应用公钥证书（可选）
	IsProd       bool   `json:"is_prod"`          // 是否生产环境
	SignType     string `json:"sign_type"`        // RSA2
}

// AlipayAdapter 支付宝适配器
type AlipayAdapter struct {
	client *alipay.Client
	config *AlipayConfig
}

// NewAlipayAdapter 创建支付宝适配器
func NewAlipayAdapter(configJSON json.RawMessage) (PaymentAdapter, error) {
	var cfg AlipayConfig
	if err := json.Unmarshal(configJSON, &cfg); err != nil {
		return nil, err
	}

	client, err := alipay.NewClient(cfg.AppID, cfg.PrivateKey, cfg.IsProd)
	if err != nil {
		return nil, err
	}

	// 设置支付宝公钥
	if err := client.SetCertSnByContent(nil, nil, []byte(cfg.PublicKey)); err != nil {
		// 如果证书方式失败，尝试公钥方式
		client.SetAliPayPublicKey(cfg.PublicKey)
	}

	return &AlipayAdapter{
		client: client,
		config: &cfg,
	}, nil
}

// CreateOrder 创建支付订单
func (a *AlipayAdapter) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error) {
	amount := req.Amount.StringFixed(2)

	bm := make(gopay.BodyMap)
	bm.Set("out_trade_no", req.TradeNo)
	bm.Set("total_amount", amount)
	bm.Set("subject", req.Subject)
	bm.Set("notify_url", req.NotifyURL)

	switch req.PayMethod {
	case "scan", "qrcode":
		// 扫码支付
		bm.Set("product_code", "FACE_TO_FACE_PAYMENT")
		resp, err := a.client.TradePrecreate(ctx, bm)
		if err != nil {
			return nil, err
		}
		if resp.Response.Code != "10000" {
			return nil, errors.New(resp.Response.SubMsg)
		}
		return &CreateOrderResponse{
			PayType: "qrcode",
			PayURL:  resp.Response.QrCode,
		}, nil

	case "h5", "wap":
		// H5支付
		bm.Set("product_code", "QUICK_WAP_WAY")
		bm.Set("return_url", req.ReturnURL)
		payURL, err := a.client.TradeWapPay(ctx, bm)
		if err != nil {
			return nil, err
		}
		return &CreateOrderResponse{
			PayType: "redirect",
			PayURL:  payURL,
		}, nil

	case "web", "pc":
		// PC网页支付
		bm.Set("product_code", "FAST_INSTANT_TRADE_PAY")
		bm.Set("return_url", req.ReturnURL)
		payURL, err := a.client.TradePagePay(ctx, bm)
		if err != nil {
			return nil, err
		}
		return &CreateOrderResponse{
			PayType: "redirect",
			PayURL:  payURL,
		}, nil

	default:
		return nil, errors.New("unsupported pay method: " + req.PayMethod)
	}
}

// QueryOrder 查询订单
func (a *AlipayAdapter) QueryOrder(ctx context.Context, tradeNo string) (*QueryOrderResponse, error) {
	bm := make(gopay.BodyMap)
	bm.Set("out_trade_no", tradeNo)

	resp, err := a.client.TradeQuery(ctx, bm)
	if err != nil {
		return nil, err
	}

	if resp.Response.Code != "10000" {
		return nil, errors.New(resp.Response.SubMsg)
	}

	status := "pending"
	switch resp.Response.TradeStatus {
	case "TRADE_SUCCESS", "TRADE_FINISHED":
		status = "paid"
	case "TRADE_CLOSED":
		status = "closed"
	}

	amount, _ := decimal.NewFromString(resp.Response.TotalAmount)

	return &QueryOrderResponse{
		TradeNo:    tradeNo,
		ApiTradeNo: resp.Response.TradeNo,
		Amount:     amount,
		Status:     status,
		PaidAt:     resp.Response.SendPayDate,
	}, nil
}

// Refund 退款
func (a *AlipayAdapter) Refund(ctx context.Context, req *RefundRequest) (*RefundResponse, error) {
	bm := make(gopay.BodyMap)
	bm.Set("out_trade_no", req.TradeNo)
	bm.Set("out_request_no", req.RefundNo)
	bm.Set("refund_amount", req.Amount.StringFixed(2))
	if req.RefundDesc != "" {
		bm.Set("refund_reason", req.RefundDesc)
	}

	resp, err := a.client.TradeRefund(ctx, bm)
	if err != nil {
		return nil, err
	}

	if resp.Response.Code != "10000" {
		return &RefundResponse{
			RefundNo:     req.RefundNo,
			Status:       "failed",
			ErrorMessage: resp.Response.SubMsg,
		}, nil
	}

	return &RefundResponse{
		RefundNo:    req.RefundNo,
		ApiRefundNo: resp.Response.TradeNo,
		Status:      "success",
	}, nil
}

// ParseNotify 解析回调通知
func (a *AlipayAdapter) ParseNotify(ctx context.Context, r *http.Request) (*NotifyResult, error) {
	notifyReq, err := alipay.ParseNotifyToBodyMap(r)
	if err != nil {
		return nil, err
	}

	// 验签
	ok, err := alipay.VerifySign(a.config.PublicKey, notifyReq)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("签名验证失败")
	}

	tradeStatus := notifyReq.Get("trade_status")
	status := "fail"
	if tradeStatus == "TRADE_SUCCESS" || tradeStatus == "TRADE_FINISHED" {
		status = "success"
	}

	amount, _ := decimal.NewFromString(notifyReq.Get("total_amount"))

	return &NotifyResult{
		TradeNo:    notifyReq.Get("out_trade_no"),
		ApiTradeNo: notifyReq.Get("trade_no"),
		Amount:     amount,
		Buyer:      notifyReq.Get("buyer_id"),
		Status:     status,
	}, nil
}

// NotifySuccess 返回成功响应
func (a *AlipayAdapter) NotifySuccess() string {
	return "success"
}

func init() {
	Register("alipay", NewAlipayAdapter)
}
```

**Step 2: 提交**

```bash
git add internal/payment/alipay.go
git commit -m "feat: add alipay payment adapter with gopay"
```

---

## Task 10: 创建微信支付适配器

**Files:**
- Create: `epay-go/internal/payment/wechat.go`

**Step 1: 创建微信支付适配器**

```go
// internal/payment/wechat.go
package payment

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/wechat/v3"
	"github.com/shopspring/decimal"
)

// WechatConfig 微信支付配置
type WechatConfig struct {
	MchID               string `json:"mch_id"`                 // 商户号
	AppID               string `json:"app_id"`                 // 应用ID
	APIv3Key            string `json:"api_v3_key"`             // APIv3密钥
	SerialNo            string `json:"serial_no"`              // 证书序列号
	PrivateKey          string `json:"private_key"`            // 私钥内容
	PlatformSerialNo    string `json:"platform_serial_no"`     // 平台证书序列号
	PlatformCertContent string `json:"platform_cert_content"`  // 平台证书内容
}

// WechatAdapter 微信支付适配器
type WechatAdapter struct {
	client *wechat.ClientV3
	config *WechatConfig
}

// NewWechatAdapter 创建微信支付适配器
func NewWechatAdapter(configJSON json.RawMessage) (PaymentAdapter, error) {
	var cfg WechatConfig
	if err := json.Unmarshal(configJSON, &cfg); err != nil {
		return nil, err
	}

	client, err := wechat.NewClientV3(cfg.MchID, cfg.SerialNo, cfg.APIv3Key, cfg.PrivateKey)
	if err != nil {
		return nil, err
	}

	// 设置平台证书
	if cfg.PlatformCertContent != "" {
		client.SetPlatformCert([]byte(cfg.PlatformCertContent), cfg.PlatformSerialNo)
	}

	return &WechatAdapter{
		client: client,
		config: &cfg,
	}, nil
}

// CreateOrder 创建支付订单
func (w *WechatAdapter) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error) {
	// 金额转为分
	amountFen := req.Amount.Mul(decimal.NewFromInt(100)).IntPart()

	bm := make(gopay.BodyMap)
	bm.Set("appid", w.config.AppID)
	bm.Set("mchid", w.config.MchID)
	bm.Set("description", req.Subject)
	bm.Set("out_trade_no", req.TradeNo)
	bm.Set("notify_url", req.NotifyURL)
	bm.SetBodyMap("amount", func(b gopay.BodyMap) {
		b.Set("total", amountFen)
		b.Set("currency", "CNY")
	})

	switch req.PayMethod {
	case "scan", "qrcode":
		// Native 扫码支付
		resp, err := w.client.V3TransactionNative(ctx, bm)
		if err != nil {
			return nil, err
		}
		if resp.Code != wechat.Success {
			return nil, errors.New(resp.Error)
		}
		return &CreateOrderResponse{
			PayType: "qrcode",
			PayURL:  resp.Response.CodeUrl,
		}, nil

	case "h5", "wap":
		// H5 支付
		bm.SetBodyMap("scene_info", func(b gopay.BodyMap) {
			b.Set("payer_client_ip", req.ClientIP)
			b.SetBodyMap("h5_info", func(h gopay.BodyMap) {
				h.Set("type", "Wap")
			})
		})
		resp, err := w.client.V3TransactionH5(ctx, bm)
		if err != nil {
			return nil, err
		}
		if resp.Code != wechat.Success {
			return nil, errors.New(resp.Error)
		}
		// 拼接 redirect_url
		payURL := resp.Response.H5Url
		if req.ReturnURL != "" {
			payURL += "&redirect_url=" + req.ReturnURL
		}
		return &CreateOrderResponse{
			PayType: "redirect",
			PayURL:  payURL,
		}, nil

	case "jsapi":
		// JSAPI 支付（需要 openid）
		openid := req.Extra["openid"]
		if openid == "" {
			return nil, errors.New("jsapi pay requires openid")
		}
		bm.SetBodyMap("payer", func(b gopay.BodyMap) {
			b.Set("openid", openid)
		})
		resp, err := w.client.V3TransactionJsapi(ctx, bm)
		if err != nil {
			return nil, err
		}
		if resp.Code != wechat.Success {
			return nil, errors.New(resp.Error)
		}
		// 生成 JSAPI 调起参数
		jsapiParams, err := w.client.PaySignOfJSAPI(w.config.AppID, resp.Response.PrepayId)
		if err != nil {
			return nil, err
		}
		paramsJSON, _ := json.Marshal(jsapiParams)
		return &CreateOrderResponse{
			PayType:   "jsapi",
			PayParams: string(paramsJSON),
		}, nil

	default:
		return nil, errors.New("unsupported pay method: " + req.PayMethod)
	}
}

// QueryOrder 查询订单
func (w *WechatAdapter) QueryOrder(ctx context.Context, tradeNo string) (*QueryOrderResponse, error) {
	resp, err := w.client.V3TransactionQueryOrder(ctx, wechat.OutTradeNo, tradeNo)
	if err != nil {
		return nil, err
	}
	if resp.Code != wechat.Success {
		return nil, errors.New(resp.Error)
	}

	status := "pending"
	switch resp.Response.TradeState {
	case "SUCCESS":
		status = "paid"
	case "CLOSED", "REVOKED", "PAYERROR":
		status = "closed"
	case "REFUND":
		status = "refunded"
	}

	// 金额从分转为元
	amount := decimal.NewFromInt(int64(resp.Response.Amount.Total)).Div(decimal.NewFromInt(100))

	return &QueryOrderResponse{
		TradeNo:    tradeNo,
		ApiTradeNo: resp.Response.TransactionId,
		Amount:     amount,
		Status:     status,
		PaidAt:     resp.Response.SuccessTime,
	}, nil
}

// Refund 退款
func (w *WechatAdapter) Refund(ctx context.Context, req *RefundRequest) (*RefundResponse, error) {
	// 金额转为分
	amountFen := req.Amount.Mul(decimal.NewFromInt(100)).IntPart()

	bm := make(gopay.BodyMap)
	bm.Set("out_trade_no", req.TradeNo)
	bm.Set("out_refund_no", req.RefundNo)
	bm.Set("reason", req.RefundDesc)
	bm.SetBodyMap("amount", func(b gopay.BodyMap) {
		b.Set("refund", amountFen)
		b.Set("total", amountFen) // 简化处理，实际应查询原订单金额
		b.Set("currency", "CNY")
	})

	resp, err := w.client.V3Refund(ctx, bm)
	if err != nil {
		return nil, err
	}

	if resp.Code != wechat.Success {
		return &RefundResponse{
			RefundNo:     req.RefundNo,
			Status:       "failed",
			ErrorMessage: resp.Error,
		}, nil
	}

	status := "processing"
	if resp.Response.Status == "SUCCESS" {
		status = "success"
	}

	return &RefundResponse{
		RefundNo:    req.RefundNo,
		ApiRefundNo: resp.Response.RefundId,
		Status:      status,
	}, nil
}

// ParseNotify 解析回调通知
func (w *WechatAdapter) ParseNotify(ctx context.Context, r *http.Request) (*NotifyResult, error) {
	notifyReq, err := wechat.V3ParseNotify(r)
	if err != nil {
		return nil, err
	}

	// 解密回调内容
	result, err := notifyReq.DecryptCipherText(w.config.APIv3Key)
	if err != nil {
		return nil, err
	}

	status := "fail"
	if result.TradeState == "SUCCESS" {
		status = "success"
	}

	// 金额从分转为元
	amount := decimal.NewFromInt(int64(result.Amount.Total)).Div(decimal.NewFromInt(100))

	return &NotifyResult{
		TradeNo:    result.OutTradeNo,
		ApiTradeNo: result.TransactionId,
		Amount:     amount,
		Buyer:      result.Payer.Openid,
		Status:     status,
	}, nil
}

// NotifySuccess 返回成功响应
func (w *WechatAdapter) NotifySuccess() string {
	return `{"code":"SUCCESS","message":"成功"}`
}

func init() {
	Register("wechat", NewWechatAdapter)
}
```

**Step 2: 提交**

```bash
git add internal/payment/wechat.go
git commit -m "feat: add wechat payment adapter with gopay v3"
```

---

## Task 11: 创建订单服务

**Files:**
- Create: `epay-go/internal/service/order.go`

**Step 1: 创建订单服务**

```go
// internal/service/order.go
package service

import (
	"context"
	"errors"

	"github.com/example/epay-go/internal/database"
	"github.com/example/epay-go/internal/model"
	"github.com/example/epay-go/internal/payment"
	"github.com/example/epay-go/internal/repository"
	"github.com/example/epay-go/pkg/utils"
	"github.com/shopspring/decimal"
)

type OrderService struct {
	orderRepo    *repository.OrderRepository
	channelRepo  *repository.ChannelRepository
	merchantRepo *repository.MerchantRepository
	recordRepo   *repository.BalanceRecordRepository
}

func NewOrderService() *OrderService {
	return &OrderService{
		orderRepo:    repository.NewOrderRepository(),
		channelRepo:  repository.NewChannelRepository(),
		merchantRepo: repository.NewMerchantRepository(),
		recordRepo:   repository.NewBalanceRecordRepository(),
	}
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	MerchantID  int64           `json:"-"`
	OutTradeNo  string          `json:"out_trade_no" binding:"required"`
	Amount      decimal.Decimal `json:"money" binding:"required"`
	Name        string          `json:"name" binding:"required"`
	PayType     string          `json:"type" binding:"required"` // alipay, wxpay
	NotifyURL   string          `json:"notify_url" binding:"required,url"`
	ReturnURL   string          `json:"return_url" binding:"omitempty,url"`
	ClientIP    string          `json:"-"`
	PayMethod   string          `json:"pay_method"` // scan, h5, jsapi, web
	Extra       map[string]string `json:"extra"`
}

// CreateOrderResponse 创建订单响应
type CreateOrderResponse struct {
	TradeNo   string `json:"trade_no"`
	PayType   string `json:"pay_type"`
	PayURL    string `json:"pay_url,omitempty"`
	PayParams string `json:"pay_params,omitempty"`
}

// Create 创建订单
func (s *OrderService) Create(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error) {
	// 检查商户订单号是否重复
	existOrder, _ := s.orderRepo.GetByOutTradeNo(req.MerchantID, req.OutTradeNo)
	if existOrder != nil {
		return nil, errors.New("商户订单号已存在")
	}

	// 获取可用通道
	channel, err := s.channelRepo.GetAvailableByPayType(req.PayType)
	if err != nil {
		return nil, errors.New("暂无可用的支付通道")
	}

	// 创建支付适配器
	adapter, err := payment.NewAdapter(channel.Plugin, channel.Config)
	if err != nil {
		return nil, errors.New("支付通道配置错误")
	}

	// 生成订单号
	tradeNo := utils.GenerateTradeNo()

	// 计算手续费
	fee := req.Amount.Mul(channel.Rate).Round(2)
	realAmount := req.Amount

	// 创建订单记录
	order := &model.Order{
		TradeNo:      tradeNo,
		OutTradeNo:   req.OutTradeNo,
		MerchantID:   req.MerchantID,
		ChannelID:    channel.ID,
		PayType:      req.PayType,
		Amount:       req.Amount,
		RealAmount:   realAmount,
		Fee:          fee,
		Name:         req.Name,
		NotifyURL:    req.NotifyURL,
		ReturnURL:    req.ReturnURL,
		ClientIP:     req.ClientIP,
		Status:       model.OrderStatusUnpaid,
		NotifyStatus: model.NotifyStatusPending,
	}

	if err := s.orderRepo.Create(order); err != nil {
		return nil, err
	}

	// 调用支付接口
	payMethod := req.PayMethod
	if payMethod == "" {
		payMethod = "scan" // 默认扫码
	}

	payReq := &payment.CreateOrderRequest{
		TradeNo:   tradeNo,
		Amount:    realAmount,
		Subject:   req.Name,
		ClientIP:  req.ClientIP,
		NotifyURL: req.NotifyURL,
		ReturnURL: req.ReturnURL,
		PayMethod: payMethod,
		Extra:     req.Extra,
	}

	payResp, err := adapter.CreateOrder(ctx, payReq)
	if err != nil {
		return nil, err
	}

	return &CreateOrderResponse{
		TradeNo:   tradeNo,
		PayType:   payResp.PayType,
		PayURL:    payResp.PayURL,
		PayParams: payResp.PayParams,
	}, nil
}

// GetByTradeNo 根据订单号获取订单
func (s *OrderService) GetByTradeNo(tradeNo string) (*model.Order, error) {
	return s.orderRepo.GetByTradeNo(tradeNo)
}

// List 分页查询订单
func (s *OrderService) List(page, pageSize int, merchantID *int64, status *int8) ([]model.Order, int64, error) {
	return s.orderRepo.List(page, pageSize, merchantID, status)
}

// ProcessPayNotify 处理支付回调
func (s *OrderService) ProcessPayNotify(tradeNo, apiTradeNo, buyer string, amount decimal.Decimal) error {
	order, err := s.orderRepo.GetByTradeNo(tradeNo)
	if err != nil {
		return errors.New("订单不存在")
	}

	if order.Status != model.OrderStatusUnpaid {
		return nil // 订单已处理，跳过
	}

	// 验证金额
	if !order.Amount.Equal(amount) {
		return errors.New("支付金额不匹配")
	}

	// 开启事务
	tx := database.Get().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新订单状态
	if err := s.orderRepo.UpdatePayInfo(tradeNo, apiTradeNo, buyer); err != nil {
		tx.Rollback()
		return err
	}

	// 获取商户信息
	merchant, err := s.merchantRepo.GetByID(order.MerchantID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 计算商户收入（订单金额 - 手续费）
	income := order.Amount.Sub(order.Fee)
	newBalance := merchant.Balance.Add(income)

	// 更新商户余额
	if err := s.merchantRepo.UpdateBalance(tx, merchant.ID, income.InexactFloat64()); err != nil {
		tx.Rollback()
		return err
	}

	// 添加资金记录
	if err := repository.AddBalanceRecord(tx, merchant.ID, model.RecordActionIncome, income, merchant.Balance, newBalance, "order_income", tradeNo); err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

// GetTodayStats 获取今日统计
func (s *OrderService) GetTodayStats(merchantID *int64) (int64, decimal.Decimal, error) {
	return s.orderRepo.GetTodayStats(merchantID)
}
```

**Step 2: 提交**

```bash
git add internal/service/order.go
git commit -m "feat: add order service with create, query, notify processing"
```

---

## Task 12: 创建结算服务

**Files:**
- Create: `epay-go/internal/service/settlement.go`

**Step 1: 创建结算服务**

```go
// internal/service/settlement.go
package service

import (
	"errors"

	"github.com/example/epay-go/internal/database"
	"github.com/example/epay-go/internal/model"
	"github.com/example/epay-go/internal/repository"
	"github.com/example/epay-go/pkg/utils"
	"github.com/shopspring/decimal"
)

type SettlementService struct {
	settleRepo   *repository.SettlementRepository
	merchantRepo *repository.MerchantRepository
}

func NewSettlementService() *SettlementService {
	return &SettlementService{
		settleRepo:   repository.NewSettlementRepository(),
		merchantRepo: repository.NewMerchantRepository(),
	}
}

// ApplyRequest 申请结算请求
type ApplyRequest struct {
	MerchantID  int64           `json:"-"`
	Amount      decimal.Decimal `json:"amount" binding:"required"`
	AccountType string          `json:"account_type" binding:"required,oneof=alipay bank"`
	AccountNo   string          `json:"account_no" binding:"required"`
	AccountName string          `json:"account_name" binding:"required"`
}

// Apply 申请结算
func (s *SettlementService) Apply(req *ApplyRequest) (*model.Settlement, error) {
	// 检查是否有待处理的结算
	hasPending, err := s.settleRepo.HasPendingSettlement(req.MerchantID)
	if err != nil {
		return nil, err
	}
	if hasPending {
		return nil, errors.New("您有待处理的结算申请，请等待处理完成")
	}

	// 获取商户信息
	merchant, err := s.merchantRepo.GetByID(req.MerchantID)
	if err != nil {
		return nil, err
	}

	// 检查余额是否充足
	if merchant.Balance.LessThan(req.Amount) {
		return nil, errors.New("余额不足")
	}

	// 最小结算金额检查（假设最小10元）
	minAmount := decimal.NewFromInt(10)
	if req.Amount.LessThan(minAmount) {
		return nil, errors.New("最小结算金额为10元")
	}

	// 计算手续费（假设2%）
	feeRate := decimal.NewFromFloat(0.02)
	fee := req.Amount.Mul(feeRate).Round(2)
	actualAmount := req.Amount.Sub(fee)

	// 创建结算记录
	settlement := &model.Settlement{
		SettleNo:     utils.GenerateSettleNo(),
		MerchantID:   req.MerchantID,
		Amount:       req.Amount,
		Fee:          fee,
		ActualAmount: actualAmount,
		AccountType:  req.AccountType,
		AccountNo:    req.AccountNo,
		AccountName:  req.AccountName,
		Status:       model.SettleStatusPending,
	}

	// 开启事务
	tx := database.Get().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 冻结余额
	newBalance := merchant.Balance.Sub(req.Amount)
	newFrozen := merchant.FrozenBalance.Add(req.Amount)

	if err := tx.Model(&model.Merchant{}).Where("id = ?", req.MerchantID).Updates(map[string]interface{}{
		"balance":        newBalance,
		"frozen_balance": newFrozen,
	}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 添加资金记录
	if err := repository.AddBalanceRecord(tx, req.MerchantID, model.RecordActionExpense, req.Amount, merchant.Balance, newBalance, "settle_freeze", settlement.SettleNo); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 创建结算记录
	if err := tx.Create(settlement).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	return settlement, nil
}

// Approve 审核通过
func (s *SettlementService) Approve(id int64) error {
	settlement, err := s.settleRepo.GetByID(id)
	if err != nil {
		return err
	}

	if settlement.Status != model.SettleStatusPending {
		return errors.New("结算状态不正确")
	}

	// 更新状态为处理中
	return s.settleRepo.UpdateStatus(id, model.SettleStatusProcessing, "审核通过，处理中")
}

// Complete 完成结算
func (s *SettlementService) Complete(id int64) error {
	settlement, err := s.settleRepo.GetByID(id)
	if err != nil {
		return err
	}

	if settlement.Status != model.SettleStatusProcessing {
		return errors.New("结算状态不正确")
	}

	// 开启事务
	tx := database.Get().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 扣除冻结余额
	if err := tx.Model(&model.Merchant{}).Where("id = ?", settlement.MerchantID).
		Update("frozen_balance", gorm.Expr("frozen_balance - ?", settlement.Amount)).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 更新结算状态
	if err := s.settleRepo.UpdateStatus(id, model.SettleStatusCompleted, "结算完成"); err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

// Reject 驳回结算
func (s *SettlementService) Reject(id int64, remark string) error {
	settlement, err := s.settleRepo.GetByID(id)
	if err != nil {
		return err
	}

	if settlement.Status != model.SettleStatusPending {
		return errors.New("结算状态不正确")
	}

	// 开启事务
	tx := database.Get().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取商户
	merchant, err := s.merchantRepo.GetByID(settlement.MerchantID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 解冻余额
	newBalance := merchant.Balance.Add(settlement.Amount)
	newFrozen := merchant.FrozenBalance.Sub(settlement.Amount)

	if err := tx.Model(&model.Merchant{}).Where("id = ?", settlement.MerchantID).Updates(map[string]interface{}{
		"balance":        newBalance,
		"frozen_balance": newFrozen,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 添加资金记录
	if err := repository.AddBalanceRecord(tx, settlement.MerchantID, model.RecordActionIncome, settlement.Amount, merchant.Balance, newBalance, "settle_unfreeze", settlement.SettleNo); err != nil {
		tx.Rollback()
		return err
	}

	// 更新结算状态
	if err := s.settleRepo.UpdateStatus(id, model.SettleStatusRejected, remark); err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

// List 分页查询结算列表
func (s *SettlementService) List(page, pageSize int, merchantID *int64, status *int8) ([]model.Settlement, int64, error) {
	return s.settleRepo.List(page, pageSize, merchantID, status)
}

// GetByID 根据ID获取结算记录
func (s *SettlementService) GetByID(id int64) (*model.Settlement, error) {
	return s.settleRepo.GetByID(id)
}
```

需要在顶部添加 gorm 导入：

```go
import (
	"errors"

	"github.com/example/epay-go/internal/database"
	"github.com/example/epay-go/internal/model"
	"github.com/example/epay-go/internal/repository"
	"github.com/example/epay-go/pkg/utils"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)
```

**Step 2: 提交**

```bash
git add internal/service/settlement.go
git commit -m "feat: add settlement service with apply, approve, reject, complete"
```

---

## Task 13: 创建异步通知服务

**Files:**
- Create: `epay-go/internal/service/notify.go`

**Step 1: 创建通知服务**

```go
// internal/service/notify.go
package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/example/epay-go/internal/model"
	"github.com/example/epay-go/internal/repository"
)

type NotifyService struct {
	orderRepo    *repository.OrderRepository
	merchantRepo *repository.MerchantRepository
	httpClient   *http.Client
}

func NewNotifyService() *NotifyService {
	return &NotifyService{
		orderRepo:    repository.NewOrderRepository(),
		merchantRepo: repository.NewMerchantRepository(),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// NotifyRetryIntervals 通知重试间隔
var NotifyRetryIntervals = []time.Duration{
	0,                // 立即
	1 * time.Minute,  // 1分钟后
	3 * time.Minute,  // 3分钟后
	20 * time.Minute, // 20分钟后
	1 * time.Hour,    // 1小时后
	2 * time.Hour,    // 2小时后
}

// SendNotify 发送回调通知
func (s *NotifyService) SendNotify(order *model.Order) error {
	// 获取商户信息
	merchant, err := s.merchantRepo.GetByID(order.MerchantID)
	if err != nil {
		return err
	}

	// 构建通知参数
	params := s.buildNotifyParams(order, merchant)

	// 发送请求
	success := s.doNotify(order.NotifyURL, params)

	// 更新通知状态
	if success {
		return s.orderRepo.UpdateNotifyStatus(order.TradeNo, model.NotifyStatusSuccess, nil)
	}

	// 通知失败，计算下次重试时间
	nextCount := order.NotifyCount + 1
	if nextCount >= len(NotifyRetryIntervals) {
		// 重试次数用尽
		return s.orderRepo.UpdateNotifyStatus(order.TradeNo, model.NotifyStatusFailed, nil)
	}

	nextTime := time.Now().Add(NotifyRetryIntervals[nextCount])
	return s.orderRepo.UpdateNotifyStatus(order.TradeNo, model.NotifyStatusSending, &nextTime)
}

// buildNotifyParams 构建通知参数
func (s *NotifyService) buildNotifyParams(order *model.Order, merchant *model.Merchant) url.Values {
	params := url.Values{}
	params.Set("pid", fmt.Sprintf("%d", order.MerchantID))
	params.Set("trade_no", order.TradeNo)
	params.Set("out_trade_no", order.OutTradeNo)
	params.Set("type", order.PayType)
	params.Set("name", order.Name)
	params.Set("money", order.Amount.String())
	params.Set("trade_status", "TRADE_SUCCESS")

	if order.ApiTradeNo != "" {
		params.Set("api_trade_no", order.ApiTradeNo)
	}
	if order.Buyer != "" {
		params.Set("buyer", order.Buyer)
	}

	// 生成签名
	sign := s.generateSign(params, merchant.ApiKey)
	params.Set("sign", sign)
	params.Set("sign_type", "MD5")

	return params
}

// generateSign 生成MD5签名（与原epay兼容）
func (s *NotifyService) generateSign(params url.Values, key string) string {
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
	return hex.EncodeToString(hash[:])
}

// doNotify 执行通知请求
func (s *NotifyService) doNotify(notifyURL string, params url.Values) bool {
	// 构建完整URL
	fullURL := notifyURL
	if strings.Contains(notifyURL, "?") {
		fullURL += "&" + params.Encode()
	} else {
		fullURL += "?" + params.Encode()
	}

	resp, err := s.httpClient.Get(fullURL)
	if err != nil {
		log.Printf("Notify request failed: %v", err)
		return false
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	response := strings.ToLower(strings.TrimSpace(string(body)))

	// 检查响应是否为 success
	return response == "success"
}

// StartNotifyWorker 启动通知工作协程
func (s *NotifyService) StartNotifyWorker(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Notify worker stopped")
			return
		case <-ticker.C:
			s.processNotifyQueue()
		}
	}
}

// processNotifyQueue 处理通知队列
func (s *NotifyService) processNotifyQueue() {
	orders, err := s.orderRepo.GetPendingNotifyOrders(50)
	if err != nil {
		log.Printf("Get pending notify orders failed: %v", err)
		return
	}

	for _, order := range orders {
		if err := s.SendNotify(&order); err != nil {
			log.Printf("Send notify failed for %s: %v", order.TradeNo, err)
		}
	}
}
```

**Step 2: 提交**

```bash
git add internal/service/notify.go
git commit -m "feat: add async notify service with retry mechanism"
```

---

## Task 14: 更新主程序入口，集成所有服务

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

	// TODO: 在阶段三注册具体路由

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
git commit -m "feat: integrate all services and notify worker into main"
```

---

## 阶段二完成检查清单

- [ ] 通用工具函数（密码、订单号生成）
- [ ] Repository 层（商户、管理员、通道、订单、结算、资金记录）
- [ ] 商户服务（注册、登录、信息管理）
- [ ] 管理员服务（登录、密码管理）
- [ ] 通道服务（CRUD）
- [ ] 支付适配器接口和工厂
- [ ] 支付宝适配器（gopay）
- [ ] 微信支付适配器（gopay v3）
- [ ] 订单服务（创建、查询、回调处理）
- [ ] 结算服务（申请、审核、驳回、完成）
- [ ] 异步通知服务（重试机制）
- [ ] 主程序集成

---

**下一阶段：** 阶段三将实现 API Handler 层，包括管理后台 API、商户端 API 和对外支付 API。
