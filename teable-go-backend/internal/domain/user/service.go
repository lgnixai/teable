package user

import (
	"context"
	"errors"
	"time"
)

// Service 用户领域服务接口
type Service interface {
	// 用户管理
	CreateUser(ctx context.Context, req CreateUserRequest) (*User, error)
	GetUser(ctx context.Context, id string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, id string, req UpdateUserRequest) (*User, error)
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, filter ListFilter) (*PaginatedResult, error)
	
	// 认证相关
	Authenticate(ctx context.Context, email, password string) (*User, error)
	ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error
	ResetPassword(ctx context.Context, email string) error
	
	// 用户状态管理
	ActivateUser(ctx context.Context, userID string) error
	DeactivateUser(ctx context.Context, userID string) error
	
	// 权限管理
	PromoteToAdmin(ctx context.Context, userID string) error
	DemoteFromAdmin(ctx context.Context, userID string) error
	
	// 第三方账户
	LinkAccount(ctx context.Context, userID string, req LinkAccountRequest) error
	UnlinkAccount(ctx context.Context, userID, accountID string) error
	GetUserByProvider(ctx context.Context, provider, providerID string) (*User, error)
}

// ServiceImpl 用户领域服务实现
type ServiceImpl struct {
	repo Repository
}

// NewService 创建用户服务
func NewService(repo Repository) Service {
	return &ServiceImpl{
		repo: repo,
	}
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Name     string  `json:"name" validate:"required,min=1,max=100"`
	Email    string  `json:"email" validate:"required,email,max=255"`
	Password *string `json:"password,omitempty" validate:"omitempty,min=8,max=128"`
	Phone    *string `json:"phone,omitempty" validate:"omitempty,max=50"`
	Avatar   *string `json:"avatar,omitempty" validate:"omitempty,url,max=500"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Name   *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Phone  *string `json:"phone,omitempty" validate:"omitempty,max=50"`
	Avatar *string `json:"avatar,omitempty" validate:"omitempty,url,max=500"`
}

// LinkAccountRequest 关联账户请求
type LinkAccountRequest struct {
	Type       AccountType `json:"type" validate:"required"`
	Provider   Provider    `json:"provider" validate:"required"`
	ProviderID string      `json:"provider_id" validate:"required"`
}

// CreateUser 创建用户
func (s *ServiceImpl) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
	// 检查邮箱是否已存在
	exists, err := s.repo.Exists(ctx, ExistsFilter{Email: &req.Email})
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailExists
	}
	
	// 检查手机号是否已存在
	if req.Phone != nil {
		exists, err := s.repo.Exists(ctx, ExistsFilter{Phone: req.Phone})
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrPhoneExists
		}
	}
	
	// 创建用户
	var user *User
	if req.Password != nil {
		user, err = NewUserWithPassword(req.Name, req.Email, *req.Password)
	} else {
		user, err = NewUser(req.Name, req.Email)
	}
	if err != nil {
		return nil, err
	}
	
	// 设置其他属性
	if req.Phone != nil {
		user.Phone = req.Phone
	}
	if req.Avatar != nil {
		user.Avatar = req.Avatar
	}
	
	// 保存到数据库
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	
	return user, nil
}

// GetUser 获取用户
func (s *ServiceImpl) GetUser(ctx context.Context, id string) (*User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// GetUserByEmail 通过邮箱获取用户
func (s *ServiceImpl) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// UpdateUser 更新用户
func (s *ServiceImpl) UpdateUser(ctx context.Context, id string, req UpdateUserRequest) (*User, error) {
	user, err := s.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	
	// 检查用户状态
	if !user.IsActive() {
		return nil, ErrUserDeactivated
	}
	
	// 更新用户信息
	user.UpdateProfile(req.Name, req.Phone, req.Avatar)
	
	// 保存更新
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}
	
	return user, nil
}

// DeleteUser 删除用户
func (s *ServiceImpl) DeleteUser(ctx context.Context, id string) error {
	user, err := s.GetUser(ctx, id)
	if err != nil {
		return err
	}
	
	// 软删除
	user.SoftDelete()
	
	return s.repo.Update(ctx, user)
}

// ListUsers 列出用户
func (s *ServiceImpl) ListUsers(ctx context.Context, filter ListFilter) (*PaginatedResult, error) {
	users, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	
	// 计算总数
	countFilter := CountFilter{
		Name:           filter.Name,
		Email:          filter.Email,
		IsActive:       filter.IsActive,
		IsAdmin:        filter.IsAdmin,
		IsSystem:       filter.IsSystem,
		CreatedAfter:   filter.CreatedAfter,
		CreatedBefore:  filter.CreatedBefore,
		ModifiedAfter:  filter.ModifiedAfter,
		ModifiedBefore: filter.ModifiedBefore,
		Search:         filter.Search,
	}
	
	total, err := s.repo.Count(ctx, countFilter)
	if err != nil {
		return nil, err
	}
	
	// 计算分页信息
	page := filter.Offset/filter.Limit + 1
	totalPages := int(total)/filter.Limit
	if int(total)%filter.Limit > 0 {
		totalPages++
	}
	
	return &PaginatedResult{
		Users:      users,
		Total:      total,
		Page:       page,
		PageSize:   filter.Limit,
		TotalPages: totalPages,
	}, nil
}

// Authenticate 用户认证
func (s *ServiceImpl) Authenticate(ctx context.Context, email, password string) (*User, error) {
	user, err := s.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	
	// 检查用户状态
	if !user.IsActive() {
		if user.GetStatus() == UserStatusDeactivated {
			return nil, ErrUserDeactivated
		}
		return nil, ErrUserDeleted
	}
	
	// 验证密码
	if err := user.CheckPassword(password); err != nil {
		return nil, ErrInvalidPassword
	}
	
	// 记录登录时间
	user.RecordSignIn()
	if err := s.repo.Update(ctx, user); err != nil {
		// 登录时间更新失败不影响认证结果
		// TODO: 记录日志
	}
	
	return user, nil
}

// ChangePassword 修改密码
func (s *ServiceImpl) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	user, err := s.GetUser(ctx, userID)
	if err != nil {
		return err
	}
	
	// 检查用户状态
	if !user.IsActive() {
		return ErrUserDeactivated
	}
	
	// 验证旧密码
	if err := user.CheckPassword(oldPassword); err != nil {
		return ErrInvalidPassword
	}
	
	// 设置新密码
	if err := user.SetPassword(newPassword); err != nil {
		return err
	}
	
	return s.repo.Update(ctx, user)
}

// ResetPassword 重置密码
func (s *ServiceImpl) ResetPassword(ctx context.Context, email string) error {
	user, err := s.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	
	// 检查用户状态
	if !user.IsActive() {
		return ErrUserDeactivated
	}
	
	// TODO: 实现密码重置逻辑
	// 1. 生成重置令牌
	// 2. 发送重置邮件
	// 3. 存储令牌到缓存
	
	return errors.New("password reset not implemented yet")
}

// ActivateUser 激活用户
func (s *ServiceImpl) ActivateUser(ctx context.Context, userID string) error {
	user, err := s.GetUser(ctx, userID)
	if err != nil {
		return err
	}
	
	user.Activate()
	return s.repo.Update(ctx, user)
}

// DeactivateUser 停用用户
func (s *ServiceImpl) DeactivateUser(ctx context.Context, userID string) error {
	user, err := s.GetUser(ctx, userID)
	if err != nil {
		return err
	}
	
	user.Deactivate()
	return s.repo.Update(ctx, user)
}

// PromoteToAdmin 提升为管理员
func (s *ServiceImpl) PromoteToAdmin(ctx context.Context, userID string) error {
	user, err := s.GetUser(ctx, userID)
	if err != nil {
		return err
	}
	
	user.PromoteToAdmin()
	return s.repo.Update(ctx, user)
}

// DemoteFromAdmin 取消管理员
func (s *ServiceImpl) DemoteFromAdmin(ctx context.Context, userID string) error {
	user, err := s.GetUser(ctx, userID)
	if err != nil {
		return err
	}
	
	user.DemoteFromAdmin()
	return s.repo.Update(ctx, user)
}

// LinkAccount 关联第三方账户
func (s *ServiceImpl) LinkAccount(ctx context.Context, userID string, req LinkAccountRequest) error {
	user, err := s.GetUser(ctx, userID)
	if err != nil {
		return err
	}
	
	// 检查提供商账户是否已被其他用户使用
	existingAccount, err := s.repo.GetAccountByProvider(ctx, string(req.Provider), req.ProviderID)
	if err != nil {
		return err
	}
	if existingAccount != nil {
		return errors.New("provider account already linked to another user")
	}
	
	// 创建账户关联
	account := user.AddAccount(req.Type, req.Provider, req.ProviderID)
	return s.repo.CreateAccount(ctx, account)
}

// UnlinkAccount 取消关联第三方账户
func (s *ServiceImpl) UnlinkAccount(ctx context.Context, userID, accountID string) error {
	// TODO: 验证账户属于该用户
	return s.repo.DeleteAccount(ctx, accountID)
}

// GetUserByProvider 通过第三方提供商获取用户
func (s *ServiceImpl) GetUserByProvider(ctx context.Context, provider, providerID string) (*User, error) {
	account, err := s.repo.GetAccountByProvider(ctx, provider, providerID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, ErrUserNotFound
	}
	
	return s.GetUser(ctx, account.UserID)
}