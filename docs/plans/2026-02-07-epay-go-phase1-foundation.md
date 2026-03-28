# EPay Go 重构 - 阶段一：基础框架

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 搭建 Go + Vue3 项目基础框架，包括项目结构、数据库连接、Redis 连接、JWT 认证中间件和统一响应格式。

**Architecture:** 采用 Gin 框架构建 RESTful API，使用 GORM 作为 ORM，Viper 管理配置，JWT 进行身份认证。前端使用 Vue3 + Vite + Arco Design。

**Tech Stack:** Go 1.22+, Gin, GORM, PostgreSQL, Redis, Viper, JWT, Vue3, Vite, Arco Design, Docker

---

## Task 1: 初始化 Go 项目

**Files:**
- Create: `epay-go/go.mod`
- Create: `epay-go/cmd/server/main.go`
- Create: `epay-go/.gitignore`

**Step 1: 创建项目目录结构**

```bash
mkdir -p d:/project/payment/epay-go
cd d:/project/payment/epay-go
```

**Step 2: 初始化 Go 模块**

Run: `go mod init github.com/example/epay-go`

**Step 3: 创建 .gitignore**

```gitignore
# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
epay-server

# Test binary
*.test

# Output of the go coverage tool
*.out

# Dependency directories
vendor/

# IDE
.idea/
.vscode/
*.swp
*.swo

# Environment
.env
.env.local
*.local

# Logs
*.log
logs/

# OS
.DS_Store
Thumbs.db

# Build
dist/
build/

# Frontend
web/node_modules/
web/dist/
```

**Step 4: 创建入口文件骨架**

```go
// cmd/server/main.go
package main

import (
	"fmt"
)

func main() {
	fmt.Println("EPay Go Server Starting...")
}
```

**Step 5: 验证项目可运行**

Run: `go run cmd/server/main.go`
Expected: `EPay Go Server Starting...`

**Step 6: 提交**

```bash
git init
git add .
git commit -m "chore: initialize go project structure"
```

---

## Task 2: 添加核心依赖

**Files:**
- Modify: `epay-go/go.mod`

**Step 1: 安装核心依赖**

```bash
cd d:/project/payment/epay-go
go get github.com/gin-gonic/gin@v1.9.1
go get gorm.io/gorm@v1.25.7
go get gorm.io/driver/postgres@v1.5.7
go get github.com/redis/go-redis/v9@v9.5.1
go get github.com/spf13/viper@v1.18.2
go get github.com/golang-jwt/jwt/v5@v5.2.1
go get github.com/shopspring/decimal@v1.3.1
go get github.com/go-pay/gopay@v1.5.102
go get golang.org/x/crypto@v0.21.0
go get github.com/google/uuid@v1.6.0
```

**Step 2: 整理依赖**

Run: `go mod tidy`

**Step 3: 验证依赖安装成功**

Run: `go mod verify`
Expected: `all modules verified`

**Step 4: 提交**

```bash
git add go.mod go.sum
git commit -m "chore: add core dependencies"
```

---

## Task 3: 创建配置管理

**Files:**
- Create: `epay-go/internal/config/config.go`
- Create: `epay-go/config.yaml`
- Create: `epay-go/config.example.yaml`

**Step 1: 创建配置结构体**

```go
// internal/config/config.go
package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug, release, test
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	ExpireHour int    `mapstructure:"expire_hour"`
}

var Cfg *Config

func Load(path string) error {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	// 环境变量覆盖
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	Cfg = &Config{}
	if err := viper.Unmarshal(Cfg); err != nil {
		return err
	}

	return nil
}

func Get() *Config {
	return Cfg
}
```

**Step 2: 创建示例配置文件**

```yaml
# config.example.yaml
server:
  port: 8080
  mode: debug  # debug, release, test

database:
  host: localhost
  port: 5432
  user: epay
  password: your_password_here
  dbname: epay
  sslmode: disable

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

jwt:
  secret: your_jwt_secret_here_change_in_production
  expire_hour: 168  # 7 days
```

**Step 3: 创建本地配置文件**

```yaml
# config.yaml
server:
  port: 8080
  mode: debug

database:
  host: localhost
  port: 5432
  user: epay
  password: epay123
  dbname: epay
  sslmode: disable

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

jwt:
  secret: dev_jwt_secret_key_12345678
  expire_hour: 168
```

