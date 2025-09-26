#!/usr/bin/env python3
"""
Teable Go Backend 全面测试脚本
用于测试重构后的Golang后端并与原版NestJS API进行对比
"""

import requests
import json
import time
import threading
import statistics
from typing import Dict, List, Any, Optional
from dataclasses import dataclass
from concurrent.futures import ThreadPoolExecutor, as_completed
import argparse
import sys

@dataclass
class TestResult:
    """测试结果数据类"""
    endpoint: str
    method: str
    status_code: int
    response_time: float
    success: bool
    error_message: Optional[str] = None
    response_data: Optional[Dict] = None

@dataclass
class PerformanceMetrics:
    """性能指标数据类"""
    endpoint: str
    avg_response_time: float
    min_response_time: float
    max_response_time: float
    success_rate: float
    total_requests: int
    successful_requests: int

class TeableGoBackendTester:
    """Teable Go Backend 测试器"""
    
    def __init__(self, base_url: str = "http://localhost:3000"):
        self.base_url = base_url.rstrip('/')
        self.session = requests.Session()
        self.auth_token = None
        self.test_results: List[TestResult] = []
        self.performance_metrics: List[PerformanceMetrics] = []
        
        # 禁用代理
        self.session.proxies = {'http': None, 'https': None}
        
        # 设置请求头
        self.session.headers.update({
            'Content-Type': 'application/json',
            'User-Agent': 'Teable-Go-Backend-Tester/1.0'
        })
    
    def log(self, message: str, level: str = "INFO"):
        """日志输出"""
        timestamp = time.strftime("%Y-%m-%d %H:%M:%S")
        print(f"[{timestamp}] [{level}] {message}")
    
    def make_request(self, method: str, endpoint: str, data: Dict = None, 
                    params: Dict = None, headers: Dict = None) -> TestResult:
        """发送HTTP请求"""
        url = f"{self.base_url}{endpoint}"
        start_time = time.time()
        
        try:
            # 合并请求头
            request_headers = self.session.headers.copy()
            if headers:
                request_headers.update(headers)
            
            # 添加认证token
            if self.auth_token:
                request_headers['Authorization'] = f'Bearer {self.auth_token}'
            
            # 发送请求
            response = self.session.request(
                method=method,
                url=url,
                json=data,
                params=params,
                headers=request_headers,
                timeout=30,
                proxies={'http': None, 'https': None}
            )
            
            response_time = time.time() - start_time
            
            # 解析响应
            try:
                response_data = response.json() if response.content else None
            except:
                response_data = response.text if response.content else None
            
            return TestResult(
                endpoint=endpoint,
                method=method,
                status_code=response.status_code,
                response_time=response_time,
                success=200 <= response.status_code < 300,
                response_data=response_data
            )
            
        except requests.exceptions.RequestException as e:
            response_time = time.time() - start_time
            return TestResult(
                endpoint=endpoint,
                method=method,
                status_code=0,
                response_time=response_time,
                success=False,
                error_message=str(e)
            )
    
    def test_health_check(self) -> TestResult:
        """测试健康检查端点"""
        self.log("测试健康检查端点...")
        return self.make_request('GET', '/health')
    
    def test_ping(self) -> TestResult:
        """测试ping端点"""
        self.log("测试ping端点...")
        return self.make_request('GET', '/ping')
    
    def test_user_registration(self) -> TestResult:
        """测试用户注册"""
        self.log("测试用户注册...")
        test_user = {
            "name": "Test User",
            "email": f"test_{int(time.time())}@example.com",
            "password": "TestPassword123!"
        }
        result = self.make_request('POST', '/api/auth/register', data=test_user)
        
        # 如果注册成功，保存认证token
        if result.success and result.response_data and 'access_token' in result.response_data:
            self.auth_token = result.response_data['access_token']
            self.log(f"用户注册成功，获得认证token: {self.auth_token[:20]}...")
        
        return result
    
    def test_user_login(self) -> TestResult:
        """测试用户登录"""
        self.log("测试用户登录...")
        login_data = {
            "email": "test@example.com",
            "password": "TestPassword123!"
        }
        result = self.make_request('POST', '/api/auth/login', data=login_data)
        
        # 如果登录成功，保存认证token
        if result.success and result.response_data and 'access_token' in result.response_data:
            self.auth_token = result.response_data['access_token']
            self.log(f"用户登录成功，获得认证token: {self.auth_token[:20]}...")
        
        return result
    
    def test_user_profile(self) -> TestResult:
        """测试获取用户资料"""
        self.log("测试获取用户资料...")
        return self.make_request('GET', '/api/users/profile')
    
    def test_update_user_profile(self) -> TestResult:
        """测试更新用户资料"""
        self.log("测试更新用户资料...")
        update_data = {
            "name": "Updated Test User",
            "phone": "+1234567890"
        }
        return self.make_request('PUT', '/api/users/profile', data=update_data)
    
    def test_change_password(self) -> TestResult:
        """测试修改密码"""
        self.log("测试修改密码...")
        password_data = {
            "old_password": "TestPassword123!",
            "new_password": "NewTestPassword123!"
        }
        return self.make_request('POST', '/api/users/change-password', data=password_data)
    
    def test_create_space(self) -> TestResult:
        """测试创建空间"""
        self.log("测试创建空间...")
        space_data = {
            "name": "Test Space",
            "description": "A test space for testing purposes"
        }
        return self.make_request('POST', '/api/spaces', data=space_data)
    
    def test_list_spaces(self) -> TestResult:
        """测试获取空间列表"""
        self.log("测试获取空间列表...")
        return self.make_request('GET', '/api/spaces')
    
    def test_get_space(self, space_id: str = "test-space-id") -> TestResult:
        """测试获取单个空间"""
        self.log(f"测试获取空间: {space_id}")
        return self.make_request('GET', f'/api/spaces/{space_id}')
    
    def test_update_space(self, space_id: str = "test-space-id") -> TestResult:
        """测试更新空间"""
        self.log(f"测试更新空间: {space_id}")
        update_data = {
            "name": "Updated Test Space",
            "description": "Updated description"
        }
        return self.make_request('PUT', f'/api/spaces/{space_id}', data=update_data)
    
    def test_delete_space(self, space_id: str = "test-space-id") -> TestResult:
        """测试删除空间"""
        self.log(f"测试删除空间: {space_id}")
        return self.make_request('DELETE', f'/api/spaces/{space_id}')
    
    def test_admin_list_users(self) -> TestResult:
        """测试管理员获取用户列表"""
        self.log("测试管理员获取用户列表...")
        return self.make_request('GET', '/api/admin/users')
    
    def test_admin_get_user(self, user_id: str = "test-user-id") -> TestResult:
        """测试管理员获取单个用户"""
        self.log(f"测试管理员获取用户: {user_id}")
        return self.make_request('GET', f'/api/admin/users/{user_id}')
    
    def run_basic_functionality_tests(self) -> List[TestResult]:
        """运行基础功能测试"""
        self.log("开始运行基础功能测试...")
        results = []
        
        # 基础端点测试
        results.append(self.test_health_check())
        results.append(self.test_ping())
        
        # 认证相关测试
        results.append(self.test_user_registration())
        results.append(self.test_user_login())
        
        # 用户相关测试
        results.append(self.test_user_profile())
        results.append(self.test_update_user_profile())
        results.append(self.test_change_password())
        
        # 空间相关测试
        results.append(self.test_create_space())
        results.append(self.test_list_spaces())
        results.append(self.test_get_space())
        results.append(self.test_update_space())
        results.append(self.test_delete_space())
        
        # 管理员功能测试
        results.append(self.test_admin_list_users())
        results.append(self.test_admin_get_user())
        
        self.test_results.extend(results)
        return results
    
    def run_performance_test(self, endpoint: str, method: str = 'GET', 
                           data: Dict = None, num_requests: int = 100, 
                           concurrent_users: int = 10) -> PerformanceMetrics:
        """运行性能测试"""
        self.log(f"开始性能测试: {method} {endpoint} ({num_requests} 请求, {concurrent_users} 并发用户)")
        
        response_times = []
        successful_requests = 0
        
        def make_request():
            result = self.make_request(method, endpoint, data)
            response_times.append(result.response_time)
            if result.success:
                nonlocal successful_requests
                successful_requests += 1
            return result
        
        # 使用线程池进行并发测试
        with ThreadPoolExecutor(max_workers=concurrent_users) as executor:
            futures = [executor.submit(make_request) for _ in range(num_requests)]
            
            for future in as_completed(futures):
                try:
                    future.result()
                except Exception as e:
                    self.log(f"性能测试请求失败: {e}", "ERROR")
        
        # 计算性能指标
        if response_times:
            metrics = PerformanceMetrics(
                endpoint=endpoint,
                avg_response_time=statistics.mean(response_times),
                min_response_time=min(response_times),
                max_response_time=max(response_times),
                success_rate=(successful_requests / num_requests) * 100,
                total_requests=num_requests,
                successful_requests=successful_requests
            )
        else:
            metrics = PerformanceMetrics(
                endpoint=endpoint,
                avg_response_time=0,
                min_response_time=0,
                max_response_time=0,
                success_rate=0,
                total_requests=num_requests,
                successful_requests=0
            )
        
        self.performance_metrics.append(metrics)
        return metrics
    
    def run_comprehensive_performance_tests(self) -> List[PerformanceMetrics]:
        """运行全面的性能测试"""
        self.log("开始运行全面的性能测试...")
        metrics = []
        
        # 测试不同端点的性能
        test_cases = [
            ('/health', 'GET', None, 200, 20),
            ('/ping', 'GET', None, 200, 20),
            ('/api/users/profile', 'GET', None, 100, 10),
            ('/api/spaces', 'GET', None, 100, 10),
        ]
        
        for endpoint, method, data, num_requests, concurrent_users in test_cases:
            try:
                metric = self.run_performance_test(
                    endpoint, method, data, num_requests, concurrent_users
                )
                metrics.append(metric)
            except Exception as e:
                self.log(f"性能测试失败 {endpoint}: {e}", "ERROR")
        
        return metrics
    
    def generate_test_report(self) -> Dict[str, Any]:
        """生成测试报告"""
        self.log("生成测试报告...")
        
        # 基础功能测试统计
        total_tests = len(self.test_results)
        successful_tests = sum(1 for r in self.test_results if r.success)
        failed_tests = total_tests - successful_tests
        success_rate = (successful_tests / total_tests * 100) if total_tests > 0 else 0
        
        # 响应时间统计
        response_times = [r.response_time for r in self.test_results if r.success]
        avg_response_time = statistics.mean(response_times) if response_times else 0
        min_response_time = min(response_times) if response_times else 0
        max_response_time = max(response_times) if response_times else 0
        
        # 按端点分组统计
        endpoint_stats = {}
        for result in self.test_results:
            endpoint = result.endpoint
            if endpoint not in endpoint_stats:
                endpoint_stats[endpoint] = {
                    'total': 0,
                    'successful': 0,
                    'failed': 0,
                    'response_times': []
                }
            
            endpoint_stats[endpoint]['total'] += 1
            if result.success:
                endpoint_stats[endpoint]['successful'] += 1
                endpoint_stats[endpoint]['response_times'].append(result.response_time)
            else:
                endpoint_stats[endpoint]['failed'] += 1
        
        # 计算每个端点的成功率
        for endpoint, stats in endpoint_stats.items():
            stats['success_rate'] = (stats['successful'] / stats['total'] * 100) if stats['total'] > 0 else 0
            stats['avg_response_time'] = statistics.mean(stats['response_times']) if stats['response_times'] else 0
        
        report = {
            'test_summary': {
                'total_tests': total_tests,
                'successful_tests': successful_tests,
                'failed_tests': failed_tests,
                'success_rate': round(success_rate, 2),
                'avg_response_time': round(avg_response_time, 3),
                'min_response_time': round(min_response_time, 3),
                'max_response_time': round(max_response_time, 3)
            },
            'endpoint_statistics': endpoint_stats,
            'performance_metrics': [
                {
                    'endpoint': m.endpoint,
                    'avg_response_time': round(m.avg_response_time, 3),
                    'min_response_time': round(m.min_response_time, 3),
                    'max_response_time': round(m.max_response_time, 3),
                    'success_rate': round(m.success_rate, 2),
                    'total_requests': m.total_requests,
                    'successful_requests': m.successful_requests
                }
                for m in self.performance_metrics
            ],
            'failed_tests': [
                {
                    'endpoint': r.endpoint,
                    'method': r.method,
                    'status_code': r.status_code,
                    'error_message': r.error_message,
                    'response_time': round(r.response_time, 3)
                }
                for r in self.test_results if not r.success
            ]
        }
        
        return report
    
    def print_test_report(self, report: Dict[str, Any]):
        """打印测试报告"""
        print("\n" + "="*80)
        print("TEABLE GO BACKEND 测试报告")
        print("="*80)
        
        # 测试摘要
        summary = report['test_summary']
        print(f"\n📊 测试摘要:")
        print(f"   总测试数: {summary['total_tests']}")
        print(f"   成功测试: {summary['successful_tests']}")
        print(f"   失败测试: {summary['failed_tests']}")
        print(f"   成功率: {summary['success_rate']}%")
        print(f"   平均响应时间: {summary['avg_response_time']}s")
        print(f"   最小响应时间: {summary['min_response_time']}s")
        print(f"   最大响应时间: {summary['max_response_time']}s")
        
        # 端点统计
        print(f"\n🔗 端点统计:")
        for endpoint, stats in report['endpoint_statistics'].items():
            print(f"   {endpoint}:")
            print(f"     成功率: {stats['success_rate']:.1f}% ({stats['successful']}/{stats['total']})")
            print(f"     平均响应时间: {stats['avg_response_time']:.3f}s")
        
        # 性能指标
        if report['performance_metrics']:
            print(f"\n⚡ 性能测试结果:")
            for metric in report['performance_metrics']:
                print(f"   {metric['endpoint']}:")
                print(f"     平均响应时间: {metric['avg_response_time']}s")
                print(f"     成功率: {metric['success_rate']}%")
                print(f"     请求数: {metric['total_requests']}")
        
        # 失败测试
        if report['failed_tests']:
            print(f"\n❌ 失败测试详情:")
            for failed in report['failed_tests']:
                print(f"   {failed['method']} {failed['endpoint']}")
                print(f"     状态码: {failed['status_code']}")
                print(f"     错误信息: {failed['error_message']}")
                print(f"     响应时间: {failed['response_time']}s")
        
        print("\n" + "="*80)

