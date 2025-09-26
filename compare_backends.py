#!/usr/bin/env python3
"""
Teable 后端对比测试脚本
比较重构的Golang后端与原版NestJS后端的性能和功能
"""

import requests
import json
import time
import statistics
from typing import Dict, List, Any, Optional, Tuple
from dataclasses import dataclass
from concurrent.futures import ThreadPoolExecutor, as_completed
import argparse
import sys

@dataclass
class BackendComparison:
    """后端对比结果"""
    endpoint: str
    method: str
    go_response_time: float
    nestjs_response_time: float
    go_success: bool
    nestjs_success: bool
    go_status_code: int
    nestjs_status_code: int
    performance_improvement: float  # Go相对于NestJS的性能提升百分比

class BackendComparator:
    """后端对比器"""
    
    def __init__(self, go_url: str = "http://localhost:3000", 
                 nestjs_url: str = "http://localhost:3001"):
        self.go_url = go_url.rstrip('/')
        self.nestjs_url = nestjs_url.rstrip('/')
        self.go_session = requests.Session()
        self.nestjs_session = requests.Session()
        self.go_auth_token = None
        self.nestjs_auth_token = None
        
        # 禁用代理
        self.go_session.proxies = {'http': None, 'https': None}
        self.nestjs_session.proxies = {'http': None, 'https': None}
        
        # 设置请求头
        headers = {
            'Content-Type': 'application/json',
            'User-Agent': 'Teable-Backend-Comparator/1.0'
        }
        self.go_session.headers.update(headers)
        self.nestjs_session.headers.update(headers)
    
    def log(self, message: str, level: str = "INFO"):
        """日志输出"""
        timestamp = time.strftime("%Y-%m-%d %H:%M:%S")
        print(f"[{timestamp}] [{level}] {message}")
    
    def make_request(self, session: requests.Session, base_url: str, 
                    method: str, endpoint: str, data: Dict = None, 
                    auth_token: str = None) -> Tuple[float, bool, int, Any]:
        """发送HTTP请求并返回结果"""
        url = f"{base_url}{endpoint}"
        start_time = time.time()
        
        try:
            headers = {}
            if auth_token:
                headers['Authorization'] = f'Bearer {auth_token}'
            
            response = session.request(
                method=method,
                url=url,
                json=data,
                headers=headers,
                timeout=30,
                proxies={'http': None, 'https': None}
            )
            
            response_time = time.time() - start_time
            success = 200 <= response.status_code < 300
            
            try:
                response_data = response.json() if response.content else None
            except:
                response_data = response.text if response.content else None
            
            return response_time, success, response.status_code, response_data
            
        except requests.exceptions.RequestException as e:
            response_time = time.time() - start_time
            return response_time, False, 0, str(e)
    
    def test_endpoint_comparison(self, endpoint: str, method: str = 'GET', 
                               data: Dict = None, num_requests: int = 10) -> BackendComparison:
        """对比测试单个端点"""
        self.log(f"对比测试: {method} {endpoint}")
        
        # 测试Go后端
        go_times = []
        go_successes = 0
        go_status_codes = []
        
        for _ in range(num_requests):
            response_time, success, status_code, _ = self.make_request(
                self.go_session, self.go_url, method, endpoint, data, self.go_auth_token
            )
            go_times.append(response_time)
            if success:
                go_successes += 1
            go_status_codes.append(status_code)
        
        # 测试NestJS后端
        nestjs_times = []
        nestjs_successes = 0
        nestjs_status_codes = []
        
        for _ in range(num_requests):
            response_time, success, status_code, _ = self.make_request(
                self.nestjs_session, self.nestjs_url, method, endpoint, data, self.nestjs_auth_token
            )
            nestjs_times.append(response_time)
            if success:
                nestjs_successes += 1
            nestjs_status_codes.append(status_code)
        
        # 计算平均响应时间
        go_avg_time = statistics.mean(go_times) if go_times else 0
        nestjs_avg_time = statistics.mean(nestjs_times) if nestjs_times else 0
        
        # 计算性能提升
        if nestjs_avg_time > 0:
            performance_improvement = ((nestjs_avg_time - go_avg_time) / nestjs_avg_time) * 100
        else:
            performance_improvement = 0
        
        return BackendComparison(
            endpoint=endpoint,
            method=method,
            go_response_time=go_avg_time,
            nestjs_response_time=nestjs_avg_time,
            go_success=go_successes > 0,
            nestjs_success=nestjs_successes > 0,
            go_status_code=go_status_codes[0] if go_status_codes else 0,
            nestjs_status_code=nestjs_status_codes[0] if nestjs_status_codes else 0,
            performance_improvement=performance_improvement
        )
    
    def setup_authentication(self):
        """设置认证"""
        self.log("设置认证...")
        
        # 尝试登录Go后端
        try:
            login_data = {"email": "test@example.com", "password": "TestPassword123!"}
            response_time, success, status_code, response_data = self.make_request(
                self.go_session, self.go_url, 'POST', '/api/auth/login', login_data
            )
            if success and response_data and 'access_token' in response_data:
                self.go_auth_token = response_data['access_token']
                self.log("Go后端认证成功")
        except Exception as e:
            self.log(f"Go后端认证失败: {e}", "WARN")
        
        # 尝试登录NestJS后端
        try:
            login_data = {"email": "test@example.com", "password": "TestPassword123!"}
            response_time, success, status_code, response_data = self.make_request(
                self.nestjs_session, self.nestjs_url, 'POST', '/api/auth/login', login_data
            )
            if success and response_data and 'access_token' in response_data:
                self.nestjs_auth_token = response_data['access_token']
                self.log("NestJS后端认证成功")
        except Exception as e:
            self.log(f"NestJS后端认证失败: {e}", "WARN")
    
    def run_comprehensive_comparison(self) -> List[BackendComparison]:
        """运行全面的对比测试"""
        self.log("开始全面的后端对比测试...")
        
        # 设置认证
        self.setup_authentication()
        
        # 定义测试用例
        test_cases = [
            # 基础端点
            ('/health', 'GET', None),
            ('/ping', 'GET', None),
            
            # 认证相关
            ('/api/auth/register', 'POST', {
                "name": "Test User",
                "email": f"test_{int(time.time())}@example.com",
                "password": "TestPassword123!"
            }),
            
            # 用户相关
            ('/api/users/profile', 'GET', None),
            ('/api/users/profile', 'PUT', {
                "name": "Updated Test User",
                "phone": "+1234567890"
            }),
            
            # 空间相关
            ('/api/spaces', 'GET', None),
            ('/api/spaces', 'POST', {
                "name": "Test Space",
                "description": "A test space"
            }),
            
            # 管理员功能
            ('/api/admin/users', 'GET', None),
        ]
        
        comparisons = []
        
        for endpoint, method, data in test_cases:
            try:
                comparison = self.test_endpoint_comparison(endpoint, method, data, 5)
                comparisons.append(comparison)
            except Exception as e:
                self.log(f"对比测试失败 {endpoint}: {e}", "ERROR")
        
        return comparisons
    
    def generate_comparison_report(self, comparisons: List[BackendComparison]) -> Dict[str, Any]:
        """生成对比报告"""
        self.log("生成对比报告...")
        
        # 计算总体统计
        total_tests = len(comparisons)
        go_successful = sum(1 for c in comparisons if c.go_success)
        nestjs_successful = sum(1 for c in comparisons if c.nestjs_success)
        
        # 计算平均响应时间
        go_times = [c.go_response_time for c in comparisons if c.go_success]
        nestjs_times = [c.nestjs_response_time for c in comparisons if c.nestjs_success]
        
        go_avg_time = statistics.mean(go_times) if go_times else 0
        nestjs_avg_time = statistics.mean(nestjs_times) if nestjs_times else 0
        
        # 计算性能提升
        overall_improvement = 0
        if nestjs_avg_time > 0:
            overall_improvement = ((nestjs_avg_time - go_avg_time) / nestjs_avg_time) * 100
        
        # 按性能提升排序
        sorted_comparisons = sorted(comparisons, key=lambda x: x.performance_improvement, reverse=True)
        
        report = {
            'summary': {
                'total_tests': total_tests,
                'go_successful': go_successful,
                'nestjs_successful': nestjs_successful,
                'go_success_rate': (go_successful / total_tests * 100) if total_tests > 0 else 0,
                'nestjs_success_rate': (nestjs_successful / total_tests * 100) if total_tests > 0 else 0,
                'go_avg_response_time': round(go_avg_time, 3),
                'nestjs_avg_response_time': round(nestjs_avg_time, 3),
                'overall_performance_improvement': round(overall_improvement, 2)
            },
            'detailed_comparisons': [
                {
                    'endpoint': c.endpoint,
                    'method': c.method,
                    'go_response_time': round(c.go_response_time, 3),
                    'nestjs_response_time': round(c.nestjs_response_time, 3),
                    'go_success': c.go_success,
                    'nestjs_success': c.nestjs_success,
                    'go_status_code': c.go_status_code,
                    'nestjs_status_code': c.nestjs_status_code,
                    'performance_improvement': round(c.performance_improvement, 2)
                }
                for c in sorted_comparisons
            ],
            'best_performers': [
                {
                    'endpoint': c.endpoint,
                    'improvement': round(c.performance_improvement, 2)
                }
                for c in sorted_comparisons[:5] if c.performance_improvement > 0
            ],
            'worst_performers': [
                {
                    'endpoint': c.endpoint,
                    'improvement': round(c.performance_improvement, 2)
                }
                for c in sorted_comparisons[-5:] if c.performance_improvement < 0
            ]
        }
        
        return report
    
    def print_comparison_report(self, report: Dict[str, Any]):
        """打印对比报告"""
        print("\n" + "="*80)
        print("TEABLE 后端对比测试报告")
        print("="*80)
        
        summary = report['summary']
        print(f"\n📊 总体对比:")
        print(f"   总测试数: {summary['total_tests']}")
        print(f"   Go后端成功率: {summary['go_success_rate']:.1f}% ({summary['go_successful']}/{summary['total_tests']})")
        print(f"   NestJS后端成功率: {summary['nestjs_success_rate']:.1f}% ({summary['nestjs_successful']}/{summary['total_tests']})")
        print(f"   Go后端平均响应时间: {summary['go_avg_response_time']}s")
        print(f"   NestJS后端平均响应时间: {summary['nestjs_avg_response_time']}s")
        print(f"   整体性能提升: {summary['overall_performance_improvement']:.1f}%")
        
        if summary['overall_performance_improvement'] > 0:
            print(f"   🚀 Go后端性能更优!")
        elif summary['overall_performance_improvement'] < 0:
            print(f"   ⚠️  NestJS后端性能更优")
        else:
            print(f"   ⚖️  两个后端性能相当")
        
        # 详细对比
        print(f"\n🔍 详细对比结果:")
        for comp in report['detailed_comparisons']:
            status_go = "✅" if comp['go_success'] else "❌"
            status_nestjs = "✅" if comp['nestjs_success'] else "❌"
            improvement = comp['performance_improvement']
            
            if improvement > 0:
                perf_indicator = f"🚀 +{improvement:.1f}%"
            elif improvement < 0:
                perf_indicator = f"⚠️  {improvement:.1f}%"
            else:
                perf_indicator = "⚖️  0%"
            
            print(f"   {comp['method']} {comp['endpoint']}:")
            print(f"     Go: {status_go} {comp['go_response_time']}s (状态码: {comp['go_status_code']})")
            print(f"     NestJS: {status_nestjs} {comp['nestjs_response_time']}s (状态码: {comp['nestjs_status_code']})")
            print(f"     性能对比: {perf_indicator}")
        
        # 最佳表现
        if report['best_performers']:
            print(f"\n🏆 Go后端表现最佳的端点:")
            for perf in report['best_performers']:
                print(f"   {perf['endpoint']}: +{perf['improvement']:.1f}% 性能提升")
        
        # 需要改进的端点
        if report['worst_performers']:
            print(f"\n⚠️  需要改进的端点:")
            for perf in report['worst_performers']:
                print(f"   {perf['endpoint']}: {perf['improvement']:.1f}% 性能下降")
        
        print("\n" + "="*80)

