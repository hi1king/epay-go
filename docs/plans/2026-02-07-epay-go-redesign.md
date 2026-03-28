# EPay Go 重构设计文档

> 日期: 2026-02-07
> 状态: 待实施

## 一、项目概述

将原 PHP 版 EPay 聚合支付系统重构为 Go + Vue3 前后端分离架构，保留核心支付功能，提升安全性、性能和可维护性。

### 技术栈

| 模块 | 技术选型 |
|------|---------|
| 后端框架 | Go + Gin |
| 支付SDK | gopay (github.com/go-pay/gopay) |
| 数据库 | PostgreSQL 16 + GORM |
| 缓存/队列 | Redis 7 |
| 前端 | Vue3 + Arco Design + Vite |
| 状态管理 | Pinia |
| 认证 | JWT (admin/merchant 双token体系) |
| 部署 | Docker + docker-compose |
| 配置管理 | Viper (支持yaml/env) |

### 功能范围

**保留核心功能：**
- 商户管理（注册、登录、信息管理）
- 订单管理（创建、查询、退款）
- 支付通道管理（多渠道配置）
- 结算管理（申请、审核、打款）
- 支付 API 接口（对外提供）

**支持的支付渠道：**
- 支付宝（gopay 原生支持）
- 微信支付（gopay 原生支持）
- PayPal（gopay 原生支持）
- QQ 支付（gopay 原生支持）
- 通联支付（gopay 原生支持）
- 拉卡拉（gopay 原生支持）

---

## 二、整体架构

```
┌─────────────────────────────────────────────────────┐
│                    Nginx / Traefik                    │
│              (反向代理 + SSL 终端)                     │
├──────────────────┬──────────────────────────────────┤
│   Vue3 前端       │         Go 后端 API               │
│  (Arco Design)   │        (Gin Framework)            │
│                  │                                    │
│  ┌────────────┐  │  ┌──────────────────────────────┐ │
│  │ 管理后台    │  │  │  API Gateway Layer           │ │
│  │ (admin)    │  │  │  - JWT 认证中间件             │ │
│  ├────────────┤  │  │  - 限流中间件                 │ │
│  │ 商户中心    │  │  │  - 日志中间件                 │ │
│  │ (merchant) │  │  │  - CORS 中间件                │ │
│  ├────────────┤  │  ├──────────────────────────────┤ │
│  │ 收银台页面  │  │  │  Business Layer              │ │
│  │ (cashier)  │  │  │  - 商户服务                   │ │
│  └────────────┘  │  │  - 订单服务                   │ │
│                  │  │  - 支付服务                   │ │
│                  │  │  - 通道服务                   │ │
│                  │  │  - 结算服务                   │ │
│                  │  ├──────────────────────────────┤ │
│                  │  │  Payment Adapter Layer        │ │
│                  │  │  (gopay SDK 适配层)            │ │
│                  │  │  - Alipay Adapter             │ │
│                  │  │  - WechatPay Adapter          │ │
│                  │  │  - PayPal Adapter             │ │
│                  │  │  - QQ/通联/拉卡拉 Adapter      │ │
│                  │  ├──────────────────────────────┤ │
│                  │  │  Data Layer                   │ │
│                  │  │  - PostgreSQL (GORM)          │ │
│                  │  │  - Redis (缓存+队列)           │ │
│                  │  └──────────────────────────────┘ │
└──────────────────┴──────────────────────────────────┘
```

**核心设计原则：**
- **分层架构**：API层 → 业务层 → 支付适配层 → 数据层，职责清晰
- **支付适配器模式**：所有支付渠道实现统一的 `PaymentAdapter` 接口，新增渠道只需实现接口
- **前后端完全分离**：后端只提供 RESTful API，前端通过 API 交互
- **Redis** 用于缓存配置、限流计数、异步通知队列

---

## 三、项目目录结构

