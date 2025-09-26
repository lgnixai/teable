# Teable 后端系统 Go + PostgreSQL 重构方案

## 一、当前系统分析

### 1.1 技术栈分析

**当前架构：**
- **后端框架**: NestJS (TypeScript)
- **数据库**: PostgreSQL + Redis
- **ORM**: Prisma
- **WebSocket**: Socket.io
- **队列系统**: BullMQ
- **文件存储**: Minio/S3
- **认证**: Passport.js (支持多种策略)
- **API文档**: Swagger/OpenAPI

**项目规模:**
- 41个数据模型
- 超过50个功能模块
- 多租户架构(Space -> Base -> Table)
- 实时协作功能
- 复杂的权限系统

### 1.2 核心功能模块

1. **用户管理系统**
   - 用户注册、登录、权限管理
   - 多种认证方式（Local/OAuth/GitHub/Google）
   - 用户配置和偏好设置

2. **空间和基础管理**
   - 多租户空间管理
   - 基础(Base)的CRUD操作
   - 协作者管理和权限控制

3. **表格系统**
   - 动态表格创建和管理
   - 字段类型系统（文本、数字、日期、选择等）
   - 视图系统（网格、表单、看板、日历、画廊）

4. **数据操作**
   - 记录的增删改查
   - 批量操作
   - 导入导出功能

5. **实时协作**
   - WebSocket实时通信
   - 操作记录和版本控制
   - 评论系统

6. **高级功能**
   - 公式计算引擎
   - 聚合查询
   - 仪表板和图表
   - 插件系统

### 1.3 数据库架构分析

**核心实体关系:**
```
Space (租户空间)
├── Base (数据库)
    ├── TableMeta (表元数据)
        ├── Field (字段定义)
        ├── View (视图定义)
        ├── Record (动态数据表)
        └── Attachments (附件)
    └── Dashboard (仪表板)
├── User (用户)
├── Collaborator (协作者)
└── Permission (权限)
```

**关键特性:**
- 软删除设计
- 审计字段(created_time, modified_time, created_by等)
- 多级权限控制
- 版本控制支持
- 插件扩展支持

## 二、Go + PostgreSQL 重构架构设计

### 2.1 技术栈选择

**核心技术栈:**
- **语言**: Go 1.21+
- **Web框架**: Gin + Fiber (高性能)
- **数据库**: PostgreSQL 15+
- **ORM**: GORM v2 (成熟稳定)
- **缓存**: Redis 7+
- **队列**: Asynq (Go原生队列)
- **WebSocket**: Gorilla WebSocket
- **配置管理**: Viper
- **日志**: Zap
- **监控**: Prometheus + Grafana
- **容器化**: Docker + Docker Compose

**辅助工具:**
- **API文档**: Swag (Swagger for Go)
- **测试**: Testify + GoMock
- **代码生成**: Wire (依赖注入)
- **数据库迁移**: golang-migrate
- **静态分析**: golangci-lint

### 2.2 项目架构设计

采用 **领域驱动设计(DDD) + 清洁架构** 模式：

```
cmd/
├── server/                 # 服务器启动入口
└── migrate/               # 数据库迁移工具

internal/
├── domain/                # 领域层
│   ├── user/             # 用户领域
│   ├── space/            # 空间领域  
│   ├── base/             # 基础领域
│   ├── table/            # 表格领域
│   ├── field/            # 字段领域
│   ├── record/           # 记录领域
│   ├── view/             # 视图领域
│   ├── permission/       # 权限领域
│   └── common/           # 公共领域对象
├── application/          # 应用层
│   ├── services/         # 应用服务
│   ├── handlers/         # HTTP处理器
│   ├── ws/              # WebSocket处理
│   └── dto/             # 数据传输对象
├── infrastructure/       # 基础设施层
│   ├── repository/       # 数据访问层
│   ├── cache/           # 缓存层
│   ├── queue/           # 队列系统
│   ├── storage/         # 文件存储
│   ├── auth/            # 认证系统
│   └── database/        # 数据库配置
├── interfaces/           # 接口层
│   ├── http/            # HTTP API
│   ├── grpc/            # gRPC API (可选)
│   └── middleware/      # 中间件
└── config/              # 配置管理

pkg/                     # 公共包
├── logger/              # 日志包
├── errors/              # 错误处理
├── utils/               # 工具函数
├── validator/           # 验证器
├── encryption/          # 加密工具
└── types/               # 类型定义

api/                     # API定义
├── openapi/             # OpenAPI规范
└── proto/               # Protobuf定义(可选)

scripts/                 # 脚本文件
migrations/              # 数据库迁移文件
deployments/             # 部署配置
```

