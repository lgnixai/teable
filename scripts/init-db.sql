-- Teable 测试数据库初始化脚本

-- 创建测试用户
INSERT INTO users (id, name, email, password_hash, is_active, is_admin, created_time, updated_time)
VALUES (
    'test-user-001',
    'Test User',
    'test@example.com',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- password: TestPassword123!
    true,
    false,
    NOW(),
    NOW()
) ON CONFLICT (email) DO NOTHING;

-- 创建管理员用户
INSERT INTO users (id, name, email, password_hash, is_active, is_admin, created_time, updated_time)
VALUES (
    'admin-user-001',
    'Admin User',
    'admin@example.com',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- password: TestPassword123!
    true,
    true,
    NOW(),
    NOW()
) ON CONFLICT (email) DO NOTHING;

-- 创建测试空间
INSERT INTO spaces (id, name, description, created_by, created_time, updated_time)
VALUES (
    'test-space-001',
    'Test Space',
    'A test space for testing purposes',
    'test-user-001',
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- 创建空间协作者
INSERT INTO space_collaborators (id, space_id, user_id, role, created_time)
VALUES (
    'spcusr-001',
    'test-space-001',
    'test-user-001',
    'owner',
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- 创建测试基础表
INSERT INTO bases (id, name, description, space_id, created_by, created_time, updated_time)
VALUES (
    'test-base-001',
    'Test Base',
    'A test base for testing purposes',
    'test-space-001',
    'test-user-001',
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- 创建测试表元数据
INSERT INTO table_metas (id, name, description, base_id, created_by, created_time, updated_time)
VALUES (
    'test-table-001',
    'Test Table',
    'A test table for testing purposes',
    'test-base-001',
    'test-user-001',
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- 创建测试字段
INSERT INTO fields (id, name, type, table_id, created_by, created_time, updated_time)
VALUES 
    ('test-field-001', 'Name', 'singleLineText', 'test-table-001', 'test-user-001', NOW(), NOW()),
    ('test-field-002', 'Email', 'singleLineText', 'test-table-001', 'test-user-001', NOW(), NOW()),
    ('test-field-003', 'Age', 'number', 'test-table-001', 'test-user-001', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;