**Step 4: 更新 .gitignore 排除 config.yaml**

在 .gitignore 中添加:
```
config.yaml
```

**Step 5: 验证配置加载**

临时修改 main.go 测试:
```go
// cmd/server/main.go
package main

import (
	"fmt"
	"log"

	"github.com/example/epay-go/internal/config"
)

func main() {
	if err := config.Load("config.yaml"); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Printf("Server will run on port: %d\n", config.Get().Server.Port)
}
```

Run: `go run cmd/server/main.go`
Expected: `Server will run on port: 8080`

**Step 6: 提交**

```bash
git add .
git commit -m "feat: add configuration management with viper"
```

---

## Task 4: 创建数据库连接

**Files:**
- Create: `epay-go/internal/database/database.go`

**Step 1: 创建数据库连接模块**

```go
// internal/database/database.go
package database

import (
	"fmt"
	"log"
	"time"

	"github.com/example/epay-go/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init() error {
	cfg := config.Get().Database

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	var gormConfig *gorm.Config
	if config.Get().Server.Mode == "debug" {
		gormConfig = &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		}
	} else {
		gormConfig = &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		}
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 连接池配置
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Database connected successfully")
	return nil
}

func Get() *gorm.DB {
	return DB
}

func Close() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
```

**Step 2: 提交**

```bash
git add internal/database/database.go
git commit -m "feat: add postgresql database connection with gorm"
```

---

## Task 5: 创建 Redis 连接

**Files:**
- Create: `epay-go/internal/cache/redis.go`

**Step 1: 创建 Redis 连接模块**

```go
// internal/cache/redis.go
package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/example/epay-go/internal/config"
	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func Init() error {
	cfg := config.Get().Redis

	RDB = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RDB.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect redis: %w", err)
	}

	log.Println("Redis connected successfully")
	return nil
}

func Get() *redis.Client {
	return RDB
}

func Close() error {
	return RDB.Close()
}
```

**Step 2: 提交**

```bash
git add internal/cache/redis.go
git commit -m "feat: add redis connection"
```

---

## Task 6: 创建统一响应格式

**Files:**
- Create: `epay-go/pkg/response/response.go`

**Step 1: 创建响应工具**

```go
// pkg/response/response.go
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 错误码定义
const (
	CodeSuccess       = 0
	CodeParamError    = 1001
	CodeUnauthorized  = 1002
	CodeForbidden     = 1003
	CodeNotFound      = 1004
	CodeServerError   = 5000
	CodeDatabaseError = 5001
)

// Response 统一响应结构
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// PageData 分页数据结构
type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: CodeSuccess,
		Msg:  "success",
		Data: data,
	})
}

// SuccessWithMessage 带消息的成功响应
func SuccessWithMessage(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: CodeSuccess,
		Msg:  msg,
		Data: data,
	})
}

// SuccessPage 分页成功响应
func SuccessPage(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, Response{
		Code: CodeSuccess,
		Msg:  "success",
		Data: PageData{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
	})
}

// ParamError 参数错误
func ParamError(c *gin.Context, msg string) {
	Error(c, CodeParamError, msg)
}

// Unauthorized 未授权
func Unauthorized(c *gin.Context, msg string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code: CodeUnauthorized,
		Msg:  msg,
	})
}

// Forbidden 禁止访问
func Forbidden(c *gin.Context, msg string) {
	c.JSON(http.StatusForbidden, Response{
		Code: CodeForbidden,
		Msg:  msg,
	})
}

// NotFound 未找到
func NotFound(c *gin.Context, msg string) {
	c.JSON(http.StatusNotFound, Response{
		Code: CodeNotFound,
		Msg:  msg,
	})
}

// ServerError 服务器错误
func ServerError(c *gin.Context, msg string) {
	c.JSON(http.StatusInternalServerError, Response{
		Code: CodeServerError,
		Msg:  msg,
	})
}
```

**Step 2: 提交**

```bash
git add pkg/response/response.go
git commit -m "feat: add unified response format"
```

---

## Task 7: 创建 JWT 认证中间件

