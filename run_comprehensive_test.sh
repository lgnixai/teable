#!/bin/bash

# Teable Go Backend 综合测试脚本

set -e

echo "🧪 Teable Go Backend 综合测试开始..."
echo "=================================="

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."
    
    # 检查Python
    if ! command -v python3 &> /dev/null; then
        log_error "Python3 未安装"
        exit 1
    fi
    
    # 检查pip
    if ! command -v pip3 &> /dev/null; then
        log_error "pip3 未安装"
        exit 1
    fi
    
    # 检查Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装"
        exit 1
    fi
    
    # 检查Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose 未安装"
        exit 1
    fi
    
    log_success "所有依赖检查通过"
}

# 安装Python依赖
install_python_deps() {
    log_info "安装Python依赖..."
    
    # 创建requirements.txt
    cat > requirements.txt << EOF
requests>=2.28.0
statistics
concurrent.futures
argparse
dataclasses
EOF
    
    pip3 install -r requirements.txt
    log_success "Python依赖安装完成"
}

# 启动测试环境
start_test_environment() {
    log_info "启动测试环境..."
    
    # 停止可能存在的容器
    docker-compose -f docker-compose.test.yml down 2>/dev/null || true
    
    # 启动服务
    docker-compose -f docker-compose.test.yml up -d postgres redis
    
    # 等待服务启动
    log_info "等待数据库和Redis启动..."
    sleep 10
    
    # 检查服务状态
    if ! docker-compose -f docker-compose.test.yml ps | grep -q "Up"; then
        log_error "测试环境启动失败"
        exit 1
    fi
    
    log_success "测试环境启动成功"
}

# 构建并启动Go后端
start_go_backend() {
    log_info "构建并启动Go后端..."
    
    cd teable-go-backend
    
    # 检查Go是否安装
    if ! command -v go &> /dev/null; then
        log_error "Go 未安装，请先安装 Go 1.19+"
        exit 1
    fi
    
    # 下载依赖
    go mod download
    
    # 构建应用
    go build -o teable-go-backend ./cmd/server
    
    # 启动应用（后台运行）
    ./teable-go-backend &
    GO_BACKEND_PID=$!
    
    cd ..
    
    # 等待服务启动
    log_info "等待Go后端启动..."
    sleep 5
    
    # 检查服务是否启动
    if ! curl -f http://localhost:3000/health &>/dev/null; then
        log_error "Go后端启动失败"
        kill $GO_BACKEND_PID 2>/dev/null || true
        exit 1
    fi
    
    log_success "Go后端启动成功 (PID: $GO_BACKEND_PID)"
    echo $GO_BACKEND_PID > go_backend.pid
}

# 运行基础功能测试
run_basic_tests() {
    log_info "运行基础功能测试..."
    
    python3 test_go_backend.py --test-type basic --output basic_test_report.json
    
    if [ $? -eq 0 ]; then
        log_success "基础功能测试通过"
    else
        log_error "基础功能测试失败"
        return 1
    fi
}

# 运行性能测试
run_performance_tests() {
    log_info "运行性能测试..."
    
    python3 test_go_backend.py --test-type performance --requests 200 --concurrent 20 --output performance_test_report.json
    
    if [ $? -eq 0 ]; then
        log_success "性能测试完成"
    else
        log_warning "性能测试部分失败"
    fi
}

# 运行对比测试（如果有NestJS后端）
run_comparison_tests() {
    log_info "检查NestJS后端..."
    
    # 检查NestJS后端是否运行
    if curl -f http://localhost:3001/health &>/dev/null; then
        log_info "发现NestJS后端，运行对比测试..."
        python3 compare_backends.py --output comparison_report.json
        
        if [ $? -eq 0 ]; then
            log_success "对比测试完成"
        else
            log_warning "对比测试部分失败"
        fi
    else
        log_warning "NestJS后端未运行，跳过对比测试"
    fi
}

# 生成综合报告
generate_comprehensive_report() {
    log_info "生成综合测试报告..."
    
    # 创建报告目录
    mkdir -p test_reports
    
    # 移动报告文件
    mv basic_test_report.json test_reports/ 2>/dev/null || true
    mv performance_test_report.json test_reports/ 2>/dev/null || true
    mv comparison_report.json test_reports/ 2>/dev/null || true
    
    # 生成HTML报告
    cat > test_reports/index.html << 'EOF'
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Teable Go Backend 测试报告</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f0f0f0; padding: 20px; border-radius: 5px; }
        .section { margin: 20px 0; }
        .success { color: green; }
        .error { color: red; }
        .warning { color: orange; }
        pre { background: #f5f5f5; padding: 10px; border-radius: 3px; overflow-x: auto; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Teable Go Backend 综合测试报告</h1>
        <p>生成时间: $(date)</p>
    </div>
    
    <div class="section">
        <h2>测试概述</h2>
        <p>本报告包含了Teable Go Backend的全面测试结果，包括功能测试、性能测试和与原版NestJS后端的对比测试。</p>
    </div>
    
    <div class="section">
        <h2>报告文件</h2>
        <ul>
            <li><a href="basic_test_report.json">基础功能测试报告</a></li>
            <li><a href="performance_test_report.json">性能测试报告</a></li>
            <li><a href="comparison_report.json">对比测试报告</a></li>
        </ul>
    </div>
    
    <div class="section">
        <h2>测试环境</h2>
        <ul>
            <li>Go版本: $(go version 2>/dev/null || echo "未安装")</li>
            <li>Python版本: $(python3 --version 2>/dev/null || echo "未安装")</li>
            <li>Docker版本: $(docker --version 2>/dev/null || echo "未安装")</li>
            <li>测试时间: $(date)</li>
        </ul>
    </div>
</body>
</html>
EOF
    
    log_success "综合测试报告已生成到 test_reports/ 目录"
}

# 清理环境
cleanup() {
    log_info "清理测试环境..."
    
    # 停止Go后端
    if [ -f go_backend.pid ]; then
        GO_PID=$(cat go_backend.pid)
        kill $GO_PID 2>/dev/null || true
        rm -f go_backend.pid
    fi
    
    # 停止Docker服务
    docker-compose -f docker-compose.test.yml down 2>/dev/null || true
    
    # 清理临时文件
    rm -f requirements.txt
    
    log_success "环境清理完成"
}

# 主函数
main() {
    # 设置错误处理
    trap cleanup EXIT
    
    # 检查参数
    if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
        echo "用法: $0 [选项]"
        echo "选项:"
        echo "  --help, -h     显示帮助信息"
        echo "  --basic-only   仅运行基础功能测试"
        echo "  --perf-only    仅运行性能测试"
        echo "  --no-cleanup   不清理测试环境"
        exit 0
    fi
    
    # 检查依赖
    check_dependencies
    
    # 安装Python依赖
    install_python_deps
    
    # 启动测试环境
    start_test_environment
    
    # 启动Go后端
    start_go_backend
    
    # 根据参数运行测试
    if [ "$1" = "--basic-only" ]; then
        run_basic_tests
    elif [ "$1" = "--perf-only" ]; then
        run_performance_tests
    else
        # 运行所有测试
        run_basic_tests
        run_performance_tests
        run_comparison_tests
    fi
    
    # 生成综合报告
    generate_comprehensive_report
    
    log_success "所有测试完成！"
    log_info "查看测试报告: test_reports/index.html"
}

# 运行主函数
main "$@"
