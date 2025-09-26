package application

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"teable-go-backend/internal/config"
	"teable-go-backend/internal/domain/user"
	"teable-go-backend/internal/infrastructure/cache"
	"teable-go-backend/pkg/errors"
	"teable-go-backend/pkg/logger"
)

// UserService 用户应用服务
type UserService struct {
	userDomainService user.Service
	cacheService      cache.CacheService
	jwtConfig         config.JWTConfig
}

// NewUserService 创建用户应用服务
func NewUserService(
	userDomainService user.Service,
	cacheService cache.CacheService,
	jwtConfig config.JWTConfig,
) *UserService {
	return &UserService{
		userDomainService: userDomainService,
		cacheService:      cacheService,
		jwtConfig:         jwtConfig,
	}
}

// AuthResponse 认证响应
type AuthResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token,omitempty"`
	ExpiresIn    int64         `json:"expires_in"`
	TokenType    string        `json:"token_type"`
}

// UserResponse 用户响应
type UserResponse struct {
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	Email            string     `json:"email"`
	Phone            *string    `json:"phone"`
	Avatar           *string    `json:"avatar"`
	IsAdmin          bool       `json:"is_admin"`
	IsSystem         bool       `json:"is_system"`
	IsTrialUsed      bool       `json:"is_trial_used"`
	LastSignTime     *time.Time `json:"last_sign_time"`
	CreatedTime      time.Time  `json:"created_time"`
	LastModifiedTime *time.Time `json:"last_modified_time"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Name     string  `json:"name" validate:"required,min=1,max=100"`
	Email    string  `json:"email" validate:"required,email,max=255"`
	Password string  `json:"password" validate:"required,min=8,max=128"`
	Phone    *string `json:"phone,omitempty" validate:"omitempty,max=50"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=128"`
}

// UpdateProfileRequest 更新资料请求
type UpdateProfileRequest struct {
	Name   *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Phone  *string `json:"phone,omitempty" validate:"omitempty,max=50"`
	Avatar *string `json:"avatar,omitempty" validate:"omitempty,url,max=500"`
}