```
epay-go/
├── cmd/
│   └── server/
│       └── main.go                 # 程序入口
├── internal/
│   ├── config/
│   │   └── config.go               # 配置加载（Viper）
│   ├── middleware/
│   │   ├── auth.go                 # JWT 认证
│   │   ├── cors.go                 # 跨域
│   │   ├── ratelimit.go            # 限流
│   │   └── logger.go               # 请求日志
│   ├── model/
│   │   ├── merchant.go             # 商户模型
│   │   ├── order.go                # 订单模型
│   │   ├── channel.go              # 支付通道模型
│   │   ├── settlement.go           # 结算模型
│   │   └── record.go               # 资金记录模型
│   ├── handler/
│   │   ├── admin/                  # 管理后台 API
│   │   │   ├── merchant.go         # 商户管理
│   │   │   ├── order.go            # 订单管理
│   │   │   ├── channel.go          # 通道管理
│   │   │   ├── settlement.go       # 结算管理
│   │   │   └── dashboard.go        # 仪表盘统计
│   │   ├── merchant/               # 商户端 API
│   │   │   ├── auth.go             # 登录注册
│   │   │   ├── order.go            # 订单查询
│   │   │   ├── settlement.go       # 结算申请
│   │   │   └── profile.go          # 商户信息
│   │   └── payment/                # 支付 API（对外）
│   │       ├── create.go           # 创建订单
│   │       ├── query.go            # 查询订单
│   │       ├── notify.go           # 异步回调
│   │       └── refund.go           # 退款
│   ├── service/
│   │   ├── merchant.go             # 商户业务逻辑
│   │   ├── order.go                # 订单业务逻辑
│   │   ├── payment.go              # 支付核心逻辑
│   │   ├── channel.go              # 通道选择逻辑
│   │   ├── settlement.go           # 结算逻辑
│   │   └── notify.go               # 回调通知逻辑
│   ├── repository/
│   │   ├── merchant.go             # 商户数据访问
│   │   ├── order.go                # 订单数据访问
│   │   ├── channel.go              # 通道数据访问
│   │   └── settlement.go           # 结算数据访问
│   └── payment/                    # 支付适配器层
│       ├── adapter.go              # 统一接口定义
│       ├── factory.go              # 适配器工厂
│       ├── alipay.go               # 支付宝适配器
│       ├── wechat.go               # 微信支付适配器
│       ├── paypal.go               # PayPal 适配器
│       ├── qqpay.go                # QQ 支付适配器
│       ├── allinpay.go             # 通联支付适配器
│       └── lakala.go               # 拉卡拉适配器
├── pkg/
│   ├── response/
│   │   └── response.go             # 统一响应格式
│   ├── sign/
│   │   └── sign.go                 # 签名工具
│   └── utils/
│       └── utils.go                # 通用工具函数
├── migrations/
│   └── 001_init.sql                # 数据库迁移
├── web/                            # Vue3 前端项目
│   ├── src/
│   │   ├── views/
│   │   │   ├── admin/              # 管理后台页面
│   │   │   ├── merchant/           # 商户中心页面
│   │   │   └── cashier/            # 收银台页面
│   │   ├── api/                    # API 请求封装
│   │   ├── store/                  # Pinia 状态管理
│   │   ├── router/                 # 路由配置
│   │   └── components/             # 公共组件
│   ├── package.json
│   └── vite.config.ts
├── docker-compose.yml
├── Dockerfile
├── go.mod
└── go.sum
```

**关键设计点：**
- `internal/` 包含所有不对外暴露的业务代码，Go 编译器强制保护
- `handler → service → repository` 三层分离
- `internal/payment/` 是支付适配器层，所有支付渠道实现统一的 `PaymentAdapter` 接口
- `web/` 是独立的 Vue3 项目，构建后可嵌入 Docker 镜像

---

## 四、数据库设计

