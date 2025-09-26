#!/bin/bash

# Teable Go Backend 启动脚本

set -e

echo "🚀 启动 Teable Go Backend..."

# 检查Go是否安装
if ! command -v go &> /dev/null; then
    echo "❌ Go 未安装，请先安装 Go 1.19+"
    exit 1
fi

# 检查Go版本
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_VERSION="1.19"

if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    echo "❌ Go 版本过低，需要 1.19+，当前版本: $GO_VERSION"
    exit 1
fi

# 进入Go后端目录
cd "$(dirname "$0")/teable-go-backend"

echo "📁 当前目录: $(pwd)"

# 检查配置文件
if [ ! -f "config.yaml" ]; then
    echo "❌ 配置文件 config.yaml 不存在"
    exit 1
fi

# 检查依赖
echo "📦 检查Go模块依赖..."
if [ ! -f "go.mod" ]; then
    echo "❌ go.mod 文件不存在"
    exit 1
fi

# 下载依赖
echo "⬇️  下载Go模块依赖..."
go mod download

# 检查数据库连接
echo "🗄️  检查数据库连接..."
if ! pg_isready -h localhost -p 5433 &> /dev/null; then
    echo "⚠️  警告: PostgreSQL 数据库未运行或无法连接"
    echo "   请确保 PostgreSQL 在 localhost:5433 运行"
fi

# 检查Redis连接
echo "🔴 检查Redis连接..."
if ! redis-cli -h localhost -p 6380 ping &> /dev/null; then
    echo "⚠️  警告: Redis 未运行或无法连接"
    echo "   请确保 Redis 在 localhost:6380 运行"
fi

# 构建应用
echo "🔨 构建Go应用..."
go build -o teable-go-backend ./cmd/server

# 启动应用
echo "🎯 启动 Teable Go Backend 服务..."
echo "   服务地址: http://localhost:3000"
echo "   健康检查: http://localhost:3000/health"
echo "   API文档: http://localhost:3000/swagger/index.html"
echo ""
echo "按 Ctrl+C 停止服务"
echo ""

./teable-go-backend
