# EPay Go 阶段五：Docker 容器化部署 实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将 EPay Go 系统容器化，实现一键部署

**Architecture:** 使用多阶段构建优化镜像大小，docker-compose 编排多服务（Go后端、Vue前端、PostgreSQL、Redis）

**Tech Stack:** Docker, Docker Compose, Nginx, PostgreSQL, Redis

---

## Task 1: 创建后端 Dockerfile

**Files:**
- Create: `epay-go/Dockerfile`

**Step 1: 创建多阶段构建 Dockerfile**

```dockerfile
# epay-go/Dockerfile
# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 安装依赖
RUN apk add --no-cache git

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o epay-server ./cmd/server

# 运行阶段
FROM alpine:3.19

WORKDIR /app

# 安装必要工具
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 从构建阶段复制二进制文件
COPY --from=builder /app/epay-server .
COPY --from=builder /app/configs ./configs

# 暴露端口
EXPOSE 8080

# 运行
CMD ["./epay-server"]
```

**Step 2: 创建 .dockerignore**

```
# epay-go/.dockerignore
.git
.gitignore
README.md
docs
web/node_modules
web/dist
*.log
*.test
.env.local
```

**Verification:**
```bash
cd epay-go && docker build -t epay-server:test .
```
Expected: 构建成功

---

## Task 2: 创建前端 Dockerfile

**Files:**
- Create: `epay-go/web/Dockerfile`
- Create: `epay-go/web/nginx.conf`

**Step 1: 创建前端 Dockerfile**

```dockerfile
# epay-go/web/Dockerfile
# 构建阶段
FROM node:20-alpine AS builder

WORKDIR /app

# 复制 package.json
COPY package*.json ./
RUN npm ci

# 复制源代码并构建
COPY . .
RUN npm run build

# 运行阶段
FROM nginx:alpine

# 复制构建产物
COPY --from=builder /app/dist /usr/share/nginx/html

# 复制 nginx 配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
```

**Step 2: 创建 Nginx 配置**

```nginx
# epay-go/web/nginx.conf
server {
    listen 80;
    server_name localhost;
    root /usr/share/nginx/html;
    index index.html;

    # Gzip 压缩
    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml;
    gzip_min_length 1000;

    # 静态资源缓存
    location /assets/ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # API 代理
    location /api/ {
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /admin/ {
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }

    location /merchant/ {
        # 区分 API 请求和前端路由
        # API 请求代理到后端
        if ($request_method != GET) {
            proxy_pass http://backend:8080;
        }

        # GET 请求如果是 API 也代理
        set $is_api 0;
        if ($uri ~ ^/merchant/api/) {
            set $is_api 1;
        }
        if ($is_api = 1) {
            proxy_pass http://backend:8080;
        }

        # 其他 GET 请求返回前端页面
        try_files $uri $uri/ /index.html;
    }

    # SPA 路由支持
    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

**Step 3: 创建前端 .dockerignore**

```
# epay-go/web/.dockerignore
node_modules
dist
.git
*.log
.env.local
```

**Verification:**
```bash
cd epay-go/web && docker build -t epay-web:test .
```
Expected: 构建成功

---

## Task 3: 创建 Docker Compose 配置

**Files:**
- Create: `epay-go/docker-compose.yml`
- Create: `epay-go/.env.example`

**Step 1: 创建 docker-compose.yml**

```yaml
# epay-go/docker-compose.yml
version: '3.8'

services:
  # PostgreSQL 数据库
  postgres:
    image: postgres:16-alpine
    container_name: epay-postgres
    restart: unless-stopped
    environment:
      POSTGRES_USER: ${DB_USER:-epay}
      POSTGRES_PASSWORD: ${DB_PASSWORD:-epay123}
      POSTGRES_DB: ${DB_NAME:-epay}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER:-epay}"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Redis 缓存
  redis:
    image: redis:7-alpine
    container_name: epay-redis
    restart: unless-stopped
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Go 后端
  backend:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: epay-backend
    restart: unless-stopped
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=${DB_USER:-epay}
      - DB_PASSWORD=${DB_PASSWORD:-epay123}
      - DB_NAME=${DB_NAME:-epay}
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - JWT_SECRET=${JWT_SECRET:-your-secret-key}
      - GIN_MODE=release
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy

  # Vue 前端
  frontend:
    build:
      context: ./web
      dockerfile: Dockerfile
    container_name: epay-frontend
    restart: unless-stopped
    ports:
      - "80:80"
    depends_on:
      - backend

volumes:
  postgres_data:
  redis_data:
```

**Step 2: 创建环境变量示例文件**

```bash
# epay-go/.env.example
# 数据库配置
DB_USER=epay
DB_PASSWORD=epay123
DB_NAME=epay

# JWT 密钥
JWT_SECRET=your-super-secret-key-change-in-production

