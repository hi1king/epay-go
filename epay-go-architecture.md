# 彩虹易支付 Go 重构版 — 完整架构设计方案

> 版本：v1.0.0 · 技术栈：Go 1.22 / Gin / GORM / PostgreSQL · MySQL / Redis / Vue 3 / Docker Compose

---

## 目录

1. [项目概述](#1-项目概述)
2. [技术选型](#2-技术选型)
3. [整体架构](#3-整体架构)
4. [目录结构](#4-目录结构)
5. [数据库设计](#5-数据库设计)
6. [核心功能模块](#6-核心功能模块)
   - 6.1 [商户管理模块](#61-商户管理模块)
   - 6.2 [支付通道模块（插件系统）](#62-支付通道模块插件系统)
   - 6.3 [订单管理模块](#63-订单管理模块)
   - 6.4 [收银台模块](#64-收银台模块)
   - 6.5 [回调通知模块](#65-回调通知模块)
   - 6.6 [提现结算模块](#66-提现结算模块)
   - 6.7 [风控与安全模块](#67-风控与安全模块)
   - 6.8 [统计报表模块](#68-统计报表模块)
   - 6.9 [系统配置模块](#69-系统配置模块)
   - 6.10 [异步任务模块](#610-异步任务模块)
7. [API 接口设计](#7-api-接口设计)
8. [中间件设计](#8-中间件设计)
9. [前端模块设计](#9-前端模块设计)
10. [部署架构](#10-部署架构)
11. [安全设计](#11-安全设计)
12. [性能设计](#12-性能设计)
13. [监控与日志](#13-监控与日志)
14. [开发规范](#14-开发规范)

---

## 1. 项目概述

### 1.1 背景

彩虹易支付（Epay）原版基于 PHP + MySQL 构建，依赖宝塔面板部署，存在以下痛点：

- 高并发下回调通知处理能力不足，易出现重复回调
- PHP 进程模型导致并发处理依赖 Nginx + PHP-FPM 配置调优
- 宝塔面板强依赖，容器化部署困难
- 定时任务基于系统 crontab，无法做到毫秒级精度与监控
- 代码结构耦合严重，新增支付通道需改动多处核心代码

### 1.2 目标

本方案将彩虹易支付完整功能用 **Go** 重写，目标如下：

- 单二进制部署，消除 PHP 运行时依赖
- 原生并发处理，支持每秒万级回调通知
- 支付通道插件化，新增渠道无需改动核心代码
- Docker Compose 一键部署，支持水平扩容
- 完整保留原版功能：多商户、多通道、提现结算、风控、统计报表

### 1.3 功能范围

| 功能域 | 子功能 |
|--------|--------|
| 商户管理 | 注册 / 登录、API 密钥、余额、费率配置 |
| 支付通道 | 支付宝、微信支付、QQ 钱包、银联、自定义通道 |
| 订单系统 | 创建订单、状态机流转、查单、关单 |
| 收银台 | PC / H5 收银页、二维码展示、轮询状态 |
| 回调通知 | 同步跳转、异步通知、重试机制 |
| 提现结算 | 余额管理、提现申请、代付接口 |
| 风控安全 | IP 黑名单、频率限制、金额规则、设备指纹 |
| 统计报表 | 交易统计、利润分析、通道报表、商户报表 |
| 系统配置 | 站点设置、支付通道配置、邮件通知 |
| 平台管理 | 管理员账号、操作日志、系统监控 |

---

## 2. 技术选型

### 2.1 后端

| 组件 | 选型 | 版本 | 理由 |
|------|------|------|------|
| 语言 | Go | 1.22+ | 原生并发、静态编译、部署简单 |
| Web 框架 | Gin | v1.9+ | 高性能路由、中间件生态成熟 |
| ORM | GORM | v2 | 支持 PostgreSQL/MySQL、钩子机制完善 |
| 数据库 | PostgreSQL 16 / MySQL 8 | — | 二选一，默认 PostgreSQL |
| 缓存 | Redis | 7.x | 分布式锁、限流、Session、队列 |
| 配置管理 | Viper | v1.18+ | 支持 YAML/ENV/热重载 |
| 日志 | Zap | v1.27+ | 结构化日志、高性能 |
| 定时任务 | robfig/cron | v3 | 支持秒级 cron 表达式 |
| 参数校验 | validator | v10 | 声明式校验、国际化错误信息 |
| JWT | golang-jwt | v5 | 标准实现、支持 RS256 |
| 加密 | 标准库 crypto | — | RSA / AES / MD5 / SHA256 |

### 2.2 前端

| 组件 | 选型 | 理由 |
|------|------|------|
| 框架 | Vue 3 + Composition API | 响应式、组件复用性强 |
| 构建工具 | Vite 5 | 极速 HMR、ESM 原生支持 |
| UI 组件库 | Arco Design Vue | 字节跳动出品、企业级组件丰富 |
| 状态管理 | Pinia | 轻量、TS 友好 |
| 路由 | Vue Router 4 | 官方标准 |
| HTTP 客户端 | Axios | 拦截器、统一错误处理 |
| 图表 | ECharts 5 | 统计报表可视化 |
| 二维码 | qrcode.js | 前端生成支付二维码 |

### 2.3 基础设施

| 组件 | 选型 | 用途 |
|------|------|------|
| 容器化 | Docker + Docker Compose | 本地开发 / 单机生产部署 |
| 反向代理 | Nginx / Caddy | TLS 终止、静态资源、限流 |
| 监控 | Prometheus + Grafana | 指标采集与可视化 |
| 链路追踪 | OpenTelemetry（可选） | 分布式追踪 |
| CI/CD | GitHub Actions | 自动构建、镜像推送 |

---

## 3. 整体架构

### 3.1 分层架构

```
┌─────────────────────────────────────────────────────────┐
│                      客户端层                            │
│   商户前台(Vue3)   管理后台(Vue3)   收银台 H5(Vue3)      │
└────────────────────┬────────────────────────────────────┘
                     │ HTTPS
┌────────────────────▼────────────────────────────────────┐
│                   接入层 (Nginx/Caddy)                   │
│     TLS 终止 · 限流 · IP 过滤 · 静态资源分发             │
└────────────────────┬────────────────────────────────────┘
                     │ HTTP/1.1 反代
┌────────────────────▼────────────────────────────────────┐
│                  API 网关层 (Gin)                         │
│   路由注册 · 中间件链 · 参数绑定 · 统一响应格式           │
│                                                          │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐   │
│  │ 商户 API │ │  支付核心 │ │ 提现结算 │ │ Admin API│   │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘   │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│                   业务服务层 (Service)                    │
│  订单服务  通道路由  商户服务  风控服务  通知服务  结算    │
└────────────────────┬────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────┐
│                  数据访问层 (Repository)                  │
│              GORM (PostgreSQL/MySQL)                     │
│              go-redis (Redis 7)                          │
└──────────────┬──────────────────────┬───────────────────┘
               │                      │
┌──────────────▼──────┐  ┌────────────▼──────────────────┐
│  PostgreSQL / MySQL │  │  Redis (缓存/锁/队列/限流)      │
└─────────────────────┘  └───────────────────────────────┘
```

### 3.2 支付流程主链路

```
商户系统                  Epay Go                    第三方支付
    │                        │                           │
    │── POST /api/submit ────▶│                           │
    │                        │ 1. 验签 (MD5/RSA)          │
    │                        │ 2. 商户鉴权                │
    │                        │ 3. 风控检查                │
    │                        │ 4. 创建订单 (DB)           │
    │                        │ 5. 选择通道 (路由策略)     │
    │                        │── 创建支付请求 ────────────▶│
    │                        │◀─ 返回支付 URL/参数 ────── │
    │◀── 跳转到收银台 ────── │                           │
    │                        │                           │
  [用户付款]                 │                           │
    │                        │◀─── 异步通知 (notify) ─── │
    │                        │ 1. 验证第三方签名          │
    │                        │ 2. 幂等检查 (Redis)        │
    │                        │ 3. 更新订单状态 (DB)       │
    │                        │ 4. 计算商户余额            │
    │                        │── 写入回调队列 ───────────▶│(Redis List)
    │                        │                           │
    │                        │ [Worker 异步消费]          │
    │                        │── POST notify_url ────────▶│(商户服务器)
    │◀── GET return_url ─── │                           │
```

---

## 4. 目录结构

```
epay-go/
├── cmd/
│   └── server/
│       └── main.go                      # 程序入口，初始化所有组件
│
├── internal/                            # 业务核心（不对外暴露）
│   ├── config/
│   │   ├── config.go                    # 配置结构体定义
│   │   └── loader.go                    # Viper 加载，支持 .env 覆盖
│   │
│   ├── router/
│   │   ├── router.go                    # 根路由注册，分组挂载
│   │   ├── merchant.go                  # 商户开放 API 路由
│   │   ├── pay.go                       # 支付 / 收银台路由
│   │   ├── admin.go                     # 后台管理路由
│   │   └── webhook.go                   # 第三方回调路由
│   │
│   ├── middleware/
│   │   ├── auth_jwt.go                  # JWT 鉴权（商户登录态）
│   │   ├── auth_admin.go                # 管理员鉴权
│   │   ├── sign_verify.go               # 商户 API 签名校验（MD5/RSA）
│   │   ├── rate_limit.go                # 基于 Redis 的滑动窗口限流
│   │   ├── idempotent.go                # 幂等去重（Redis SETNX）
│   │   ├── ip_blacklist.go              # IP 黑名单过滤
│   │   ├── cors.go                      # 跨域处理
│   │   ├── request_log.go               # 请求日志（结构化）
│   │   └── recovery.go                  # Panic 恢复
│   │
│   ├── api/                             # Handler 层（薄层）
│   │   ├── merchant/
│   │   │   ├── auth.go                  # 登录、注册、找回密码
│   │   │   ├── profile.go               # 个人信息、API 密钥
│   │   │   ├── order.go                 # 下单、查单
│   │   │   ├── withdraw.go              # 提现申请
│   │   │   └── stat.go                  # 商户侧统计
│   │   │
│   │   ├── pay/
│   │   │   ├── submit.go                # 统一下单入口（兼容 Epay 协议）
│   │   │   ├── cashier.go               # 收银台页面渲染
│   │   │   ├── query.go                 # 前端轮询订单状态
│   │   │   └── notify.go                # 第三方异步回调接收
│   │   │
│   │   ├── admin/
│   │   │   ├── auth.go                  # 管理员登录
│   │   │   ├── merchant.go              # 商户管理 CRUD
│   │   │   ├── channel.go               # 支付通道配置
│   │   │   ├── order.go                 # 订单管理、搜索
│   │   │   ├── withdraw.go              # 提现审核
│   │   │   ├── blacklist.go             # 黑名单管理
│   │   │   ├── stat.go                  # 平台统计报表
│   │   │   └── system.go                # 系统配置
│   │   │
│   │   └── common/
│   │       └── response.go              # 统一响应格式工具函数
│   │
│   ├── service/                         # 业务逻辑层
│   │   ├── order_service.go             # 订单创建、状态机、查询
│   │   ├── channel_service.go           # 通道路由、负载均衡、降级
│   │   ├── merchant_service.go          # 商户注册、余额、费率
│   │   ├── settlement_service.go        # 提现、代付、对账
│   │   ├── risk_service.go              # 风控规则引擎
│   │   ├── notify_service.go            # 回调通知生产
│   │   ├── stat_service.go              # 统计数据聚合
│   │   └── auth_service.go              # 认证、Token 颁发
│   │
│   ├── model/                           # GORM 数据模型
│   │   ├── base.go                      # 公共字段 (ID, CreatedAt, UpdatedAt)
│   │   ├── merchant.go
│   │   ├── merchant_balance_log.go
│   │   ├── channel.go
│   │   ├── order.go
│   │   ├── notify_log.go
│   │   ├── withdraw.go
│   │   ├── blacklist.go
│   │   ├── admin_user.go
│   │   ├── operation_log.go
│   │   └── system_config.go
│   │
│   ├── repository/                      # 数据访问层
│   │   ├── db.go                        # GORM 初始化、读写分离配置
│   │   ├── redis.go                     # Redis 客户端初始化
│   │   ├── order_repo.go
│   │   ├── merchant_repo.go
│   │   ├── channel_repo.go
│   │   ├── notify_repo.go
│   │   └── withdraw_repo.go
│   │
│   ├── plugin/                          # 支付通道插件系统
│   │   ├── interface.go                 # PayChannel 接口定义
│   │   ├── registry.go                  # 插件注册表
│   │   ├── alipay/
│   │   │   ├── alipay.go                # 支付宝扫码、H5、App
│   │   │   └── notify.go                # 支付宝回调解析
│   │   ├── wechat/
│   │   │   ├── wechat.go                # 微信 Native、JSAPI、H5
│   │   │   └── notify.go
│   │   ├── qqpay/
│   │   │   └── qqpay.go
│   │   ├── unionpay/
│   │   │   └── unionpay.go
│   │   └── custom/
│   │       └── custom.go                # 自定义通道（HTTP 转发模式）
│   │
│   └── worker/
│       ├── cron.go                      # 定时任务注册与启动
│       ├── order_expire.go              # 超时关单任务
│       ├── reconcile.go                 # 对账任务
│       └── notify_worker.go             # 回调通知消费 Worker
│
├── pkg/                                 # 公共工具包（可被外部引用）
│   ├── response/
│   │   └── response.go                  # 统一 JSON 响应结构
│   ├── crypto/
│   │   ├── md5.go                       # MD5 签名
│   │   ├── rsa.go                       # RSA 加解密、签名验证
│   │   └── aes.go                       # AES 加密（敏感配置）
│   ├── qrcode/
│   │   └── qrcode.go                    # 二维码生成（Base64 / 文件）
│   ├── paginate/
│   │   └── paginate.go                  # 分页查询工具
│   ├── snowflake/
│   │   └── snowflake.go                 # 分布式 ID 生成（订单号）
│   ├── ip/
│   │   └── ip.go                        # IP 解析、归属地查询
│   └── validator/
│       └── validator.go                 # 自定义校验规则注册
│
├── frontend/
│   ├── merchant-portal/                 # 商户工作台（Vue 3）
│   ├── admin-panel/                     # 平台管理后台（Vue 3）
│   └── cashier/                         # 收银台 H5（Vue 3）
│
├── deploy/
│   ├── docker-compose.yml               # 生产环境编排
│   ├── docker-compose.dev.yml           # 开发环境编排
│   ├── Dockerfile                       # 多阶段构建
│   ├── nginx/
│   │   ├── nginx.conf
│   │   └── conf.d/
│   │       └── epay.conf
│   ├── postgres/
│   │   └── init.sql                     # 初始化 DDL
│   └── grafana/
│       └── dashboard.json               # Grafana 仪表盘预设
│
├── migrations/                          # 数据库迁移文件
│   ├── 000001_init_schema.up.sql
│   └── 000001_init_schema.down.sql
│
├── scripts/
│   ├── build.sh                         # 编译脚本
│   └── gen_key.sh                       # RSA 密钥对生成
│
├── configs/
│   ├── config.yaml                      # 默认配置
│   └── config.prod.yaml                 # 生产覆盖配置
│
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

---

## 5. 数据库设计

### 5.1 数据模型概览

```
merchants (商户表)
    │
    ├─── merchant_balance_logs (余额变动日志)
    │
    ├─── orders (订单表)
    │        │
    │        └─── notify_logs (回调通知日志)
    │
    └─── withdrawals (提现申请表)

channels (支付通道表)
    │
    └─── orders.channel_id → channels.id

admin_users (管理员表)
    │
    └─── operation_logs (操作日志)

blacklists (黑名单表)
system_configs (系统配置表)
```

### 5.2 核心表结构

#### merchants — 商户表

```sql
CREATE TABLE merchants (
    id              BIGSERIAL PRIMARY KEY,
    merchant_no     VARCHAR(32) UNIQUE NOT NULL,      -- 商户号，雪花ID生成
    name            VARCHAR(100) NOT NULL,             -- 商户名称
    email           VARCHAR(100) UNIQUE NOT NULL,      -- 登录邮箱
    password_hash   VARCHAR(255) NOT NULL,             -- bcrypt 哈希
    api_key         VARCHAR(64) NOT NULL,              -- MD5签名密钥
    public_key      TEXT,                              -- 商户RSA公钥（可选）
    notify_url      VARCHAR(512),                      -- 默认异步通知地址
    return_url      VARCHAR(512),                      -- 默认同步跳转地址
    balance         DECIMAL(18,2) NOT NULL DEFAULT 0,  -- 可用余额（分）
    frozen_balance  DECIMAL(18,2) NOT NULL DEFAULT 0,  -- 冻结余额
    rate            DECIMAL(5,4) NOT NULL DEFAULT 0.006, -- 费率（千分之六）
    status          SMALLINT NOT NULL DEFAULT 1,       -- 1正常 0禁用
    level           SMALLINT NOT NULL DEFAULT 1,       -- 商户等级
    remark          TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

#### channels — 支付通道表

```sql
CREATE TABLE channels (
    id              BIGSERIAL PRIMARY KEY,
    name            VARCHAR(100) NOT NULL,             -- 通道名称，如「支付宝扫码」
    code            VARCHAR(50) UNIQUE NOT NULL,       -- 通道代码，如 alipay_qr
    plugin          VARCHAR(50) NOT NULL,              -- 插件标识，如 alipay
    pay_type        VARCHAR(20) NOT NULL,              -- 支付类型：alipay/wxpay/qqpay
    icon            VARCHAR(255),                      -- 展示图标 URL
    config          JSONB NOT NULL DEFAULT '{}',       -- 通道配置（appid/key等，AES加密存储）
    rate            DECIMAL(5,4) NOT NULL DEFAULT 0,   -- 通道费率
    daily_limit     DECIMAL(18,2) DEFAULT 0,           -- 日限额，0不限
    single_min      DECIMAL(18,2) DEFAULT 0.01,        -- 单笔最小金额
    single_max      DECIMAL(18,2) DEFAULT 100000,      -- 单笔最大金额
    weight          INT NOT NULL DEFAULT 100,          -- 路由权重（多通道均衡）
    status          SMALLINT NOT NULL DEFAULT 1,       -- 1启用 0停用 2维护
    sort            INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

#### orders — 订单表

```sql
CREATE TABLE orders (
    id              BIGSERIAL PRIMARY KEY,
    order_no        VARCHAR(64) UNIQUE NOT NULL,       -- 平台订单号（雪花ID）
    out_trade_no    VARCHAR(128) NOT NULL,             -- 商户订单号
    merchant_id     BIGINT NOT NULL REFERENCES merchants(id),
    channel_id      BIGINT REFERENCES channels(id),
    pay_type        VARCHAR(20) NOT NULL,              -- 前端传入支付类型
    channel_order_no VARCHAR(128),                    -- 第三方平台订单号
    amount          DECIMAL(18,2) NOT NULL,            -- 订单金额
    real_amount     DECIMAL(18,2),                    -- 实际支付金额（部分通道有差异）
    fee             DECIMAL(18,2) NOT NULL DEFAULT 0, -- 手续费
    profit          DECIMAL(18,2) NOT NULL DEFAULT 0, -- 平台利润
    subject         VARCHAR(255) NOT NULL,             -- 商品名称
    body            TEXT,                             -- 商品描述
    notify_url      VARCHAR(512) NOT NULL,
    return_url      VARCHAR(512),
    attach          TEXT,                             -- 透传参数
    client_ip       VARCHAR(64),                      -- 下单IP
    device          VARCHAR(20),                      -- pc/h5/app
    status          SMALLINT NOT NULL DEFAULT 0,       -- 0待支付 1支付中 2成功 3失败 4关闭 5退款
    is_notified     BOOLEAN NOT NULL DEFAULT FALSE,   -- 是否已成功通知商户
    notify_times    SMALLINT NOT NULL DEFAULT 0,      -- 已通知次数
    paid_at         TIMESTAMPTZ,
    expired_at      TIMESTAMPTZ NOT NULL,             -- 订单过期时间
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_merchant_id ON orders(merchant_id);
CREATE INDEX idx_orders_out_trade_no ON orders(out_trade_no);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at);
```

#### notify_logs — 回调通知日志表

```sql
CREATE TABLE notify_logs (
    id              BIGSERIAL PRIMARY KEY,
    order_id        BIGINT NOT NULL REFERENCES orders(id),
    order_no        VARCHAR(64) NOT NULL,
    notify_url      VARCHAR(512) NOT NULL,
    request_body    TEXT,                             -- 发送内容
    response_body   TEXT,                             -- 商户响应
    status_code     INT,                              -- HTTP 状态码
    success         BOOLEAN NOT NULL DEFAULT FALSE,
    attempt         SMALLINT NOT NULL DEFAULT 1,      -- 第几次尝试
    next_at         TIMESTAMPTZ,                      -- 下次重试时间
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

#### withdrawals — 提现申请表

```sql
CREATE TABLE withdrawals (
    id              BIGSERIAL PRIMARY KEY,
    withdraw_no     VARCHAR(64) UNIQUE NOT NULL,
    merchant_id     BIGINT NOT NULL REFERENCES merchants(id),
    amount          DECIMAL(18,2) NOT NULL,
    fee             DECIMAL(18,2) NOT NULL DEFAULT 0,
    real_amount     DECIMAL(18,2) NOT NULL,           -- 实际到账金额
    account_type    VARCHAR(20) NOT NULL,             -- alipay/bank/wechat
    account         VARCHAR(100) NOT NULL,            -- 收款账号
    account_name    VARCHAR(100),                     -- 真实姓名
    bank_name       VARCHAR(100),
    status          SMALLINT NOT NULL DEFAULT 0,       -- 0待审核 1审核中 2已打款 3拒绝
    admin_remark    TEXT,                             -- 审核备注
    reviewed_by     BIGINT REFERENCES admin_users(id),
    reviewed_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

#### blacklists — 黑名单表

```sql
CREATE TABLE blacklists (
    id          BIGSERIAL PRIMARY KEY,
    type        VARCHAR(20) NOT NULL,     -- ip / merchant_no / device / card
    value       VARCHAR(255) NOT NULL,
    reason      TEXT,
    expired_at  TIMESTAMPTZ,             -- NULL 表示永久
    created_by  BIGINT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(type, value)
);
```

#### system_configs — 系统配置表

```sql
CREATE TABLE system_configs (
    id          BIGSERIAL PRIMARY KEY,
    key         VARCHAR(100) UNIQUE NOT NULL,
    value       TEXT NOT NULL,
    type        VARCHAR(20) DEFAULT 'string',  -- string/json/bool/int
    group       VARCHAR(50),                   -- 配置分组
    label       VARCHAR(100),                  -- 前端展示标签
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

## 6. 核心功能模块

### 6.1 商户管理模块

#### 功能说明

商户是系统的核心使用者，通过商户 API 接入收款功能，通过商户后台查看订单、申请提现、管理 API 密钥。

#### 商户注册与登录

- 邮箱 + 密码注册，bcrypt 加密存储，支持邮箱验证码激活
- 支持管理员手动开户（后台创建商户）
- 登录颁发 JWT AccessToken（有效期 2 小时）+ RefreshToken（有效期 7 天）
- 支持 IP 绑定（可选），异地登录触发邮件告警

```go
// internal/service/auth_service.go
type AuthService struct {
    merchantRepo *repository.MerchantRepo
    redis        *redis.Client
}

func (s *AuthService) Login(email, password string) (*TokenPair, error) {
    merchant, err := s.merchantRepo.FindByEmail(email)
    if err != nil || !bcrypt.CheckPasswordHash(password, merchant.PasswordHash) {
        return nil, ErrInvalidCredentials
    }
    if merchant.Status == StatusDisabled {
        return nil, ErrMerchantDisabled
    }
    return s.generateTokenPair(merchant)
}
```

#### API 密钥管理

- 每个商户拥有一个 `api_key`（MD5 签名用），可重置
- 可选上传 RSA 公钥，开启 RSA 签名模式（更安全）
- 密钥重置后旧密钥立即失效，Redis 中刷新缓存

#### 商户费率配置

- 平台默认费率（通过 system_configs 配置，如 0.6%）
- 可对单个商户设置个性化费率（覆盖默认值）
- 费率在订单成功时计算并记录：`fee = amount × rate`，`profit = fee - 通道成本`

#### 商户余额管理

- 订单成功后自动入账：`balance += (amount - fee)`
- 所有余额变动写 `merchant_balance_logs` 保证可审计
- 余额操作必须在数据库事务中进行，结合 Redis 分布式锁防止并发扣款

```go
// 余额变动类型枚举
const (
    BalanceInOrder    = "order_in"     // 订单收款入账
    BalanceOutWithdraw = "withdraw_out" // 提现扣款
    BalanceAdjustIn   = "adjust_in"    // 平台人工加款
    BalanceAdjustOut  = "adjust_out"   // 平台人工扣款
)
```

---

### 6.2 支付通道模块（插件系统）

#### 设计原则

支付通道采用**插件化接口**设计，所有通道实现同一个 `PayChannel` 接口，通过注册表统一管理，新增通道无需改动核心代码。

#### 插件接口定义

```go
// internal/plugin/interface.go

// CreateOrderReq 创建支付订单请求
type CreateOrderReq struct {
    OrderNo    string          // 平台订单号
    Subject    string          // 商品名称
    Amount     decimal.Decimal // 支付金额
    NotifyURL  string          // 回调通知地址
    ReturnURL  string          // 同步跳转地址
    ClientIP   string          // 用户 IP
    Device     string          // pc/h5/app
    Extra      map[string]any  // 通道特有参数
    Config     ChannelConfig   // 通道配置
}

// CreateOrderResp 创建支付订单响应
type CreateOrderResp struct {
    PayURL       string         // 支付跳转 URL（H5场景）
    QRCodeURL    string         // 支付二维码内容（扫码场景）
    PayParams    map[string]any // JSAPI场景返回给前端的参数
    ChannelOrderNo string       // 第三方订单号
}

// NotifyResult 第三方回调解析结果
type NotifyResult struct {
    OrderNo        string          // 平台订单号
    ChannelOrderNo string          // 第三方订单号
    Amount         decimal.Decimal // 实际支付金额
    Status         string          // success/fail
    RawParams      map[string]string // 原始参数（用于日志）
}

// PayChannel 支付通道插件接口
type PayChannel interface {
    Code() string                                                             // 唯一标识
    Name() string                                                             // 人类可读名称
    SupportedDevices() []string                                               // 支持的设备类型
    CreateOrder(ctx context.Context, req *CreateOrderReq) (*CreateOrderResp, error)
    ParseNotify(ctx context.Context, r *http.Request) (*NotifyResult, error)
    QueryOrder(ctx context.Context, channelOrderNo string, cfg ChannelConfig) (*QueryResult, error)
    ReplySuccess() string                                                     // 回调成功应答内容（如 "success"）
}
```

#### 插件注册表

```go
// internal/plugin/registry.go
var registry = make(map[string]PayChannel)

func Register(ch PayChannel) {
    registry[ch.Code()] = ch
}

func Get(code string) (PayChannel, bool) {
    ch, ok := registry[code]
    return ch, ok
}

// init() 在各插件包的 init 函数中自动注册
// 只需在 main.go 中 import _ "epay-go/internal/plugin/alipay" 即可
```

#### 通道路由策略

通道路由支持以下策略，通过系统配置切换：

- **指定通道**：商户下单时直接指定 `type` 参数（如 `alipay`），系统找到该类型下状态正常的通道
- **权重随机**：同类型多个通道按权重随机选择，用于流量分摊
- **轮询**：同类型通道依次选择，均匀分布
- **金额路由**：根据订单金额选择不同通道（如大额走通道A，小额走通道B）
- **自动降级**：通道连续失败次数超阈值时自动标记为维护状态，触发切换

```go
// internal/service/channel_service.go
func (s *ChannelService) Route(payType string, amount decimal.Decimal) (*model.Channel, error) {
    channels, err := s.channelRepo.FindAvailable(payType)
    if err != nil || len(channels) == 0 {
        return nil, ErrNoAvailableChannel
    }
    return s.weightedRandom(channels), nil
}
```

#### 内置通道实现

**支付宝通道（alipay）**

- 支持场景：扫码（PC）、H5支付、App支付
- 使用官方 SDK：`github.com/smartwalle/alipay/v3`
- 配置项：AppID、应用私钥、支付宝公钥、是否沙箱
- 回调验签：使用支付宝提供的公钥对签名进行 RSA2 验证

**微信支付通道（wechat）**

- 支持场景：Native 扫码、JSAPI（公众号/小程序）、H5支付
- 使用官方 SDK：`github.com/wechatpay-apiv3/wechatpay-go`
- 配置项：AppID、MchID、APIv3密钥、商户证书
- 回调验签：使用微信平台证书进行 AEAD_AES_256_GCM 解密

**QQ 钱包通道（qqpay）**

- 支持场景：扫码支付
- 使用腾讯支付 API 对接，签名方式同微信

**银联云闪付（unionpay）**

- 支持场景：网页支付、手机支付

**自定义通道（custom）**

- 通过 HTTP 转发方式对接任意第三方聚合支付
- 配置项：API地址、商户ID、密钥、签名方式（MD5/RSA/自定义）
- 支持配置请求映射（字段名称转换）和响应映射

---

### 6.3 订单管理模块

#### 订单状态机

```
                    ┌─────────────┐
                    │   PENDING   │  ← 订单创建
                    │  (待支付)   │
                    └──────┬──────┘
              超时关单      │  用户开始支付
                   ┌───────┴────────┐
                   ▼                ▼
            ┌─────────────┐  ┌─────────────┐
            │   CLOSED    │  │   PAYING    │
            │  (已关闭)   │  │  (支付中)   │
            └─────────────┘  └──────┬──────┘
                                    │
                          ┌─────────┴─────────┐
                          ▼                   ▼
                   ┌─────────────┐    ┌─────────────┐
                   │   SUCCESS   │    │   FAILED    │
                   │  (支付成功) │    │  (支付失败) │
                   └──────┬──────┘    └─────────────┘
                          │
                    申请退款（预留）
                          ▼
                   ┌─────────────┐
                   │  REFUNDED   │
                   │  (已退款)   │
                   └─────────────┘
```

#### 创建订单流程

```go
// internal/service/order_service.go
func (s *OrderService) Create(ctx context.Context, req *CreateOrderReq) (*CreateOrderResp, error) {
    // 1. 参数校验（金额范围、通道可用性）
    // 2. 幂等检查：同一商户同一 out_trade_no 已存在则直接返回
    existOrder, _ := s.orderRepo.FindByOutTradeNo(req.MerchantID, req.OutTradeNo)
    if existOrder != nil {
        return s.buildResp(existOrder), nil
    }
    // 3. 风控检查（调用 RiskService）
    if err := s.riskService.Check(ctx, req); err != nil {
        return nil, err
    }
    // 4. 通道路由
    channel, err := s.channelService.Route(req.PayType, req.Amount)
    // 5. 生成平台订单号（雪花ID）
    orderNo := s.snowflake.NextID()
    // 6. 调用通道插件创建支付
    plugin, _ := plugin.Get(channel.Plugin)
    payResp, err := plugin.CreateOrder(ctx, buildPluginReq(req, channel))
    // 7. 写库（事务）
    order := buildOrderModel(req, channel, payResp, orderNo)
    s.orderRepo.Create(order)
    // 8. Redis 设置过期标记（用于超时关单）
    s.redis.SetEx(ctx, "order_expire:"+orderNo, 1, req.ExpireIn)
    return buildCreateResp(order, payResp), nil
}
```

#### 订单查询

- 支持按平台订单号、商户订单号查询
- 提供 `GET /api/query` 接口供收银台轮询状态（带缓存，30秒内命中 Redis）
- 管理后台支持多条件组合查询（商户、时间范围、状态、金额范围、支付类型）

#### 订单对账

- 每日凌晨通过 `reconcile` 任务主动向第三方查询 `SUCCESS` 状态订单
- 比对本地状态与第三方状态，记录差异并告警
- 对于第三方已成功但本地未更新的订单，触发补单逻辑

---

### 6.4 收银台模块

#### 功能说明

收银台是用户实际付款的交互界面，分为 PC 版和 H5 版，支持多种支付方式切换。

#### 收银台页面逻辑

```
GET /pay/{order_no}
    │
    ├── 查询订单信息（状态、金额、商品名、过期时间）
    ├── 获取可用支付通道列表（按设备筛选）
    ├── 渲染收银台页面（Vue 3 SPA）
    └── 前端发起 POST /pay/cashier/create
            │
            ├── 后端调用通道插件生成支付参数
            │   ├── 扫码支付 → 返回二维码内容，前端生成 QR 展示
            │   ├── H5支付  → 返回跳转 URL，前端 window.location 跳转
            │   └── JSAPI   → 返回 wx.config 参数，前端调起微信支付
            │
            └── 前端每 2 秒轮询 GET /pay/query/{order_no}
                    │
                    └── 支付成功 → 跳转 return_url
```

#### 倒计时过期

- 前端读取 `expired_at` 字段展示倒计时
- 倒计时结束前端提示过期，后端定时任务处理实际关单
- 过期订单再次访问收银台返回 `订单已关闭` 提示页

#### 设备适配

- 后端通过 User-Agent 解析 `device` 字段（pc/h5/wechat/alipay）
- 收银台根据 `device` 自动展示对应支付方式
  - 微信内置浏览器：仅展示 JSAPI 支付
  - 支付宝 App：仅展示支付宝支付
  - 普通 H5：展示所有支持 H5 的通道
  - PC：展示所有支持扫码的通道

---

### 6.5 回调通知模块

#### 设计原则

回调通知采用**异步生产消费**模式，接收第三方回调与通知商户解耦，保证高可用和可重试。

#### 第三方回调接收

```go
// internal/api/pay/notify.go
func (h *NotifyHandler) HandleNotify(c *gin.Context) {
    channelCode := c.Param("channel") // 路由: /pay/notify/:channel
    plugin, ok := pluginRegistry.Get(channelCode)
    if !ok {
        c.String(400, "unknown channel")
        return
    }
    // 1. 调用插件解析并验签
    result, err := plugin.ParseNotify(c.Request.Context(), c.Request)
    if err != nil {
        c.String(400, "invalid notify")
        return
    }
    // 2. 幂等检查（Redis SETNX，key=notify:channelOrderNo，TTL=24h）
    if !h.redis.SetNX(ctx, "notify:"+result.ChannelOrderNo, 1, 24*time.Hour).Val() {
        c.String(200, plugin.ReplySuccess()) // 已处理，直接应答防重复推送
        return
    }
    // 3. 更新订单状态（数据库事务）
    if err := h.orderService.MarkSuccess(ctx, result); err != nil {
        h.redis.Del(ctx, "notify:"+result.ChannelOrderNo) // 失败则释放幂等锁
        c.String(500, "internal error")
        return
    }
    // 4. 将商户回调任务写入 Redis 队列
    h.notifyService.Enqueue(ctx, result.OrderNo)
    // 5. 立即应答第三方（必须在 5 秒内）
    c.String(200, plugin.ReplySuccess())
}
```

#### 商户回调通知 Worker

```go
// internal/worker/notify_worker.go

// 重试间隔配置（指数退避）
var retryIntervals = []time.Duration{
    10 * time.Second,
    30 * time.Second,
    2 * time.Minute,
    5 * time.Minute,
    15 * time.Minute,
    30 * time.Minute,
    1 * time.Hour,
    3 * time.Hour,
    6 * time.Hour,
}

func (w *NotifyWorker) Run(ctx context.Context) {
    for {
        // 从 Redis List 阻塞式取任务（BLPOP，超时 5s）
        result, err := w.redis.BLPop(ctx, 5*time.Second, "queue:notify").Result()
        if err != nil { continue }
        orderNo := result[1]
        go w.processNotify(ctx, orderNo)
    }
}

func (w *NotifyWorker) processNotify(ctx context.Context, orderNo string) {
    order, _ := w.orderRepo.FindByOrderNo(orderNo)
    // 构建回调参数（兼容原 Epay 协议）
    params := buildNotifyParams(order)
    params["sign"] = sign(params, order.Merchant.APIKey)
    // HTTP POST 通知商户
    resp, err := http.PostForm(order.NotifyURL, url.Values(params))
    success := err == nil && resp.StatusCode == 200 && readBody(resp) == "success"
    // 记录通知日志
    w.notifyRepo.SaveLog(order.ID, order.NotifyURL, params, resp, success)
    if !success && order.NotifyTimes < len(retryIntervals) {
        // 写入延迟队列（Redis ZADD，score=下次执行时间戳）
        nextAt := time.Now().Add(retryIntervals[order.NotifyTimes])
        w.redis.ZAdd(ctx, "queue:notify_delay", redis.Z{
            Score:  float64(nextAt.Unix()),
            Member: orderNo,
        })
    }
}
```

#### 回调参数格式（兼容原 Epay 协议）

```
pid         商户ID
trade_no    平台订单号
out_trade_no 商户订单号
type        支付类型（alipay/wxpay）
name        商品名称
money       订单金额
trade_status TRADE_SUCCESS
sign        MD5签名
sign_type   MD5
```

---

### 6.6 提现结算模块

#### 提现申请流程

```
商户申请提现
    │
    ├── 校验：余额是否足够 + 未到最低提现金额
    ├── 余额冻结（balance - amount, frozen_balance + amount）
    ├── 创建 withdrawal 记录（status=待审核）
    └── 通知平台管理员（邮件/站内通知）

管理员审核
    ├── 同意 → 触发代付
    │         ├── 调用代付接口（支付宝单笔转账/微信企业付款）
    │         ├── 成功：frozen_balance - amount，记录变动日志
    │         └── 失败：frozen_balance 回退，更新状态为失败
    └── 拒绝 → frozen_balance 回退，填写拒绝原因
```

#### 余额操作的并发安全

```go
// 使用 Redis 分布式锁 + 数据库乐观锁双重保障
func (s *SettlementService) Withdraw(ctx context.Context, merchantID int64, amount decimal.Decimal) error {
    lockKey := fmt.Sprintf("lock:merchant:balance:%d", merchantID)
    // 获取分布式锁（超时 5s）
    lock, err := s.redis.SetNX(ctx, lockKey, 1, 5*time.Second)
    if !lock { return ErrOperationTooFrequent }
    defer s.redis.Del(ctx, lockKey)

    return s.db.Transaction(func(tx *gorm.DB) error {
        var merchant model.Merchant
        // 悲观锁 FOR UPDATE
        tx.Set("gorm:query_option", "FOR UPDATE").First(&merchant, merchantID)
        if merchant.Balance.LessThan(amount) {
            return ErrInsufficientBalance
        }
        tx.Model(&merchant).Updates(map[string]any{
            "balance":        merchant.Balance.Sub(amount),
            "frozen_balance": merchant.FrozenBalance.Add(amount),
        })
        // 记录余额变动日志
        tx.Create(&model.MerchantBalanceLog{...})
        return nil
    })
}
```

---

### 6.7 风控与安全模块

#### 风控规则

风控服务在订单创建前执行，支持以下规则维度：

**IP 维度**

- IP 黑名单：精确匹配或 CIDR 段匹配
- IP 频率限制：滑动窗口计数，默认同 IP 每分钟最多下单 10 次
- IP 归属地限制：可配置禁止特定地区下单

**商户维度**

- 商户日交易限额：可对单个商户设置每日最大交易总金额
- 商户单笔限额：最小/最大金额校验

**通道维度**

- 通道日限额：防止单个通道被超额使用
- 通道单笔限额：超出通道配置的金额范围自动拒绝

**金额维度**

- 可配置异常金额范围（如 `0.01` 以下、`9999999` 以上自动拦截）
- 可配置整数金额检测（非整数金额告警）

#### 风控实现

```go
// internal/service/risk_service.go
type RiskService struct {
    redis    *redis.Client
    blacklistRepo *repository.BlacklistRepo
}

func (s *RiskService) Check(ctx context.Context, req *CreateOrderReq) error {
    checks := []func(context.Context, *CreateOrderReq) error{
        s.checkIPBlacklist,
        s.checkIPRate,
        s.checkMerchantDailyLimit,
        s.checkAmountRange,
        s.checkChannelDailyLimit,
    }
    for _, check := range checks {
        if err := check(ctx, req); err != nil {
            return err
        }
    }
    return nil
}

func (s *RiskService) checkIPRate(ctx context.Context, req *CreateOrderReq) error {
    key := fmt.Sprintf("risk:ip_rate:%s", req.ClientIP)
    count, _ := s.redis.Incr(ctx, key).Result()
    if count == 1 {
        s.redis.Expire(ctx, key, time.Minute)
    }
    if count > s.config.IPRateLimit {
        return ErrIPRateLimitExceeded
    }
    return nil
}
```

#### 黑名单管理

- 支持添加 IP 黑名单、商户号黑名单、设备指纹黑名单
- 黑名单支持设置过期时间，过期后自动解除
- Redis 缓存黑名单（TTL=5分钟），减少数据库查询
- 管理后台支持批量导入黑名单（CSV 格式）

---

### 6.8 统计报表模块

#### 统计维度

**平台总览（Admin）**

- 今日/本周/本月：交易笔数、交易金额、手续费收入、利润
- 按支付类型分布（饼图）
- 近 30 天交易趋势（折线图）
- 通道健康状态：成功率、平均响应时间

**商户报表（Merchant）**

- 商户自己的：收款金额、手续费、可用余额、待提现
- 近 30 天收款趋势
- 订单成功率

**通道报表（Admin）**

- 各通道：交易量、成功率、平均到账时间、当日已用额度

#### 统计实现策略

- 实时数据（今日）：直接聚合查询数据库（带缓存 5 分钟）
- 历史数据：通过定时任务每日凌晨聚合写入统计汇总表（`stat_daily` 表）

```sql
CREATE TABLE stat_daily (
    id          BIGSERIAL PRIMARY KEY,
    date        DATE NOT NULL,
    type        VARCHAR(20) NOT NULL,  -- platform/merchant/channel
    ref_id      BIGINT,               -- merchant_id 或 channel_id，平台级为 NULL
    order_count INT NOT NULL DEFAULT 0,
    success_count INT NOT NULL DEFAULT 0,
    total_amount DECIMAL(18,2) NOT NULL DEFAULT 0,
    fee_amount  DECIMAL(18,2) NOT NULL DEFAULT 0,
    profit_amount DECIMAL(18,2) NOT NULL DEFAULT 0,
    UNIQUE(date, type, ref_id)
);
```

---

### 6.9 系统配置模块

#### 配置分组

| 分组 | 配置项 |
|------|--------|
| site | 站点名称、Logo、备案号、客服邮箱 |
| payment | 默认费率、最低提现金额、提现手续费、订单过期时间 |
| email | SMTP 服务器、端口、发件人、密码 |
| risk | IP 频率限制、单笔限额上下限 |
| notify | 回调最大重试次数、重试间隔配置 |
| register | 是否开放注册、邀请码模式 |

#### 配置热更新

- 配置存储于 `system_configs` 表，管理员修改后同步更新 Redis 缓存
- 应用层使用 `ConfigCache` 对象，订阅 Redis 更新事件实现热更新
- 核心配置（数据库DSN、Redis地址）仅支持环境变量或 YAML 文件，不存数据库

---

### 6.10 异步任务模块

#### 任务列表

| 任务名 | Cron 表达式 | 说明 |
|--------|------------|------|
| 超时关单 | `*/1 * * * * *`（每分钟） | 扫描 Redis 中到期的订单号，批量关单 |
| 每日统计 | `0 5 0 * * *`（每日0:05） | 聚合昨日数据写入 stat_daily |
| 对账任务 | `0 30 1 * * *`（每日1:30） | 拉取第三方已完结订单，比对本地状态 |
| 延迟通知 | `*/5 * * * * *`（每5秒） | 从 Redis ZSet 取到期的延迟通知任务 |
| 余额对账 | `0 0 2 * * *`（每日2:00） | 统计所有商户余额与订单差额，告警异常 |

#### 超时关单实现

```go
// internal/worker/order_expire.go
func (w *OrderExpireWorker) Run(ctx context.Context) {
    // Redis keyspace 通知监听 key 过期事件
    // key 格式：order_expire:{order_no}，值无关紧要
    pubsub := w.redis.Subscribe(ctx, "__keyevent@0__:expired")
    for msg := range pubsub.Channel() {
        if !strings.HasPrefix(msg.Payload, "order_expire:") { continue }
        orderNo := strings.TrimPrefix(msg.Payload, "order_expire:")
        go w.closeExpiredOrder(ctx, orderNo)
    }
}
```

#### 延迟通知消费

```go
// 每 5 秒扫描 ZSet，取 score <= now 的任务
func (w *NotifyWorker) ConsumeDelayed(ctx context.Context) {
    now := float64(time.Now().Unix())
    results, _ := w.redis.ZRangeByScoreWithScores(ctx, "queue:notify_delay", &redis.ZRangeBy{
        Min: "0",
        Max: strconv.FormatFloat(now, 'f', 0, 64),
        Limit: &redis.Limit{Offset: 0, Count: 100},
    }).Result()
    for _, r := range results {
        w.redis.ZRem(ctx, "queue:notify_delay", r.Member)
        w.redis.LPush(ctx, "queue:notify", r.Member)
    }
}
```

---

## 7. API 接口设计

### 7.1 接口规范

- 基础路径：`/api/v1/`
- 请求格式：`application/json` 或 `application/x-www-form-urlencoded`
- 响应格式：统一 JSON，字段如下：

```json
{
  "code": 0,
  "msg": "success",
  "data": {}
}
```

- 错误码规范：

| 错误码 | 含义 |
|--------|------|
| 0 | 成功 |
| 1001 | 参数错误 |
| 1002 | 签名验证失败 |
| 1003 | 商户不存在或已禁用 |
| 1004 | 通道不可用 |
| 1005 | 订单不存在 |
| 1006 | 余额不足 |
| 2001 | 风控拦截 |
| 5000 | 服务器内部错误 |

### 7.2 商户开放 API

#### 统一下单（兼容原 Epay 协议）

```
POST /submit
Content-Type: application/x-www-form-urlencoded

参数：
  pid        商户ID
  type       支付类型（alipay/wxpay/qqpay）
  out_trade_no 商户订单号（唯一）
  notify_url 异步通知地址
  return_url 同步跳转地址（可选）
  name       商品名称
  money      金额（元，精确到小数点后两位）
  sign       MD5签名
  sign_type  MD5
  attach     透传参数（可选）
  device     设备类型 pc/h5（可选，默认pc）

响应（收银台模式）：
  302 跳转至 /pay/{order_no}

响应（API模式，带 ?api=1）：
  {
    "code": 0,
    "data": {
      "order_no": "202401011234567890",
      "pay_url": "https://...",  // H5 跳转 URL
      "qr_url": "https://...",   // 收银台二维码 URL
      "expired_at": "2024-01-01T12:30:00Z"
    }
  }
```

#### 查询订单

```
GET /api/query?pid={pid}&out_trade_no={out_trade_no}&sign={sign}&sign_type=MD5

响应：
  {
    "code": 0,
    "data": {
      "trade_no": "202401011234567890",
      "out_trade_no": "merchant_order_001",
      "trade_status": "TRADE_SUCCESS",
      "money": "99.00",
      "type": "alipay",
      "paid_at": "2024-01-01T12:05:00Z"
    }
  }
```

### 7.3 商户后台 API

```
POST   /api/merchant/auth/login          登录
POST   /api/merchant/auth/refresh        刷新 Token
GET    /api/merchant/profile             获取个人信息
PUT    /api/merchant/profile             更新信息
POST   /api/merchant/profile/reset-key  重置 API 密钥
GET    /api/merchant/orders              订单列表（分页、筛选）
GET    /api/merchant/orders/:order_no   订单详情
GET    /api/merchant/stat/overview       统计总览
GET    /api/merchant/stat/trend          近30天趋势
GET    /api/merchant/balance             余额信息
POST   /api/merchant/withdraw            申请提现
GET    /api/merchant/withdrawals         提现记录
```

### 7.4 平台管理 API

```
POST   /api/admin/auth/login             管理员登录
GET    /api/admin/merchants              商户列表
POST   /api/admin/merchants              创建商户
PUT    /api/admin/merchants/:id          编辑商户
PATCH  /api/admin/merchants/:id/status   启用/禁用
GET    /api/admin/channels               通道列表
POST   /api/admin/channels               创建通道
PUT    /api/admin/channels/:id           编辑通道
GET    /api/admin/orders                 订单查询
GET    /api/admin/withdrawals            提现列表
PATCH  /api/admin/withdrawals/:id/approve 审核通过
PATCH  /api/admin/withdrawals/:id/reject  拒绝
GET    /api/admin/blacklists             黑名单列表
POST   /api/admin/blacklists             添加黑名单
DELETE /api/admin/blacklists/:id         移除黑名单
GET    /api/admin/stat/platform          平台统计
GET    /api/admin/stat/channels          通道报表
GET    /api/admin/system/configs         获取系统配置
PUT    /api/admin/system/configs         保存系统配置
GET    /api/admin/operation-logs         操作日志
```

---

## 8. 中间件设计

### 8.1 JWT 鉴权中间件

```go
func JWTAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractToken(c)
        claims, err := parseJWT(token)
        if err != nil {
            c.AbortWithStatusJSON(401, response.Fail(1401, "unauthorized"))
            return
        }
        c.Set("merchant_id", claims.MerchantID)
        c.Set("merchant_no", claims.MerchantNo)
        c.Next()
    }
}
```

### 8.2 商户 API 签名校验中间件

```go
// 支持 MD5 和 RSA 两种签名模式
func MerchantSignVerify() gin.HandlerFunc {
    return func(c *gin.Context) {
        pid := c.GetParam("pid")
        merchant := merchantRepo.FindByMerchantNo(pid)
        if merchant == nil {
            c.AbortWithStatusJSON(400, response.Fail(1003, "merchant not found"))
            return
        }
        signType := c.GetParam("sign_type")
        switch signType {
        case "MD5":
            if !verifyMD5Sign(c.Request, merchant.APIKey) {
                c.Abort(); return
            }
        case "RSA":
            if !verifyRSASign(c.Request, merchant.PublicKey) {
                c.Abort(); return
            }
        }
        c.Set("merchant", merchant)
        c.Next()
    }
}
```

### 8.3 限流中间件

```go
// 基于 Redis 滑动窗口算法
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        key := fmt.Sprintf("ratelimit:%s:%s", c.ClientIP(), c.FullPath())
        now := time.Now()
        windowStart := now.Add(-window).UnixMilli()
        pipe := redis.Pipeline()
        pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10))
        pipe.ZCard(ctx, key)
        pipe.ZAdd(ctx, key, redis.Z{Score: float64(now.UnixMilli()), Member: now.UnixNano()})
        pipe.Expire(ctx, key, window)
        results, _ := pipe.Exec(ctx)
        count := results[1].(*redis.IntCmd).Val()
        if count >= int64(limit) {
            c.AbortWithStatusJSON(429, response.Fail(4290, "rate limit exceeded"))
            return
        }
        c.Next()
    }
}
```

### 8.4 操作日志中间件

```go
// 记录 Admin API 所有写操作
func OperationLog() gin.HandlerFunc {
    return func(c *gin.Context) {
        if c.Request.Method == "GET" {
            c.Next()
            return
        }
        body, _ := io.ReadAll(c.Request.Body)
        c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
        c.Next()
        adminID, _ := c.Get("admin_id")
        go saveOperationLog(adminID, c.Request, body, c.Writer.Status())
    }
}
```

---

## 9. 前端模块设计

### 9.1 商户工作台（merchant-portal）

```
src/
├── views/
│   ├── Dashboard.vue         # 总览（统计卡片 + 趋势图）
│   ├── Orders.vue            # 订单列表（搜索/筛选/分页）
│   ├── OrderDetail.vue       # 订单详情
│   ├── Withdraw.vue          # 申请提现
│   ├── WithdrawHistory.vue   # 提现记录
│   ├── Profile.vue           # 个人信息
│   └── ApiKey.vue            # API 密钥管理
├── composables/
│   ├── useOrder.js           # 订单相关 hooks
│   └── useStat.js            # 统计数据 hooks
└── components/
    ├── StatCard.vue           # 统计卡片
    ├── TrendChart.vue         # 趋势折线图（ECharts）
    └── OrderTable.vue         # 订单表格
```

### 9.2 平台管理后台（admin-panel）

```
src/
├── views/
│   ├── Dashboard.vue          # 平台总览
│   ├── merchants/
│   │   ├── MerchantList.vue   # 商户管理
│   │   └── MerchantForm.vue   # 新建/编辑商户
│   ├── channels/
│   │   ├── ChannelList.vue    # 通道管理
│   │   └── ChannelForm.vue    # 通道配置（动态表单，不同插件不同字段）
│   ├── orders/
│   │   └── OrderList.vue      # 全平台订单查询
│   ├── withdrawals/
│   │   └── WithdrawList.vue   # 提现审核
│   ├── risk/
│   │   └── Blacklist.vue      # 黑名单管理
│   ├── stat/
│   │   ├── PlatformStat.vue   # 平台统计
│   │   └── ChannelStat.vue    # 通道报表
│   └── system/
│       └── SystemConfig.vue   # 系统配置
```

### 9.3 收银台（cashier）

```
src/
├── views/
│   ├── Cashier.vue            # 主收银台（PC/H5自适应）
│   ├── Success.vue            # 支付成功页
│   ├── Fail.vue               # 支付失败页
│   └── Expired.vue            # 订单过期页
└── components/
    ├── PayTypeSelector.vue    # 支付方式选择
    ├── QRCodeDisplay.vue      # 二维码展示 + 刷新
    ├── CountdownTimer.vue     # 倒计时组件
    └── PayStatus.vue          # 支付状态轮询
```

---

## 10. 部署架构

### 10.1 Docker Compose 完整配置

```yaml
# deploy/docker-compose.yml
version: "3.9"

services:
  postgres:
    image: postgres:16-alpine
    container_name: epay_postgres
    environment:
      POSTGRES_DB: epay
      POSTGRES_USER: epay
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - pg_data:/var/lib/postgresql/data
      - ./postgres/init.sql:/docker-entrypoint-initdb.d/init.sql:ro
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U epay"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped
    networks:
      - epay_internal

  redis:
    image: redis:7-alpine
    container_name: epay_redis
    command: >
      redis-server
      --requirepass ${REDIS_PASSWORD}
      --maxmemory 512mb
      --maxmemory-policy allkeys-lru
      --notify-keyspace-events Ex
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "-a", "${REDIS_PASSWORD}", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped
    networks:
      - epay_internal

  backend:
    build:
      context: ..
      dockerfile: deploy/Dockerfile
      args:
        - BUILD_VERSION=${VERSION:-latest}
    container_name: epay_backend
    environment:
      APP_ENV: production
      APP_PORT: 8080
      DB_DSN: postgres://epay:${DB_PASSWORD}@postgres:5432/epay?sslmode=disable
      REDIS_ADDR: redis:6379
      REDIS_PASSWORD: ${REDIS_PASSWORD}
      JWT_SECRET: ${JWT_SECRET}
      CONFIG_FILE: /app/configs/config.yaml
    volumes:
      - ../configs/config.prod.yaml:/app/configs/config.yaml:ro
      - upload_data:/app/uploads
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    restart: unless-stopped
    networks:
      - epay_internal

  nginx:
    image: nginx:1.25-alpine
    container_name: epay_nginx
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/conf.d:/etc/nginx/conf.d:ro
      - ../frontend/dist:/usr/share/nginx/html:ro
      - cert_data:/etc/nginx/certs
      - nginx_logs:/var/log/nginx
    depends_on:
      - backend
    restart: unless-stopped
    networks:
      - epay_internal
      - epay_external

  prometheus:
    image: prom/prometheus:latest
    container_name: epay_prometheus
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus_data:/prometheus
    restart: unless-stopped
    networks:
      - epay_internal

  grafana:
    image: grafana/grafana:latest
    container_name: epay_grafana
    environment:
      GF_SECURITY_ADMIN_PASSWORD: ${GRAFANA_PASSWORD}
    volumes:
      - grafana_data:/var/lib/grafana
      - ./grafana/dashboard.json:/etc/grafana/provisioning/dashboards/epay.json:ro
    ports:
      - "3000:3000"
    restart: unless-stopped
    networks:
      - epay_internal

volumes:
  pg_data:
  redis_data:
  upload_data:
  cert_data:
  nginx_logs:
  prometheus_data:
  grafana_data:

networks:
  epay_internal:
    internal: true
  epay_external:
```

### 10.2 多阶段 Dockerfile

```dockerfile
# deploy/Dockerfile

# ── 阶段一：前端构建 ──
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# ── 阶段二：后端编译 ──
FROM golang:1.22-alpine AS backend-builder
WORKDIR /app
RUN apk add --no-cache gcc musl-dev
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG BUILD_VERSION=latest
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.Version=${BUILD_VERSION}" \
    -o epay-server ./cmd/server

# ── 阶段三：最终镜像 ──
FROM alpine:3.19
WORKDIR /app
RUN apk add --no-cache tzdata ca-certificates curl
COPY --from=backend-builder /app/epay-server .
COPY --from=frontend-builder /app/frontend/dist ./static
COPY configs/ ./configs/
ENV TZ=Asia/Shanghai
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=10s \
  CMD curl -f http://localhost:8080/health || exit 1
CMD ["./epay-server"]
```

### 10.3 Nginx 配置

```nginx
# deploy/nginx/conf.d/epay.conf
upstream epay_backend {
    server backend:8080;
    keepalive 32;
}

server {
    listen 80;
    server_name _;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your.domain.com;

    ssl_certificate     /etc/nginx/certs/fullchain.pem;
    ssl_certificate_key /etc/nginx/certs/privkey.pem;
    ssl_protocols       TLSv1.2 TLSv1.3;
    ssl_ciphers         HIGH:!aNULL:!MD5;

    # 安全头
    add_header X-Frame-Options SAMEORIGIN;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains";

    # 前端静态资源（带缓存）
    root /usr/share/nginx/html;
    index index.html;

    location ~* \.(js|css|png|jpg|ico|woff2?)$ {
        expires 30d;
        add_header Cache-Control "public, immutable";
    }

    # Vue SPA 路由支持
    location / {
        try_files $uri $uri/ /index.html;
    }

    # 后端 API（不缓存）
    location /api/ {
        proxy_pass http://epay_backend;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 10s;
        proxy_read_timeout 30s;
    }

    # 支付下单（限流：同 IP 每分钟 20 次）
    location /submit {
        limit_req zone=submit_limit burst=5 nodelay;
        proxy_pass http://epay_backend;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # 第三方支付回调（放宽超时）
    location /pay/notify/ {
        proxy_pass http://epay_backend;
        proxy_read_timeout 60s;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # 健康检查
    location /health {
        proxy_pass http://epay_backend;
        access_log off;
    }
}

# 限流区域定义（在 http 块中）
# limit_req_zone $binary_remote_addr zone=submit_limit:10m rate=20r/m;
```

---

## 11. 安全设计

### 11.1 签名机制

**MD5 签名（默认，兼容原协议）**

```
1. 将所有非空参数（除 sign 和 sign_type）按参数名 ASCII 升序排列
2. 拼接成 key=value&key=value 格式
3. 末尾追加 &key={api_key}
4. 对整个字符串做 MD5，结果转小写
```

**RSA 签名（高安全模式）**

```
1. 同 MD5 步骤 1-2 拼接参数字符串
2. 商户使用自己的私钥对字符串做 SHA256withRSA 签名
3. 服务器使用商户预先上传的公钥验证签名
```

### 11.2 敏感数据保护

- 支付通道配置（AppID、AppSecret、商户私钥）：AES-256-GCM 加密后存数据库
- 用户密码：bcrypt（cost=12）哈希存储，禁止明文
- JWT Secret：从环境变量注入，不写入代码或配置文件
- 数据库连接字符串：环境变量注入

### 11.3 接口防护

- 所有 Admin API 强制 JWT 鉴权
- 所有商户 API 强制签名校验
- 下单接口全局限流（IP 维度）
- 敏感操作（重置密钥、提现）需额外验证（邮件验证码/二次密码）
- 回调通知接口仅允许第三方 IP 段访问（可配置白名单）

---

## 12. 性能设计

### 12.1 缓存策略

| 缓存对象 | Key 格式 | TTL | 更新策略 |
|--------|--------|-----|--------|
| 商户信息 | `cache:merchant:{id}` | 5 分钟 | 写时更新 |
| 通道列表 | `cache:channels:{type}` | 2 分钟 | 写时清除 |
| 订单状态（轮询用） | `cache:order_status:{order_no}` | 30 秒 | 状态变更时更新 |
| 系统配置 | `cache:sys_config` | 5 分钟 | 写时更新 |
| IP 黑名单 | `blacklist:ip:{ip}` | 5 分钟 | 写时更新 |

### 12.2 数据库优化

- 订单表按月分区（PostgreSQL 声明式分区），超过 3 个月的数据归档
- 关键查询字段建立复合索引：`(merchant_id, status, created_at)`
- 读多写少的配置、商户信息通过读写分离减轻主库压力
- GORM 使用 `Preload` 替代 N+1 查询

### 12.3 并发控制

- 余额操作：Redis 分布式锁 + 数据库 `FOR UPDATE` 悲观锁
- 重复回调：Redis `SETNX` 幂等键
- 超卖控制（通道日限额）：Redis 原子 `INCRBY` + 对比限额

---

## 13. 监控与日志

### 13.1 日志规范

```go
// 结构化日志示例
zap.L().Info("order created",
    zap.String("order_no", order.OrderNo),
    zap.Int64("merchant_id", order.MerchantID),
    zap.String("channel", channel.Code),
    zap.Float64("amount", order.Amount.InexactFloat64()),
    zap.Duration("duration", time.Since(start)),
)
```

日志级别：`DEBUG`（开发）→ `INFO`（生产默认）→ `WARN`（异常）→ `ERROR`（需告警）

### 13.2 Prometheus 指标

```go
// 核心指标
var (
    OrderCreateTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{Name: "epay_order_create_total"},
        []string{"channel", "status"},
    )
    OrderCreateDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{Name: "epay_order_create_duration_seconds"},
        []string{"channel"},
    )
    NotifySuccessRate = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{Name: "epay_notify_success_rate"},
        []string{"merchant_id"},
    )
    ChannelAvailability = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{Name: "epay_channel_availability"},
        []string{"channel"},
    )
)
```

### 13.3 告警规则

| 告警名 | 触发条件 | 级别 |
|--------|--------|------|
| 通道成功率骤降 | 5 分钟内成功率 < 80% | Critical |
| 回调队列积压 | Redis `queue:notify` 长度 > 1000 | Warning |
| 订单创建失败率 | 5 分钟内失败率 > 10% | Warning |
| 数据库慢查询 | 查询耗时 > 500ms | Warning |
| 服务实例不健康 | 健康检查连续失败 3 次 | Critical |

---

## 14. 开发规范

### 14.1 分支管理

```
main          生产分支，只接受 PR，受保护
develop       开发集成分支
feature/*     功能分支，命名如 feature/channel-alipay
fix/*         Bug 修复分支
release/*     预发布分支
```

### 14.2 提交规范（Conventional Commits）

```
feat(channel): 新增支付宝 H5 支付支持
fix(notify): 修复回调重试间隔计算错误
refactor(order): 重构订单状态机实现
test(risk): 增加 IP 频率限制单元测试
docs: 更新 API 文档
```

### 14.3 错误处理规范

```go
// 定义业务错误
var (
    ErrMerchantNotFound    = errors.New("merchant not found")
    ErrInvalidSign         = errors.New("invalid signature")
    ErrNoAvailableChannel  = errors.New("no available channel")
    ErrInsufficientBalance = errors.New("insufficient balance")
    ErrIPRateLimitExceeded = errors.New("ip rate limit exceeded")
)

// Handler 层统一错误处理
func handleError(c *gin.Context, err error) {
    switch {
    case errors.Is(err, ErrInvalidSign):
        c.JSON(400, response.Fail(1002, err.Error()))
    case errors.Is(err, ErrMerchantNotFound):
        c.JSON(400, response.Fail(1003, err.Error()))
    case errors.Is(err, ErrIPRateLimitExceeded):
        c.JSON(429, response.Fail(2001, err.Error()))
    default:
        zap.L().Error("unexpected error", zap.Error(err))
        c.JSON(500, response.Fail(5000, "internal server error"))
    }
}
```

### 14.4 Makefile 常用命令

```makefile
.PHONY: dev build test lint migrate docker-up docker-down

dev:                        ## 启动开发环境
	go run ./cmd/server -config configs/config.yaml

build:                      ## 编译二进制
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/epay ./cmd/server

test:                       ## 运行单元测试
	go test ./... -v -count=1 -race

lint:                       ## 代码检查
	golangci-lint run ./...

migrate-up:                 ## 执行数据库迁移
	migrate -path migrations -database "${DB_DSN}" up

migrate-down:               ## 回滚最后一次迁移
	migrate -path migrations -database "${DB_DSN}" down 1

docker-up:                  ## 启动全部容器
	docker compose -f deploy/docker-compose.yml up -d

docker-down:                ## 停止全部容器
	docker compose -f deploy/docker-compose.yml down

docker-build:               ## 构建镜像
	docker compose -f deploy/docker-compose.yml build --no-cache

gen-key:                    ## 生成 RSA 密钥对
	bash scripts/gen_key.sh
```

---

*文档版本：v1.0.0 · 最后更新：2026-04*