### 2.3 核心设计模式

**1. 领域驱动设计**
```go
// domain/user/entity.go
type User struct {
    ID        string
    Name      string  
    Email     string
    Password  string
    CreatedAt time.Time
    UpdatedAt time.Time
}

// domain/user/repository.go
type Repository interface {
    Create(ctx context.Context, user *User) error
    FindByID(ctx context.Context, id string) (*User, error)
    FindByEmail(ctx context.Context, email string) (*User, error)
}

// domain/user/service.go  
type Service interface {
    Register(ctx context.Context, req RegisterRequest) (*User, error)
    Login(ctx context.Context, req LoginRequest) (*TokenPair, error)
}
```

**2. 依赖注入 (Wire)**
```go
// wire.go
//go:build wireinject

func InitializeUserService() *application.UserService {
    wire.Build(
        infrastructure.NewUserRepository,
        application.NewUserService,
        infrastructure.NewDatabase,
    )
    return &application.UserService{}
}
```

**3. 中间件模式**
```go
// interfaces/middleware/auth.go
func AuthMiddleware(authService auth.Service) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractToken(c)
        user, err := authService.ValidateToken(token)
        if err != nil {
            c.JSON(401, gin.H{"error": "unauthorized"})
            c.Abort()
            return
        }
        c.Set("user", user)
        c.Next()
    }
}
```

### 2.4 数据库设计优化

**1. GORM模型定义**
```go
// internal/infrastructure/database/models/user.go
type User struct {
    ID                string         `gorm:"primaryKey;type:varchar(30)"`
    Name              string         `gorm:"not null;type:varchar(255)"`
    Email             string         `gorm:"unique;not null;type:varchar(255)"`
    Password          *string        `gorm:"type:varchar(255)"`
    Salt              *string        `gorm:"type:varchar(255)"`
    Phone             *string        `gorm:"unique;type:varchar(50)"`
    Avatar            *string        `gorm:"type:varchar(500)"`
    IsSystem          *bool          `gorm:"column:is_system"`
    IsAdmin           *bool          `gorm:"column:is_admin"`
    IsTrialUsed       *bool          `gorm:"column:is_trial_used"`
    NotifyMeta        *string        `gorm:"type:text;column:notify_meta"`
    LastSignTime      *time.Time     `gorm:"column:last_sign_time"`
    DeactivatedTime   *time.Time     `gorm:"column:deactivated_time"`
    CreatedTime       time.Time      `gorm:"autoCreateTime;column:created_time"`
    DeletedTime       gorm.DeletedAt `gorm:"column:deleted_time"`
    LastModifiedTime  *time.Time     `gorm:"autoUpdateTime;column:last_modified_time"`
    PermanentDeletedTime *time.Time  `gorm:"column:permanent_deleted_time"`
    RefMeta           *string        `gorm:"type:text;column:ref_meta"`

    Accounts []Account `gorm:"foreignKey:UserID"`
}

func (User) TableName() string {
    return "users"
}
```

**2. 数据库连接池优化**
```go
// internal/infrastructure/database/connection.go
func NewDatabase(cfg *config.Database) (*gorm.DB, error) {
    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
        cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port)
    
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
        NamingStrategy: schema.NamingStrategy{
            SingularTable: true,
        },
    })
    if err != nil {
        return nil, err
    }

    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }

    // 连接池配置
    sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)   // 10
    sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)   // 100
    sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime) // 1小时

    return db, nil
}
```