```sql
-- 商户表
CREATE TABLE merchants (
    id              BIGSERIAL PRIMARY KEY,
    username        VARCHAR(64) UNIQUE NOT NULL,
    password        VARCHAR(128) NOT NULL,
    email           VARCHAR(128),
    phone           VARCHAR(20),
    api_key         VARCHAR(64) UNIQUE NOT NULL,
    public_key      TEXT,
    private_key     TEXT,
    balance         DECIMAL(12,2) DEFAULT 0,
    frozen_balance  DECIMAL(12,2) DEFAULT 0,
    status          SMALLINT DEFAULT 1,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- 管理员表
CREATE TABLE admins (
    id              BIGSERIAL PRIMARY KEY,
    username        VARCHAR(64) UNIQUE NOT NULL,
    password        VARCHAR(128) NOT NULL,
    role            VARCHAR(20) DEFAULT 'admin',
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- 支付通道表
CREATE TABLE channels (
    id              BIGSERIAL PRIMARY KEY,
    name            VARCHAR(64) NOT NULL,
    plugin          VARCHAR(32) NOT NULL,
    pay_types       VARCHAR(255),
    config          JSONB,
    rate            DECIMAL(5,4) DEFAULT 0,
    daily_limit     DECIMAL(12,2) DEFAULT 0,
    status          SMALLINT DEFAULT 1,
    sort            INT DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- 订单表
CREATE TABLE orders (
    id              BIGSERIAL PRIMARY KEY,
    trade_no        VARCHAR(32) UNIQUE NOT NULL,
    out_trade_no    VARCHAR(64) NOT NULL,
    merchant_id     BIGINT REFERENCES merchants(id),
    channel_id      BIGINT REFERENCES channels(id),
    pay_type        VARCHAR(20),
    amount          DECIMAL(12,2) NOT NULL,
    real_amount     DECIMAL(12,2),
    fee             DECIMAL(12,2) DEFAULT 0,
    name            VARCHAR(255),
    notify_url      VARCHAR(512),
    return_url      VARCHAR(512),
    api_trade_no    VARCHAR(64),
    buyer           VARCHAR(64),
    client_ip       VARCHAR(45),
    status          SMALLINT DEFAULT 0,
    notify_status   SMALLINT DEFAULT 0,
    notify_count    INT DEFAULT 0,
    next_notify_at  TIMESTAMPTZ,
    paid_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- 结算记录表
CREATE TABLE settlements (
    id              BIGSERIAL PRIMARY KEY,
    settle_no       VARCHAR(32) UNIQUE NOT NULL,
    merchant_id     BIGINT REFERENCES merchants(id),
    amount          DECIMAL(12,2) NOT NULL,
    fee             DECIMAL(12,2) DEFAULT 0,
    actual_amount   DECIMAL(12,2),
    account_type    VARCHAR(20),
    account_no      VARCHAR(64),
    account_name    VARCHAR(64),
    status          SMALLINT DEFAULT 0,
    remark          VARCHAR(255),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- 资金变动记录表
CREATE TABLE balance_records (
    id              BIGSERIAL PRIMARY KEY,
    merchant_id     BIGINT REFERENCES merchants(id),
    action          SMALLINT NOT NULL,
    amount          DECIMAL(12,2) NOT NULL,
    before_balance  DECIMAL(12,2),
    after_balance   DECIMAL(12,2),
    type            VARCHAR(32),
    trade_no        VARCHAR(64),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- 系统配置表
CREATE TABLE configs (
    key             VARCHAR(64) PRIMARY KEY,
    value           TEXT,
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_orders_merchant_id ON orders(merchant_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at);
CREATE INDEX idx_orders_notify_status ON orders(notify_status) WHERE notify_status < 2;
CREATE INDEX idx_settlements_merchant_id ON settlements(merchant_id);
CREATE INDEX idx_balance_records_merchant_id ON balance_records(merchant_id);
```

**相比原 epay 的改进：**
- 使用 `JSONB` 存储通道配置，灵活且可查询
- 使用 `bcrypt` 替代 `MD5+硬编码盐` 的密码方案
- 使用 `DECIMAL` 精确存储金额
- 通知重试机制内置到订单表
- 使用 PostgreSQL 的 `TIMESTAMPTZ` 处理时区问题

---

## 五、支付适配器接口设计

