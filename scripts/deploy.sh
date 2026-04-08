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

# 拉取镜像
echo "拉取 Docker 镜像..."
docker-compose pull backend frontend

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