# 支付宝配置
ALIPAY_APP_ID=
ALIPAY_PRIVATE_KEY=
ALIPAY_PUBLIC_KEY=

# 微信支付配置
WECHAT_APP_ID=
WECHAT_MCH_ID=
WECHAT_API_KEY=
```

**Verification:**
```bash
cd epay-go && docker-compose config
```
Expected: 配置验证通过

---

## Task 4: 创建生产环境配置

**Files:**
- Create: `epay-go/docker-compose.prod.yml`

**Step 1: 创建生产环境 compose 覆盖文件**

```yaml
# epay-go/docker-compose.prod.yml
version: '3.8'

services:
  postgres:
    environment:
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    # 生产环境不暴露端口
    ports: []

  redis:
    command: redis-server --appendonly yes --requirepass ${REDIS_PASSWORD:-}
    ports: []

  backend:
    environment:
      - REDIS_PASSWORD=${REDIS_PASSWORD:-}
    # 只暴露给内网
    ports: []
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

  frontend:
    ports:
      - "80:80"
      - "443:443"
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

---

## Task 5: 创建部署脚本和文档

**Files:**
- Create: `epay-go/scripts/deploy.sh`
- Create: `epay-go/DEPLOYMENT.md`

**Step 1: 创建部署脚本**

```bash
#!/bin/bash
# epay-go/scripts/deploy.sh

set -e

echo "=== EPay Go 部署脚本 ==="

# 检查环境变量文件
if [ ! -f .env ]; then
    echo "创建 .env 文件..."
    cp .env.example .env
    echo "请编辑 .env 文件配置必要的环境变量"
    exit 1
fi

# 构建镜像
echo "构建 Docker 镜像..."
docker-compose build

# 启动服务
echo "启动服务..."
docker-compose up -d

# 等待数据库就绪
echo "等待数据库就绪..."
sleep 10

# 运行数据库迁移
echo "运行数据库迁移..."
docker-compose exec backend ./epay-server migrate

echo "=== 部署完成 ==="
echo "前端访问: http://localhost"
echo "后端 API: http://localhost:8080"
```

**Step 2: 创建部署文档**

```markdown
# EPay Go 部署指南

## 环境要求

- Docker 20.10+
- Docker Compose 2.0+
- 2GB+ 内存
- 10GB+ 磁盘空间

## 快速开始

### 1. 克隆项目

```bash
git clone <repository-url>
cd epay-go
```

### 2. 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 文件，配置必要的环境变量
```

### 3. 启动服务

```bash
# 开发环境
docker-compose up -d

# 生产环境
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

### 4. 访问服务

- 前端: http://localhost
- 后端 API: http://localhost:8080
- 管理后台: http://localhost/admin/login
- 商户中心: http://localhost/merchant/login

## 常用命令

```bash
# 查看日志
docker-compose logs -f

# 查看服务状态
docker-compose ps

# 停止服务
docker-compose down

# 重新构建
docker-compose build --no-cache

# 进入容器
docker-compose exec backend sh
docker-compose exec postgres psql -U epay
```

## 数据备份

```bash
# 备份数据库
docker-compose exec postgres pg_dump -U epay epay > backup.sql

# 恢复数据库
cat backup.sql | docker-compose exec -T postgres psql -U epay epay
```

## 更新部署

```bash
# 拉取最新代码
git pull

# 重新构建并启动
docker-compose build
docker-compose up -d
```

## 故障排查

### 后端无法连接数据库

检查 PostgreSQL 是否就绪：
```bash
docker-compose logs postgres
```

### 前端无法访问后端 API

检查 Nginx 代理配置和后端服务状态：
```bash
docker-compose logs frontend
docker-compose logs backend
```
```

**Step 3: 添加执行权限**

```bash
chmod +x scripts/deploy.sh
```

---

## Task 6: 最终验证

**Step 1: 完整构建测试**

```bash
cd epay-go
docker-compose build
```
Expected: 所有镜像构建成功

**Step 2: 启动服务测试**

```bash
docker-compose up -d
docker-compose ps
```
Expected: 所有服务状态为 Up

**Step 3: 健康检查**

```bash
# 检查后端
curl http://localhost:8080/health

# 检查前端
curl http://localhost
```
Expected: 返回正常响应

**Step 4: 清理测试环境**

```bash
docker-compose down -v
```

---

## 文件结构总览

```
epay-go/
├── Dockerfile              # 后端 Dockerfile
├── .dockerignore           # 后端忽略文件
├── docker-compose.yml      # 开发环境编排
├── docker-compose.prod.yml # 生产环境覆盖
├── .env.example            # 环境变量示例
├── DEPLOYMENT.md           # 部署文档
├── scripts/
│   └── deploy.sh           # 部署脚本
└── web/
    ├── Dockerfile          # 前端 Dockerfile
    ├── .dockerignore       # 前端忽略文件
    └── nginx.conf          # Nginx 配置
```