**Files:**
- Create: `epay-go/internal/middleware/auth.go`
- Create: `epay-go/pkg/jwt/jwt.go`

**Step 1: 创建 JWT 工具包**

```go
// pkg/jwt/jwt.go
package jwt

import (
	"errors"
	"time"

	"github.com/example/epay-go/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

type TokenType string

const (
	TokenTypeAdmin    TokenType = "admin"
	TokenTypeMerchant TokenType = "merchant"
)

type Claims struct {
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

// GenerateToken 生成 JWT Token
func GenerateToken(userID int64, username string, tokenType TokenType) (string, error) {
	cfg := config.Get().JWT
	expireTime := time.Now().Add(time.Duration(cfg.ExpireHour) * time.Hour)

	claims := Claims{
		UserID:    userID,
		Username:  username,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "epay-go",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

// ParseToken 解析 JWT Token
func ParseToken(tokenString string) (*Claims, error) {
	cfg := config.Get().JWT

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
```

**Step 2: 创建认证中间件**

```go
// internal/middleware/auth.go
package middleware

import (
	"strings"

	"github.com/example/epay-go/pkg/jwt"
	"github.com/example/epay-go/pkg/response"
	"github.com/gin-gonic/gin"
)

const (
	ContextKeyUserID    = "user_id"
	ContextKeyUsername  = "username"
	ContextKeyTokenType = "token_type"
)

// JWTAuth JWT 认证中间件
func JWTAuth(requiredType jwt.TokenType) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}

		// Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "Token 格式错误")
			c.Abort()
			return
		}

		claims, err := jwt.ParseToken(parts[1])
		if err != nil {
			response.Unauthorized(c, "Token 无效或已过期")
			c.Abort()
			return
		}

		// 验证 token 类型
		if claims.TokenType != requiredType {
			response.Forbidden(c, "无权访问")
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyUsername, claims.Username)
		c.Set(ContextKeyTokenType, claims.TokenType)

		c.Next()
	}
}

// GetUserID 从上下文获取用户ID
func GetUserID(c *gin.Context) int64 {
	if id, exists := c.Get(ContextKeyUserID); exists {
		return id.(int64)
	}
	return 0
}

// GetUsername 从上下文获取用户名
func GetUsername(c *gin.Context) string {
	if name, exists := c.Get(ContextKeyUsername); exists {
		return name.(string)
	}
	return ""
}
```

**Step 3: 提交**

```bash
git add pkg/jwt/jwt.go internal/middleware/auth.go
git commit -m "feat: add jwt token generation and auth middleware"
```

---

## Task 8: 创建其他常用中间件

**Files:**
- Create: `epay-go/internal/middleware/cors.go`
- Create: `epay-go/internal/middleware/logger.go`
- Create: `epay-go/internal/middleware/recovery.go`

**Step 1: 创建 CORS 中间件**

```go
// internal/middleware/cors.go
package middleware

import (
	"github.com/gin-gonic/gin"
)

// Cors 跨域中间件
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
```

**Step 2: 创建日志中间件**

```go
// internal/middleware/logger.go
package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger 请求日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		if query != "" {
			path = path + "?" + query
		}

		log.Printf("[GIN] %3d | %13v | %15s | %-7s %s",
			status, latency, clientIP, method, path)
	}
}
```

**Step 3: 创建 Recovery 中间件**

```go
// internal/middleware/recovery.go
package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/example/epay-go/pkg/response"
	"github.com/gin-gonic/gin"
)

// Recovery panic 恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC] %v\n%s", err, debug.Stack())
				c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
					Code: response.CodeServerError,
					Msg:  "服务器内部错误",
				})
			}
		}()
		c.Next()
	}
}
```

**Step 4: 提交**

```bash
git add internal/middleware/
git commit -m "feat: add cors, logger and recovery middleware"
```

---

## Task 9: 创建数据模型

**Files:**
- Create: `epay-go/internal/model/base.go`
- Create: `epay-go/internal/model/admin.go`
- Create: `epay-go/internal/model/merchant.go`
- Create: `epay-go/internal/model/channel.go`
- Create: `epay-go/internal/model/order.go`
- Create: `epay-go/internal/model/settlement.go`
- Create: `epay-go/internal/model/record.go`
- Create: `epay-go/internal/model/config.go`

