# Teable Go Backend 测试指南

本文档介绍如何全面测试重构的Golang后端，并与原版NestJS API进行对比。

## 📋 测试概述

本测试套件包含以下组件：

1. **基础功能测试** - 测试所有API端点的基本功能
2. **性能测试** - 测试响应时间和并发处理能力
3. **对比测试** - 与NestJS后端进行性能和功能对比
4. **综合测试报告** - 生成详细的测试报告

## 🛠️ 环境要求

### 必需依赖
- **Go 1.19+** - 用于构建和运行Go后端
- **Python 3.7+** - 用于运行测试脚本
- **Docker & Docker Compose** - 用于启动测试环境
- **PostgreSQL** - 数据库服务
- **Redis** - 缓存服务

### 可选依赖
- **NestJS后端** - 用于对比测试（运行在3001端口）

## 🚀 快速开始

### 1. 一键运行综合测试

```bash
# 运行所有测试
./run_comprehensive_test.sh

# 仅运行基础功能测试
./run_comprehensive_test.sh --basic-only

# 仅运行性能测试
./run_comprehensive_test.sh --perf-only
```

### 2. 手动运行测试

#### 启动测试环境
```bash
# 启动PostgreSQL和Redis
docker-compose -f docker-compose.test.yml up -d postgres redis

# 等待服务启动
sleep 10
```

#### 启动Go后端
```bash
# 方式1: 使用启动脚本
./start_go_backend.sh

# 方式2: 手动启动
cd teable-go-backend
go mod download
go build -o teable-go-backend ./cmd/server
./teable-go-backend
```

#### 运行测试
```bash
# 基础功能测试
python3 test_go_backend.py --test-type basic

# 性能测试
python3 test_go_backend.py --test-type performance --requests 200 --concurrent 20

# 对比测试（需要NestJS后端运行在3001端口）
python3 compare_backends.py --go-url http://localhost:3000 --nestjs-url http://localhost:3001
```

## 📊 测试详情

### 基础功能测试

测试以下API端点：

#### 认证相关
- `GET /health` - 健康检查
- `GET /ping` - Ping测试
- `POST /api/auth/register` - 用户注册
- `POST /api/auth/login` - 用户登录
- `POST /api/auth/logout` - 用户登出

#### 用户管理
- `GET /api/users/profile` - 获取用户资料
- `PUT /api/users/profile` - 更新用户资料
- `POST /api/users/change-password` - 修改密码

#### 空间管理
- `POST /api/spaces` - 创建空间
- `GET /api/spaces` - 获取空间列表
- `GET /api/spaces/:id` - 获取单个空间
- `PUT /api/spaces/:id` - 更新空间
- `DELETE /api/spaces/:id` - 删除空间

#### 管理员功能
- `GET /api/admin/users` - 获取用户列表
- `GET /api/admin/users/:id` - 获取单个用户

### 性能测试

- **响应时间测试** - 测量每个端点的平均响应时间
- **并发测试** - 测试多用户并发访问的性能
- **负载测试** - 测试高负载下的系统表现

### 对比测试

与NestJS后端进行以下对比：
- 响应时间对比
- 成功率对比
- 功能完整性对比
- 性能提升分析

## 📈 测试报告

### 报告文件

测试完成后会生成以下报告文件：

- `test_reports/basic_test_report.json` - 基础功能测试报告
- `test_reports/performance_test_report.json` - 性能测试报告
- `test_reports/comparison_report.json` - 对比测试报告
- `test_reports/index.html` - 综合测试报告（HTML格式）

### 报告内容

#### 基础功能测试报告
```json
{
  "test_summary": {
    "total_tests": 15,
    "successful_tests": 14,
    "failed_tests": 1,
    "success_rate": 93.33,
    "avg_response_time": 0.045,
    "min_response_time": 0.012,
    "max_response_time": 0.156
  },
  "endpoint_statistics": {
    "/health": {
      "success_rate": 100.0,
      "avg_response_time": 0.023
    }
  }
}
```