def main():
    """主函数"""
    parser = argparse.ArgumentParser(description='Teable Go Backend 测试工具')
    parser.add_argument('--url', default='http://localhost:3000', 
                       help='后端服务URL (默认: http://localhost:3000)')
    parser.add_argument('--test-type', choices=['basic', 'performance', 'all'], 
                       default='all', help='测试类型')
    parser.add_argument('--requests', type=int, default=100, 
                       help='性能测试请求数 (默认: 100)')
    parser.add_argument('--concurrent', type=int, default=10, 
                       help='性能测试并发用户数 (默认: 10)')
    parser.add_argument('--output', help='输出报告到文件')
    
    args = parser.parse_args()
    
    # 创建测试器
    tester = TeableGoBackendTester(args.url)
    
    try:
        if args.test_type in ['basic', 'all']:
            # 运行基础功能测试
            tester.run_basic_functionality_tests()
        
        if args.test_type in ['performance', 'all']:
            # 运行性能测试
            tester.run_comprehensive_performance_tests()
        
        # 生成报告
        report = tester.generate_test_report()
        tester.print_test_report(report)
        
        # 保存报告到文件
        if args.output:
            with open(args.output, 'w', encoding='utf-8') as f:
                json.dump(report, f, indent=2, ensure_ascii=False)
            print(f"\n📄 测试报告已保存到: {args.output}")
        
        # 返回适当的退出码
        if report['test_summary']['failed_tests'] > 0:
            sys.exit(1)
        else:
            sys.exit(0)
            
    except KeyboardInterrupt:
        print("\n\n⚠️  测试被用户中断")
        sys.exit(130)
    except Exception as e:
        print(f"\n\n❌ 测试过程中发生错误: {e}")
        sys.exit(1)

if __name__ == '__main__':
    main()