**Step 1: 创建基础模型**

```go
// internal/model/base.go
package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 基础模型
type BaseModel struct {
	ID        int64          `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
```

**Step 2: 创建管理员模型**

```go
// internal/model/admin.go
package model

import "time"

// Admin 管理员
type Admin struct {
	BaseModel
	Username    string     `gorm:"size:64;uniqueIndex;not null" json:"username"`
	Password    string     `gorm:"size:128;not null" json:"-"`
	Role        string     `gorm:"size:20;default:admin" json:"role"` // super, admin
	LastLoginAt *time.Time `json:"last_login_at"`
}

func (Admin) TableName() string {
	return "admins"
}
```

**Step 3: 创建商户模型**

```go
// internal/model/merchant.go
package model

import (
	"github.com/shopspring/decimal"
)

// Merchant 商户
type Merchant struct {
	BaseModel
	Username      string          `gorm:"size:64;uniqueIndex;not null" json:"username"`
	Password      string          `gorm:"size:128;not null" json:"-"`
	Email         string          `gorm:"size:128" json:"email"`
	Phone         string          `gorm:"size:20" json:"phone"`
	ApiKey        string          `gorm:"size:64;uniqueIndex;not null" json:"api_key"`
	PublicKey     string          `gorm:"type:text" json:"-"`
	PrivateKey    string          `gorm:"type:text" json:"-"`
	Balance       decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"balance"`
	FrozenBalance decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"frozen_balance"`
	Status        int8            `gorm:"default:1" json:"status"` // 0禁用 1正常
}

func (Merchant) TableName() string {
	return "merchants"
}
```

**Step 4: 创建支付通道模型**

```go
// internal/model/channel.go
package model

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

// Channel 支付通道
type Channel struct {
	BaseModel
	Name       string          `gorm:"size:64;not null" json:"name"`
	Plugin     string          `gorm:"size:32;not null" json:"plugin"` // alipay, wechat, paypal...
	PayTypes   string          `gorm:"size:255" json:"pay_types"`      // 支持的支付方式，逗号分隔
	Config     json.RawMessage `gorm:"type:jsonb" json:"config"`       // 通道配置
	Rate       decimal.Decimal `gorm:"type:decimal(5,4);default:0" json:"rate"`
	DailyLimit decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"daily_limit"`
	Status     int8            `gorm:"default:1" json:"status"` // 0禁用 1启用
	Sort       int             `gorm:"default:0" json:"sort"`
}

func (Channel) TableName() string {
	return "channels"
}
```

**Step 5: 创建订单模型**

```go
// internal/model/order.go
package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// Order 订单
type Order struct {
	BaseModel
	TradeNo      string          `gorm:"size:32;uniqueIndex;not null" json:"trade_no"`
	OutTradeNo   string          `gorm:"size:64;not null" json:"out_trade_no"`
	MerchantID   int64           `gorm:"index;not null" json:"merchant_id"`
	ChannelID    int64           `gorm:"index" json:"channel_id"`
	PayType      string          `gorm:"size:20" json:"pay_type"`
	Amount       decimal.Decimal `gorm:"type:decimal(12,2);not null" json:"amount"`
	RealAmount   decimal.Decimal `gorm:"type:decimal(12,2)" json:"real_amount"`
	Fee          decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"fee"`
	Name         string          `gorm:"size:255" json:"name"`
	NotifyURL    string          `gorm:"size:512" json:"notify_url"`
	ReturnURL    string          `gorm:"size:512" json:"return_url"`
	ApiTradeNo   string          `gorm:"size:64" json:"api_trade_no"`
	Buyer        string          `gorm:"size:64" json:"buyer"`
	ClientIP     string          `gorm:"size:45" json:"client_ip"`
	Status       int8            `gorm:"default:0;index" json:"status"` // 0未支付 1已支付 2已退款
	NotifyStatus int8            `gorm:"default:0" json:"notify_status"` // 0未通知 1通知中 2已通知
	NotifyCount  int             `gorm:"default:0" json:"notify_count"`
	NextNotifyAt *time.Time      `json:"next_notify_at"`
	PaidAt       *time.Time      `json:"paid_at"`

	// 关联
	Merchant *Merchant `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
	Channel  *Channel  `gorm:"foreignKey:ChannelID" json:"channel,omitempty"`
}

func (Order) TableName() string {
	return "orders"
}

// 订单状态常量
const (
	OrderStatusUnpaid  = 0
	OrderStatusPaid    = 1
	OrderStatusRefund  = 2
)

// 通知状态常量
const (
	NotifyStatusPending = 0
	NotifyStatusSending = 1
	NotifyStatusSuccess = 2
	NotifyStatusFailed  = 3
)
```

**Step 6: 创建结算模型**

```go
// internal/model/settlement.go
package model

import (
	"github.com/shopspring/decimal"
)

// Settlement 结算记录
type Settlement struct {
	BaseModel
	SettleNo     string          `gorm:"size:32;uniqueIndex;not null" json:"settle_no"`
	MerchantID   int64           `gorm:"index;not null" json:"merchant_id"`
	Amount       decimal.Decimal `gorm:"type:decimal(12,2);not null" json:"amount"`
	Fee          decimal.Decimal `gorm:"type:decimal(12,2);default:0" json:"fee"`
	ActualAmount decimal.Decimal `gorm:"type:decimal(12,2)" json:"actual_amount"`
	AccountType  string          `gorm:"size:20" json:"account_type"` // alipay, bank
	AccountNo    string          `gorm:"size:64" json:"account_no"`
	AccountName  string          `gorm:"size:64" json:"account_name"`
	Status       int8            `gorm:"default:0" json:"status"` // 0待审核 1处理中 2已完成 3已驳回
	Remark       string          `gorm:"size:255" json:"remark"`

	// 关联
	Merchant *Merchant `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
}