#### 性能测试报告
```json
{
  "performance_metrics": [
    {
      "endpoint": "/api/users/profile",
      "avg_response_time": 0.045,
      "success_rate": 98.5,
      "total_requests": 100,
      "successful_requests": 98
    }
  ]
}
```

#### 对比测试报告
```json
{
  "summary": {
    "go_avg_response_time": 0.045,
    "nestjs_avg_response_time": 0.067,
    "overall_performance_improvement": 32.8
  }
}
```

## 🔧 配置选项

### 测试脚本参数

#### test_go_backend.py
```bash
python3 test_go_backend.py [选项]

选项:
  --url URL           后端服务URL (默认: http://localhost:3000)
  --test-type TYPE    测试类型: basic, performance, all (默认: all)
  --requests NUM      性能测试请求数 (默认: 100)
  --concurrent NUM    性能测试并发用户数 (默认: 10)
  --output FILE       输出报告到文件
```

#### compare_backends.py
```bash
python3 compare_backends.py [选项]

选项:
  --go-url URL        Go后端服务URL (默认: http://localhost:3000)
  --nestjs-url URL    NestJS后端服务URL (默认: http://localhost:3001)
  --output FILE       输出报告到文件
```

### 环境变量

可以通过环境变量配置测试：

```bash
# Go后端配置
export TEABLE_SERVER_HOST=0.0.0.0
export TEABLE_SERVER_PORT=3000
export TEABLE_DATABASE_HOST=localhost
export TEABLE_DATABASE_PORT=5432
export TEABLE_DATABASE_USER=postgres
export TEABLE_DATABASE_PASSWORD=postgres
export TEABLE_DATABASE_NAME=teable
export TEABLE_REDIS_HOST=localhost
export TEABLE_REDIS_PORT=6379
export TEABLE_JWT_SECRET=test-secret-key
```

## 🐛 故障排除

### 常见问题

#### 1. Go后端启动失败
```bash
# 检查Go版本
go version

# 检查依赖
cd teable-go-backend
go mod download

# 检查配置文件
ls -la config.yaml
```

#### 2. 数据库连接失败
```bash
# 检查PostgreSQL状态
docker-compose -f docker-compose.test.yml ps postgres

# 检查数据库连接
pg_isready -h localhost -p 5432
```

#### 3. Redis连接失败
```bash
# 检查Redis状态
docker-compose -f docker-compose.test.yml ps redis

# 检查Redis连接
redis-cli -h localhost -p 6379 ping
```

#### 4. 测试脚本错误
```bash
# 检查Python版本
python3 --version

# 安装依赖
pip3 install requests

# 检查网络连接
curl -f http://localhost:3000/health
```

### 日志查看

#### Go后端日志
```bash
# 查看Go后端日志
tail -f teable-go-backend.log

# 或者如果使用Docker
docker-compose -f docker-compose.test.yml logs go-backend
```

#### 测试日志
测试脚本会输出详细的日志信息，包括：
- 测试进度
- 错误信息
- 性能指标
- 测试结果

## 📝 测试最佳实践

### 1. 测试前准备
- 确保所有依赖服务正常运行
- 清理之前的测试数据
- 检查网络连接

### 2. 测试执行
- 先运行基础功能测试
- 再进行性能测试
- 最后进行对比测试

### 3. 结果分析
- 关注成功率指标
- 分析响应时间分布
- 识别性能瓶颈

### 4. 问题修复
- 根据测试报告定位问题
- 修复后重新运行测试
- 验证修复效果

## 🤝 贡献指南

### 添加新测试
1. 在相应的测试脚本中添加新的测试方法
2. 更新测试用例列表
3. 更新文档

### 改进测试工具
1. 优化测试性能
2. 增加更多测试指标
3. 改进报告格式

### 报告问题
1. 提供详细的错误信息
2. 包含测试环境信息
3. 提供复现步骤

## 📞 支持

如果您在使用测试工具时遇到问题，请：

1. 查看本文档的故障排除部分
2. 检查测试日志输出
3. 提交Issue到项目仓库

---

**注意**: 本测试套件仅用于开发和测试环境，请勿在生产环境中使用。