def main():
    """主函数"""
    parser = argparse.ArgumentParser(description='Teable 后端对比测试工具')
    parser.add_argument('--go-url', default='http://localhost:3000', 
                       help='Go后端服务URL (默认: http://localhost:3000)')
    parser.add_argument('--nestjs-url', default='http://localhost:3001', 
                       help='NestJS后端服务URL (默认: http://localhost:3001)')
    parser.add_argument('--output', help='输出报告到文件')
    
    args = parser.parse_args()
    
    # 创建对比器
    comparator = BackendComparator(args.go_url, args.nestjs_url)
    
    try:
        # 运行对比测试
        comparisons = comparator.run_comprehensive_comparison()
        
        # 生成报告
        report = comparator.generate_comparison_report(comparisons)
        comparator.print_comparison_report(report)
        
        # 保存报告到文件
        if args.output:
            with open(args.output, 'w', encoding='utf-8') as f:
                json.dump(report, f, indent=2, ensure_ascii=False)
            print(f"\n📄 对比报告已保存到: {args.output}")
        
        # 返回适当的退出码
        if report['summary']['go_success_rate'] < 80:
            print("\n❌ Go后端成功率过低")
            sys.exit(1)
        else:
            print("\n✅ 对比测试完成")
            sys.exit(0)
            
    except KeyboardInterrupt:
        print("\n\n⚠️  测试被用户中断")
        sys.exit(130)
    except Exception as e:
        print(f"\n\n❌ 对比测试过程中发生错误: {e}")
        sys.exit(1)

if __name__ == '__main__':
    main()