func (Settlement) TableName() string {
	return "settlements"
}

// 结算状态常量
const (
	SettleStatusPending   = 0
	SettleStatusProcessing = 1
	SettleStatusCompleted = 2
	SettleStatusRejected  = 3
)
```

**Step 7: 创建资金记录模型**

```go
// internal/model/record.go
package model

import (
	"github.com/shopspring/decimal"
)

// BalanceRecord 资金变动记录
type BalanceRecord struct {
	BaseModel
	MerchantID    int64           `gorm:"index;not null" json:"merchant_id"`
	Action        int8            `gorm:"not null" json:"action"` // 1收入 2支出
	Amount        decimal.Decimal `gorm:"type:decimal(12,2);not null" json:"amount"`
	BeforeBalance decimal.Decimal `gorm:"type:decimal(12,2)" json:"before_balance"`
	AfterBalance  decimal.Decimal `gorm:"type:decimal(12,2)" json:"after_balance"`
	Type          string          `gorm:"size:32" json:"type"` // order_income, fee, settle, refund
	TradeNo       string          `gorm:"size:64" json:"trade_no"`

	// 关联
	Merchant *Merchant `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
}

func (BalanceRecord) TableName() string {
	return "balance_records"
}

// 资金变动类型常量
const (
	RecordActionIncome  = 1
	RecordActionExpense = 2
)
```

**Step 8: 创建系统配置模型**

```go
// internal/model/config.go
package model

import "time"

// Config 系统配置
type Config struct {
	Key       string    `gorm:"size:64;primarykey" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Config) TableName() string {
	return "configs"
}
```

**Step 9: 提交**

```bash
git add internal/model/
git commit -m "feat: add all data models"
```

---

## Task 10: 创建数据库迁移和主程序入口

**Files:**
- Modify: `epay-go/cmd/server/main.go`
- Create: `epay-go/internal/database/migrate.go`

**Step 1: 创建迁移函数**

```go
// internal/database/migrate.go
package database

import (
	"log"

	"github.com/example/epay-go/internal/model"
)

// Migrate 自动迁移数据库表
func Migrate() error {
	log.Println("Running database migrations...")

	err := DB.AutoMigrate(
		&model.Admin{},
		&model.Merchant{},
		&model.Channel{},
		&model.Order{},
		&model.Settlement{},
		&model.BalanceRecord{},
		&model.Config{},
	)

	if err != nil {
		return err
	}

	log.Println("Database migrations completed")
	return nil
}
```

**Step 2: 更新主程序入口**

```go
// cmd/server/main.go
package main