```go
// internal/payment/adapter.go

// PaymentAdapter 统一支付适配器接口
type PaymentAdapter interface {
    // 创建支付订单，返回支付参数
    CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error)

    // 查询订单状态
    QueryOrder(ctx context.Context, tradeNo string) (*QueryOrderResponse, error)

    // 退款
    Refund(ctx context.Context, req *RefundRequest) (*RefundResponse, error)

    // 解析异步回调通知
    ParseNotify(ctx context.Context, req *http.Request) (*NotifyResult, error)

    // 返回回调成功响应
    NotifySuccess() string

    // 转账（用于结算打款）
    Transfer(ctx context.Context, req *TransferRequest) (*TransferResponse, error)
}

// CreateOrderRequest 统一下单请求
type CreateOrderRequest struct {
    TradeNo     string          // 系统订单号
    Amount      decimal.Decimal // 金额
    Subject     string          // 商品名称
    ClientIP    string          // 客户端IP
    NotifyURL   string          // 异步通知地址
    ReturnURL   string          // 同步跳转地址
    PayMethod   string          // 支付方式: scan/h5/jsapi/app/web
    Extra       map[string]string // 扩展参数
}

// CreateOrderResponse 统一下单响应
type CreateOrderResponse struct {
    PayType     string // redirect / qrcode / jsapi / form
    PayURL      string // 支付链接或二维码内容
    PayParams   string // JSAPI 支付参数（JSON）
    RawResponse string // 原始响应
}

// NotifyResult 统一回调结果
type NotifyResult struct {
    TradeNo      string          // 系统订单号
    ApiTradeNo   string          // 上游订单号
    Amount       decimal.Decimal // 支付金额
    Buyer        string          // 买家标识
    Status       string          // success / fail
}
```

**适配器工厂：**

```go
// internal/payment/factory.go

var adapters = map[string]func(config json.RawMessage) (PaymentAdapter, error){
    "alipay":   NewAlipayAdapter,
    "wechat":   NewWechatAdapter,
    "paypal":   NewPayPalAdapter,
    "qqpay":    NewQQPayAdapter,
    "allinpay": NewAllinpayAdapter,
    "lakala":   NewLakalaAdapter,
}

func NewAdapter(plugin string, config json.RawMessage) (PaymentAdapter, error) {
    fn, ok := adapters[plugin]
    if !ok {
        return nil, fmt.Errorf("unsupported payment plugin: %s", plugin)
    }
    return fn(config)
}
```

**新增支付渠道只需3步：**
1. 在 `internal/payment/` 下新建文件，实现 `PaymentAdapter` 接口
2. 在 `factory.go` 的 `adapters` map 中注册
3. 在管理后台添加通道配置

---

## 六、API 接口设计

### 对外支付 API（供商户网站调用）

```
POST   /api/pay/create          # 创建支付订单
GET    /api/pay/query           # 查询订单状态
POST   /api/pay/refund          # 申请退款
POST   /api/pay/notify/:channel # 支付渠道异步回调
GET    /api/pay/return/:channel # 支付渠道同步跳转
GET    /api/pay/cashier/:trade_no # 收银台页面数据
```

### 管理后台 API

```
POST   /admin/auth/login        # 管理员登录
POST   /admin/auth/logout       # 退出登录

GET    /admin/dashboard         # 仪表盘统计

GET    /admin/merchants         # 商户列表
GET    /admin/merchants/:id     # 商户详情
PUT    /admin/merchants/:id     # 编辑商户
PATCH  /admin/merchants/:id/status # 启用/禁用

GET    /admin/orders            # 订单列表
GET    /admin/orders/:trade_no  # 订单详情
POST   /admin/orders/:trade_no/refund   # 退款
POST   /admin/orders/:trade_no/renotify # 重发通知

GET    /admin/channels          # 通道列表
POST   /admin/channels          # 新增通道
PUT    /admin/channels/:id      # 编辑通道
DELETE /admin/channels/:id      # 删除通道

GET    /admin/settlements       # 结算列表
PATCH  /admin/settlements/:id/approve # 审核通过
PATCH  /admin/settlements/:id/reject  # 驳回

GET    /admin/configs           # 获取系统配置
PUT    /admin/configs           # 更新系统配置
```

### 商户端 API

```
POST   /merchant/auth/login     # 商户登录
POST   /merchant/auth/register  # 商户注册
POST   /merchant/auth/logout    # 退出登录

GET    /merchant/profile        # 商户信息
PUT    /merchant/profile        # 更新信息
PUT    /merchant/profile/password # 修改密码
POST   /merchant/profile/reset-key # 重置API密钥

GET    /merchant/orders         # 我的订单
GET    /merchant/orders/:trade_no # 订单详情

GET    /merchant/settlements    # 结算记录
POST   /merchant/settlements    # 申请结算

GET    /merchant/records        # 资金变动记录
GET    /merchant/dashboard      # 商户仪表盘
```

