-- 初始化数据库脚本

-- 创建数据库(如果不存在)
-- CREATE DATABASE IF NOT EXISTS teable;

-- 使用数据库
-- \c teable;

-- 创建扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- 创建基础表(这里只创建用户相关表作为示例)
-- 实际项目中应该使用migration工具

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(30) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255),
    salt VARCHAR(255),
    phone VARCHAR(50) UNIQUE,
    avatar VARCHAR(500),
    is_system BOOLEAN DEFAULT FALSE,
    is_admin BOOLEAN DEFAULT FALSE,
    is_trial_used BOOLEAN DEFAULT FALSE,
    notify_meta TEXT,
    last_sign_time TIMESTAMP,
    deactivated_time TIMESTAMP,
    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_time TIMESTAMP,
    last_modified_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    permanent_deleted_time TIMESTAMP,
    ref_meta TEXT
);

-- 账户表(第三方登录)
CREATE TABLE IF NOT EXISTS account (
    id VARCHAR(30) PRIMARY KEY,
    user_id VARCHAR(30) NOT NULL,
    type VARCHAR(50) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    provider_id VARCHAR(255) NOT NULL,
    created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(provider, provider_id)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone);
CREATE INDEX IF NOT EXISTS idx_users_created_time ON users(created_time);
CREATE INDEX IF NOT EXISTS idx_users_is_admin ON users(is_admin);
CREATE INDEX IF NOT EXISTS idx_users_deactivated_time ON users(deactivated_time);

CREATE INDEX IF NOT EXISTS idx_account_user_id ON account(user_id);
CREATE INDEX IF NOT EXISTS idx_account_provider ON account(provider, provider_id);

-- 插入默认系统用户
INSERT INTO users (
    id, 
    name, 
    email, 
    password,
    is_system, 
    is_admin,
    created_time
) VALUES (
    'usr_system_admin_000001',
    'System Admin',
    'admin@teable.ai',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- password: "admin123"
    true,
    true,
    CURRENT_TIMESTAMP
) ON CONFLICT (id) DO NOTHING;

-- 插入测试用户
INSERT INTO users (
    id,
    name,
    email,
    password,
    is_system,
    is_admin,
    created_time
) VALUES (
    'usr_test_user_0000001',
    'Test User',
    'test@teable.ai', 
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- password: "test123"
    false,
    false,
    CURRENT_TIMESTAMP
) ON CONFLICT (id) DO NOTHING;