import (
	"fmt"
	"log"

	"github.com/example/epay-go/internal/cache"
	"github.com/example/epay-go/internal/config"
	"github.com/example/epay-go/internal/database"
	"github.com/example/epay-go/internal/middleware"
	"github.com/gin-gonic/gin"
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

	// TODO: 注册路由

	// 启动服务
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
```

**Step 3: 验证程序可编译**

Run: `go build -o epay-server ./cmd/server`
Expected: 生成 `epay-server.exe` 文件

**Step 4: 提交**

```bash
git add .
git commit -m "feat: add database migration and complete main entry"
```

---

## Task 11: 创建 Docker 配置

**Files:**
- Create: `epay-go/Dockerfile`
- Create: `epay-go/docker-compose.yml`
- Create: `epay-go/.dockerignore`

**Step 1: 创建 Dockerfile**

```dockerfile
# Dockerfile

# 阶段1: 构建 Go 后端
FROM golang:1.22-alpine AS backend
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o epay-server ./cmd/server

# 阶段2: 最终镜像
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend /app/epay-server .
COPY --from=backend /app/config.example.yaml ./config.yaml
EXPOSE 8080
CMD ["./epay-server"]
```

**Step 2: 创建 docker-compose.yml**

```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      - GIN_MODE=release
    volumes:
      - ./config.yaml:/app/config.yaml:ro
    restart: unless-stopped

  postgres:
    image: postgres:16-alpine
    volumes:
      - pgdata:/var/lib/postgresql/data
    environment:
      POSTGRES_DB: epay
      POSTGRES_USER: epay
      POSTGRES_PASSWORD: ${DB_PASSWORD:-epay123}
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U epay -d epay"]
      interval: 5s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    volumes:
      - redisdata:/data
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5
    restart: unless-stopped

volumes:
  pgdata:
  redisdata:
```

**Step 3: 创建 .dockerignore**

```
# .dockerignore
.git
.gitignore
*.md
.idea
.vscode
*.log
logs/
.env
.env.*
web/node_modules
web/dist
*.exe
epay-server
```

**Step 4: 提交**

```bash
git add Dockerfile docker-compose.yml .dockerignore
git commit -m "feat: add docker and docker-compose configuration"
```

---

## Task 12: 初始化 Vue3 前端项目

**Files:**
- Create: `epay-go/web/` (Vue3 项目)

**Step 1: 创建 Vue3 项目**

```bash
cd d:/project/payment/epay-go
npm create vite@latest web -- --template vue-ts
```

**Step 2: 安装依赖**

```bash
cd web
npm install
npm install @arco-design/web-vue
npm install vue-router@4 pinia axios
npm install -D @types/node
```

**Step 3: 配置 Arco Design**

修改 `web/src/main.ts`:

```typescript
// web/src/main.ts
import { createApp } from 'vue'
import ArcoVue from '@arco-design/web-vue'
import '@arco-design/web-vue/dist/arco.css'
import App from './App.vue'

const app = createApp(App)
app.use(ArcoVue)
app.mount('#app')
```

**Step 4: 验证前端可运行**

Run: `npm run dev`
Expected: Vite 开发服务器启动，显示 Arco Design 样式

**Step 5: 提交**

```bash
cd d:/project/payment/epay-go
git add web/
git commit -m "feat: initialize vue3 frontend with arco design"
```

---

## 阶段一完成检查清单

- [ ] Go 项目初始化完成
- [ ] 核心依赖安装完成 (Gin, GORM, Redis, Viper, JWT, gopay)
- [ ] 配置管理模块完成
- [ ] 数据库连接模块完成
- [ ] Redis 连接模块完成
- [ ] 统一响应格式完成
- [ ] JWT 认证中间件完成
- [ ] CORS/Logger/Recovery 中间件完成
- [ ] 所有数据模型定义完成
- [ ] 数据库自动迁移完成
- [ ] 主程序入口完成
- [ ] Docker 配置完成
- [ ] Vue3 前端项目初始化完成

---

**下一阶段：** 阶段二将实现后端核心业务功能，包括商户模块、支付通道管理、订单模块、支付适配器层等。