### 统一响应格式

```json
// 成功
{ "code": 0, "msg": "success", "data": { ... } }

// 失败
{ "code": 1001, "msg": "参数错误: amount 不能为空" }

// 分页
{ "code": 0, "data": { "list": [...], "total": 100, "page": 1, "page_size": 20 } }
```

---

## 七、Docker 部署

### docker-compose.yml

```yaml
services:
  app:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - postgres
      - redis
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis

  postgres:
    image: postgres:16-alpine
    volumes:
      - pgdata:/var/lib/postgresql/data
    environment:
      - POSTGRES_DB=epay
      - POSTGRES_USER=epay
      - POSTGRES_PASSWORD=${DB_PASSWORD}

  redis:
    image: redis:7-alpine
    volumes:
      - redisdata:/data

volumes:
  pgdata:
  redisdata:
```

### Dockerfile（多阶段构建）

```dockerfile
# 阶段1: 构建 Vue 前端
FROM node:20-alpine AS frontend
WORKDIR /app/web
COPY web/ .
RUN npm ci && npm run build

# 阶段2: 构建 Go 后端
FROM golang:1.24-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o epay-server ./cmd/server

# 阶段3: 最终镜像
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=backend /app/epay-server /usr/local/bin/
COPY --from=frontend /app/web/dist /app/static
EXPOSE 8080
CMD ["epay-server"]
```

最终镜像约 **20-30MB**。

### 异步通知重试机制

Go 进程内置 Worker 协程，按策略重试：

```
第1次: 立即发送
第2次: 1分钟后
第3次: 3分钟后
第4次: 20分钟后
第5次: 1小时后
第6次: 2小时后
全部失败 → 标记通知失败，等待手动重发
```

---

## 八、核心改进（对比原 epay）

| 项目 | 原 epay (PHP) | 新系统 (Go) |
|------|--------------|-------------|
| 安全性 | MD5+硬编码盐、无JWT | bcrypt、JWT、无硬编码密钥 |
| 扩展性 | 每个渠道独立插件目录 | 统一适配器接口，3步新增渠道 |
| 性能 | PHP-FPM 单请求模型 | Go 原生高并发，数万 QPS |
| 部署 | PHP + Apache/Nginx + MySQL | Docker 单镜像 20-30MB |
| 异步任务 | 外部 cron.php | 内置 Worker 协程 |
| 代码质量 | 大量 SQL 拼接、无类型 | 类型安全、GORM 防注入 |
| API 兼容 | - | 签名协议兼容原 epay，商户可无缝迁移 |

---

## 九、实施计划

### 阶段一：基础框架（预计 2-3 天）
- [ ] 初始化 Go 项目结构
- [ ] 配置 GORM + PostgreSQL
- [ ] 配置 Redis 连接
- [ ] 实现 JWT 认证中间件
- [ ] 实现统一响应格式

### 阶段二：后端核心功能（预计 5-7 天）
- [ ] 商户模块（注册、登录、信息管理）
- [ ] 支付通道管理
- [ ] 订单模块（创建、查询）
- [ ] 支付适配器层（支付宝、微信）
- [ ] 异步回调处理
- [ ] 结算模块

### 阶段三：管理后台 API（预计 2-3 天）
- [ ] 管理员认证
- [ ] 仪表盘统计
- [ ] 商户管理 CRUD
- [ ] 订单管理
- [ ] 结算审核

### 阶段四：Vue3 前端（预计 5-7 天）
- [ ] 初始化 Vue3 + Arco Design 项目
- [ ] 管理后台页面
- [ ] 商户中心页面
- [ ] 收银台页面

### 阶段五：Docker 部署（预计 1-2 天）
- [ ] 编写 Dockerfile
- [ ] 编写 docker-compose.yml
- [ ] 测试容器化部署

**总计预估：15-22 天**