### 2.5 性能优化策略

**1. 查询优化**
```go
// 预加载优化
func (r *userRepository) FindWithAccountsByID(ctx context.Context, id string) (*User, error) {
    var user User
    err := r.db.WithContext(ctx).
        Preload("Accounts").
        Where("id = ?", id).
        First(&user).Error
    return &user, err
}

// 分页查询优化
func (r *userRepository) FindUsers(ctx context.Context, req PaginationRequest) (*PaginationResponse[User], error) {
    var users []User
    var total int64

    query := r.db.WithContext(ctx).Model(&User{})
    
    // 先获取总数
    if err := query.Count(&total).Error; err != nil {
        return nil, err
    }

    // 再获取数据
    offset := (req.Page - 1) * req.PageSize
    if err := query.Offset(offset).Limit(req.PageSize).Find(&users).Error; err != nil {
        return nil, err
    }

    return &PaginationResponse[User]{
        Data:     users,
        Total:    total,
        Page:     req.Page,
        PageSize: req.PageSize,
    }, nil
}
```

**2. 缓存策略**
```go
// internal/infrastructure/cache/redis.go
type RedisCache struct {
    client *redis.Client
}

func (c *RedisCache) GetUser(ctx context.Context, userID string) (*User, error) {
    key := fmt.Sprintf("user:%s", userID)
    data, err := c.client.Get(ctx, key).Result()
    if err == redis.Nil {
        return nil, nil // 缓存未命中
    }
    if err != nil {
        return nil, err
    }

    var user User
    err = json.Unmarshal([]byte(data), &user)
    return &user, err
}

func (c *RedisCache) SetUser(ctx context.Context, user *User, expiration time.Duration) error {
    key := fmt.Sprintf("user:%s", user.ID)
    data, err := json.Marshal(user)
    if err != nil {
        return err
    }
    return c.client.Set(ctx, key, data, expiration).Err()
}
```

### 2.6 WebSocket实时通信

```go
// internal/infrastructure/websocket/hub.go
type Hub struct {
    clients    map[string]*Client
    broadcast  chan []byte
    register   chan *Client
    unregister chan *Client
    mutex      sync.RWMutex
}

type Client struct {
    ID       string
    UserID   string
    SpaceID  string
    TableID  *string
    Conn     *websocket.Conn
    Send     chan []byte
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.mutex.Lock()
            h.clients[client.ID] = client
            h.mutex.Unlock()
            
        case client := <-h.unregister:
            h.mutex.Lock()
            if _, ok := h.clients[client.ID]; ok {
                delete(h.clients, client.ID)
                close(client.Send)
            }
            h.mutex.Unlock()
            
        case message := <-h.broadcast:
            h.mutex.RLock()
            for _, client := range h.clients {
                select {
                case client.Send <- message:
                default:
                    close(client.Send)
                    delete(h.clients, client.ID)
                }
            }
            h.mutex.RUnlock()
        }
    }
}
```

## 三、API设计

### 3.1 RESTful API设计

```go
// internal/interfaces/http/handlers/user.go
type UserHandler struct {
    userService application.UserService
}

// @Summary 用户注册
// @Description 创建新用户账户
// @Tags users
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册信息"
// @Success 201 {object} UserResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/users/register [post]
func (h *UserHandler) Register(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, ErrorResponse{Error: err.Error()})
        return
    }

    user, err := h.userService.Register(c.Request.Context(), req)
    if err != nil {
        c.JSON(400, ErrorResponse{Error: err.Error()})
        return
    }

    c.JSON(201, toUserResponse(user))
}

// @Summary 获取用户信息
// @Description 根据ID获取用户详细信息
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Success 200 {object} UserResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
    userID := c.Param("id")
    
    user, err := h.userService.GetByID(c.Request.Context(), userID)
    if err != nil {
        c.JSON(404, ErrorResponse{Error: "用户不存在"})
        return
    }

    c.JSON(200, toUserResponse(user))
}
```