// Register 用户注册
func (s *UserService) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	// 创建用户
	createReq := user.CreateUserRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: &req.Password,
		Phone:    req.Phone,
	}
	
	domainUser, err := s.userDomainService.CreateUser(ctx, createReq)
	if err != nil {
		logger.Error("Failed to create user", 
			logger.String("email", req.Email),
			logger.Error(err),
		)
		return nil, err
	}
	
	// 生成令牌
	tokens, err := s.generateTokens(domainUser)
	if err != nil {
		logger.Error("Failed to generate tokens",
			logger.String("user_id", domainUser.ID),
			logger.Error(err),
		)
		return nil, err
	}
	
	// 缓存用户信息
	if err := s.cacheUserInfo(ctx, domainUser); err != nil {
		logger.Warn("Failed to cache user info",
			logger.String("user_id", domainUser.ID),
			logger.Error(err),
		)
	}
	
	logger.Info("User registered successfully",
		logger.String("user_id", domainUser.ID),
		logger.String("email", domainUser.Email),
	)
	
	return &AuthResponse{
		User:         s.toUserResponse(domainUser),
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    int64(s.jwtConfig.AccessTokenTTL.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// Login 用户登录
func (s *UserService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	// 认证用户
	domainUser, err := s.userDomainService.Authenticate(ctx, req.Email, req.Password)
	if err != nil {
		logger.Warn("Login failed",
			logger.String("email", req.Email),
			logger.Error(err),
		)
		return nil, err
	}
	
	// 生成令牌
	tokens, err := s.generateTokens(domainUser)
	if err != nil {
		logger.Error("Failed to generate tokens",
			logger.String("user_id", domainUser.ID),
			logger.Error(err),
		)
		return nil, err
	}
	
	// 缓存用户信息
	if err := s.cacheUserInfo(ctx, domainUser); err != nil {
		logger.Warn("Failed to cache user info",
			logger.String("user_id", domainUser.ID),
			logger.Error(err),
		)
	}
	
	logger.Info("User logged in successfully",
		logger.String("user_id", domainUser.ID),
		logger.String("email", domainUser.Email),
	)
	
	return &AuthResponse{
		User:         s.toUserResponse(domainUser),
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    int64(s.jwtConfig.AccessTokenTTL.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// Logout 用户登出
func (s *UserService) Logout(ctx context.Context, userID, token string) error {
	// 将令牌加入黑名单
	if err := s.blacklistToken(ctx, token); err != nil {
		logger.Error("Failed to blacklist token",
			logger.String("user_id", userID),
			logger.Error(err),
		)
		return err
	}
	
	// 清除用户缓存
	if err := s.clearUserCache(ctx, userID); err != nil {
		logger.Warn("Failed to clear user cache",
			logger.String("user_id", userID),
			logger.Error(err),
		)
	}
	
	logger.Info("User logged out successfully",
		logger.String("user_id", userID),
	)
	
	return nil
}

// GetProfile 获取用户资料
func (s *UserService) GetProfile(ctx context.Context, userID string) (*UserResponse, error) {
	domainUser, err := s.userDomainService.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	return s.toUserResponse(domainUser), nil
}

// UpdateProfile 更新用户资料
func (s *UserService) UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) (*UserResponse, error) {
	updateReq := user.UpdateUserRequest{
		Name:   req.Name,
		Phone:  req.Phone,
		Avatar: req.Avatar,
	}
	
	domainUser, err := s.userDomainService.UpdateUser(ctx, userID, updateReq)
	if err != nil {
		logger.Error("Failed to update user profile",
			logger.String("user_id", userID),
			logger.Error(err),
		)
		return nil, err
	}
	
	// 清除用户缓存
	if err := s.clearUserCache(ctx, userID); err != nil {
		logger.Warn("Failed to clear user cache after update",
			logger.String("user_id", userID),
			logger.Error(err),
		)
	}
	
	logger.Info("User profile updated successfully",
		logger.String("user_id", userID),
	)
	
	return s.toUserResponse(domainUser), nil
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(ctx context.Context, userID string, req ChangePasswordRequest) error {
	err := s.userDomainService.ChangePassword(ctx, userID, req.OldPassword, req.NewPassword)
	if err != nil {
		logger.Error("Failed to change password",
			logger.String("user_id", userID),
			logger.Error(err),
		)
		return err
	}
	
	logger.Info("Password changed successfully",
		logger.String("user_id", userID),
	)
	
	return nil
}

// GetUser 获取用户信息(管理员功能)
func (s *UserService) GetUser(ctx context.Context, userID string) (*UserResponse, error) {
	domainUser, err := s.userDomainService.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	return s.toUserResponse(domainUser), nil
}

// ListUsers 列出用户(管理员功能)
func (s *UserService) ListUsers(ctx context.Context, filter user.ListFilter) (*user.PaginatedResult, error) {
	result, err := s.userDomainService.ListUsers(ctx, filter)
	if err != nil {
		return nil, err
	}
	
	return result, nil
}

// 私有方法

// TokenPair 令牌对
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// generateTokens 生成访问令牌和刷新令牌
func (s *UserService) generateTokens(user *user.User) (*TokenPair, error) {
	// 创建临时用户模型用于令牌生成
	userModel := &struct {
		ID       string
		Email    string
		Name     string
		IsAdmin  bool
		IsSystem bool
	}{
		ID:       user.ID,
		Email:    user.Email,
		Name:     user.Name,
		IsAdmin:  user.IsAdmin,
		IsSystem: user.IsSystem,
	}
	
	// 生成访问令牌
	accessToken, err := s.generateJWTToken(userModel, "access")
	if err != nil {
		return nil, err
	}
	
	var refreshToken string
	if s.jwtConfig.EnableRefresh {
		refreshToken, err = s.generateJWTToken(userModel, "refresh")
		if err != nil {
			return nil, err
		}
	}
	
	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// generateJWTToken 生成JWT令牌
func (s *UserService) generateJWTToken(user interface{}, tokenType string) (string, error) {
	now := time.Now()
	var exp time.Time
	
	switch tokenType {
	case "access":
		exp = now.Add(s.jwtConfig.AccessTokenTTL)
	case "refresh":
		exp = now.Add(s.jwtConfig.RefreshTokenTTL)
	default:
		exp = now.Add(s.jwtConfig.AccessTokenTTL)
	}
	
	// 根据用户类型提取信息
	var userID, email, name string
	var isAdmin, isSystem bool
	
	switch u := user.(type) {
	case *user.User:
		userID = u.ID
		email = u.Email
		name = u.Name
		isAdmin = u.IsAdmin
		isSystem = u.IsSystem
	default:
		// 处理匿名结构体
		if v, ok := user.(struct {
			ID       string
			Email    string
			Name     string
			IsAdmin  bool
			IsSystem bool
		}); ok {
			userID = v.ID
			email = v.Email
			name = v.Name
			isAdmin = v.IsAdmin
			isSystem = v.IsSystem
		} else {
			return "", errors.ErrInternalServer
		}
	}

	claims := jwt.MapClaims{
		"user_id":    userID,
		"email":      email,
		"name":       name,
		"is_admin":   isAdmin,
		"is_system":  isSystem,
		"token_type": tokenType,
		"iss":        s.jwtConfig.Issuer,
		"iat":        now.Unix(),
		"exp":        exp.Unix(),
		"nbf":        now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtConfig.Secret))
}

// cacheUserInfo 缓存用户信息
func (s *UserService) cacheUserInfo(ctx context.Context, user *user.User) error {
	key := cache.BuildCacheKey(cache.UserCachePrefix, user.ID)
	return s.cacheService.Set(ctx, key, user, 24*time.Hour)
}

// clearUserCache 清除用户缓存
func (s *UserService) clearUserCache(ctx context.Context, userID string) error {
	key := cache.BuildCacheKey(cache.UserCachePrefix, userID)
	return s.cacheService.Delete(ctx, key)
}

// blacklistToken 将令牌加入黑名单
func (s *UserService) blacklistToken(ctx context.Context, token string) error {
	key := cache.BuildCacheKey("blacklist:", token)
	// 设置过期时间为令牌的有效期
	return s.cacheService.Set(ctx, key, true, s.jwtConfig.AccessTokenTTL)
}

// toUserResponse 转换为用户响应
func (s *UserService) toUserResponse(user *user.User) *UserResponse {
	return &UserResponse{
		ID:               user.ID,
		Name:             user.Name,
		Email:            user.Email,
		Phone:            user.Phone,
		Avatar:           user.Avatar,
		IsAdmin:          user.IsAdmin,
		IsSystem:         user.IsSystem,
		IsTrialUsed:      user.IsTrialUsed,
		LastSignTime:     user.LastSignTime,
		CreatedTime:      user.CreatedTime,
		LastModifiedTime: user.LastModifiedTime,
	}
}