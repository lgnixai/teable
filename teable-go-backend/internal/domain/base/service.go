package base

import (
	"context"
	"errors"
)

// 业务错误定义
var (
	ErrBaseNotFound     = errors.New("基础表不存在")
	ErrBaseExists       = errors.New("基础表已存在")
	ErrInvalidSpaceID   = errors.New("无效的空间ID")
	ErrInvalidName      = errors.New("无效的基础表名称")
	ErrInvalidCreatedBy = errors.New("无效的创建者ID")
	ErrPermissionDenied = errors.New("权限不足")
)

// Service 基础表服务接口
type Service interface {
	// 创建基础表
	CreateBase(ctx context.Context, req CreateBaseRequest) (*Base, error)

	// 获取基础表
	GetBase(ctx context.Context, id string) (*Base, error)

	// 更新基础表
	UpdateBase(ctx context.Context, id string, req UpdateBaseRequest) (*Base, error)

	// 删除基础表
	DeleteBase(ctx context.Context, id string) error

	// 列出基础表
	ListBases(ctx context.Context, filter ListFilter) ([]*Base, error)

	// 统计基础表数量
	CountBases(ctx context.Context, filter CountFilter) (int64, error)
}

// CreateBaseRequest 创建基础表请求
type CreateBaseRequest struct {
	SpaceID     string  `json:"space_id" binding:"required"`
	Name        string  `json:"name" binding:"required,min=1,max=255"`
	Description *string `json:"description"`
	Icon        *string `json:"icon"`
	CreatedBy   string  `json:"-"` // 从JWT中获取
}

// UpdateBaseRequest 更新基础表请求
type UpdateBaseRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=255"`
	Description *string `json:"description"`
	Icon        *string `json:"icon"`
}

// PaginatedResult 分页结果
type PaginatedResult struct {
	Data   []*Base `json:"data"`
	Total  int64   `json:"total"`
	Limit  int     `json:"limit"`
	Offset int     `json:"offset"`
}

// ServiceImpl 基础表服务实现
type ServiceImpl struct {
	repo Repository
}

// NewService 创建基础表服务
func NewService(repo Repository) Service {
	return &ServiceImpl{
		repo: repo,
	}
}

// CreateBase 创建基础表
func (s *ServiceImpl) CreateBase(ctx context.Context, req CreateBaseRequest) (*Base, error) {
	// 检查同一空间下名称是否已存在
	exists, err := s.repo.Exists(ctx, ExistsFilter{
		SpaceID: &req.SpaceID,
		Name:    &req.Name,
	})
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrBaseExists
	}

	// 创建基础表
	base := NewBase(req.SpaceID, req.Name, req.CreatedBy)
	base.Description = req.Description
	base.Icon = req.Icon

	if err := base.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, base); err != nil {
		return nil, err
	}

	return base, nil
}

// GetBase 获取基础表
func (s *ServiceImpl) GetBase(ctx context.Context, id string) (*Base, error) {
	base, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if base == nil {
		return nil, ErrBaseNotFound
	}
	return base, nil
}

// UpdateBase 更新基础表
func (s *ServiceImpl) UpdateBase(ctx context.Context, id string, req UpdateBaseRequest) (*Base, error) {
	base, err := s.GetBase(ctx, id)
	if err != nil {
		return nil, err
	}

	// 如果名称改变，检查新名称是否已存在
	if base.Name != req.Name {
		exists, err := s.repo.Exists(ctx, ExistsFilter{
			SpaceID: &base.SpaceID,
			Name:    &req.Name,
		})
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrBaseExists
		}
	}

	base.Update(req.Name, req.Description, req.Icon)

	if err := s.repo.Update(ctx, base); err != nil {
		return nil, err
	}

	return base, nil
}

// DeleteBase 删除基础表
func (s *ServiceImpl) DeleteBase(ctx context.Context, id string) error {
	base, err := s.GetBase(ctx, id)
	if err != nil {
		return err
	}

	base.SoftDelete()
	return s.repo.Update(ctx, base)
}

// ListBases 列出基础表
func (s *ServiceImpl) ListBases(ctx context.Context, filter ListFilter) ([]*Base, error) {
	return s.repo.List(ctx, filter)
}

// CountBases 统计基础表数量
func (s *ServiceImpl) CountBases(ctx context.Context, filter CountFilter) (int64, error) {
	return s.repo.Count(ctx, filter)
}