### 3.2 错误处理

```go
// pkg/errors/errors.go
type Error struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details any    `json:"details,omitempty"`
}

func (e *Error) Error() string {
    return e.Message
}

var (
    ErrUserNotFound     = &Error{Code: "USER_NOT_FOUND", Message: "用户不存在"}
    ErrInvalidPassword  = &Error{Code: "INVALID_PASSWORD", Message: "密码错误"}
    ErrEmailExists      = &Error{Code: "EMAIL_EXISTS", Message: "邮箱已存在"}
    ErrUnauthorized     = &Error{Code: "UNAUTHORIZED", Message: "未授权访问"}
    ErrForbidden        = &Error{Code: "FORBIDDEN", Message: "权限不足"}
    ErrInternalServer   = &Error{Code: "INTERNAL_SERVER_ERROR", Message: "服务器内部错误"}
)
```

## 四、重构迁移计划

### 阶段一：基础设施搭建 (2-3周)
1. **项目脚手架**
   - Go项目结构搭建
   - Docker开发环境
   - CI/CD管道
   - 代码质量工具

2. **数据库层**
   - PostgreSQL数据库设计
   - GORM模型定义
   - 数据库迁移脚本
   - 种子数据

3. **基础服务**
   - 配置管理
   - 日志系统
   - 错误处理
   - 健康检查

### 阶段二：核心域实现 (4-5周)
1. **用户和认证系统**
   - 用户管理
   - JWT认证
   - OAuth集成
   - 权限系统

2. **空间和基础管理**
   - 多租户架构
   - 空间CRUD
   - 基础CRUD
   - 协作者管理

### 阶段三：表格系统 (5-6周)
1. **表格管理**
   - 表元数据管理
   - 字段类型系统
   - 动态表创建
   - 表结构变更

2. **数据操作**
   - 记录CRUD
   - 批量操作
   - 查询优化
   - 数据验证

### 阶段四：高级功能 (4-5周)
1. **视图系统**
   - 多种视图类型
   - 筛选排序
   - 聚合计算
   - 视图配置

2. **实时功能**
   - WebSocket通信
   - 操作同步
   - 冲突解决
   - 版本控制

### 阶段五：扩展功能 (3-4周)
1. **文件和附件**
   - 文件上传
   - 图片处理
   - 存储管理
   - CDN集成

2. **导入导出**
   - CSV/Excel导入
   - 数据导出
   - 模板系统
   - 批量处理

### 阶段六：插件和集成 (2-3周)
1. **插件系统**
   - 插件架构
   - API扩展
   - 配置管理
   - 生命周期

2. **第三方集成**
   - Webhook支持
   - API集成
   - 通知系统
   - 监控告警

## 五、性能和可扩展性

### 5.1 性能优化
- **数据库优化**: 索引优化、查询优化、连接池
- **缓存策略**: Redis缓存、内存缓存、CDN
- **并发处理**: Goroutine池、异步处理
- **资源管理**: 连接复用、内存优化

### 5.2 可扩展性设计
- **水平扩展**: 微服务架构准备
- **负载均衡**: Nginx/HAProxy
- **数据分片**: 按空间分片
- **消息队列**: 异步任务处理

### 5.3 监控和运维
- **指标监控**: Prometheus + Grafana
- **日志聚合**: ELK Stack
- **链路追踪**: Jaeger
- **健康检查**: 自动故障恢复

## 六、总结

本重构方案基于Go生态的最佳实践，采用现代化的架构设计模式，确保：

1. **高性能**: Go的并发优势 + PostgreSQL优化
2. **可维护**: 清洁架构 + 领域驱动设计
3. **可扩展**: 模块化设计 + 微服务准备
4. **高可用**: 容错设计 + 监控告警
5. **功能完整**: 一比一功能实现

重构后的系统将具备更好的性能表现、更强的扩展能力和更高的开发效率。