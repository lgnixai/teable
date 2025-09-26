package space

import (
	"context"
	"errors"
)

// Service 空间领域服务接口
type Service interface {
	CreateSpace(ctx context.Context, req CreateSpaceRequest) (*Space, error)
	GetSpace(ctx context.Context, id string) (*Space, error)
	UpdateSpace(ctx context.Context, id string, req UpdateSpaceRequest) (*Space, error)
	DeleteSpace(ctx context.Context, id string) error
	ListSpaces(ctx context.Context, filter ListFilter) ([]*Space, int64, error)

	AddCollaborator(ctx context.Context, spaceID, userID, role string) error
	RemoveCollaborator(ctx context.Context, collabID string) error
	ListCollaborators(ctx context.Context, spaceID string) ([]*SpaceCollaborator, error)
}

type ServiceImpl struct {
	repo Repository
}

func NewService(repo Repository) Service { return &ServiceImpl{repo: repo} }

type CreateSpaceRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=255"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=2000"`
	Icon        *string `json:"icon,omitempty" validate:"omitempty,max=100"`
	CreatedBy   string  `json:"created_by" validate:"required"`
}

type UpdateSpaceRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=2000"`
	Icon        *string `json:"icon,omitempty" validate:"omitempty,max=100"`
}

func (s *ServiceImpl) CreateSpace(ctx context.Context, req CreateSpaceRequest) (*Space, error) {
	space := NewSpace(req.Name, req.CreatedBy)
	space.Description = req.Description
	space.Icon = req.Icon
	if err := s.repo.Create(ctx, space); err != nil { return nil, err }
	return space, nil
}

func (s *ServiceImpl) GetSpace(ctx context.Context, id string) (*Space, error) {
	sp, err := s.repo.GetByID(ctx, id)
	if err != nil { return nil, err }
	if sp == nil { return nil, errors.New("space not found") }
	return sp, nil
}

func (s *ServiceImpl) UpdateSpace(ctx context.Context, id string, req UpdateSpaceRequest) (*Space, error) {
	sp, err := s.GetSpace(ctx, id)
	if err != nil { return nil, err }
	sp.Update(req.Name, req.Description, req.Icon)
	if err := s.repo.Update(ctx, sp); err != nil { return nil, err }
	return sp, nil
}

func (s *ServiceImpl) DeleteSpace(ctx context.Context, id string) error {
	sp, err := s.GetSpace(ctx, id)
	if err != nil { return err }
	sp.SoftDelete()
	return s.repo.Update(ctx, sp)
}

func (s *ServiceImpl) ListSpaces(ctx context.Context, filter ListFilter) ([]*Space, int64, error) {
	items, err := s.repo.List(ctx, filter)
	if err != nil { return nil, 0, err }
	total, err := s.repo.Count(ctx, CountFilter{Name: filter.Name, CreatedBy: filter.CreatedBy, Search: filter.Search})
	if err != nil { return nil, 0, err }
	return items, total, nil
}

func (s *ServiceImpl) AddCollaborator(ctx context.Context, spaceID, userID, role string) error {
	return s.repo.AddCollaborator(ctx, NewSpaceCollaborator(spaceID, userID, role))
}

func (s *ServiceImpl) RemoveCollaborator(ctx context.Context, collabID string) error {
	return s.repo.RemoveCollaborator(ctx, collabID)
}

func (s *ServiceImpl) ListCollaborators(ctx context.Context, spaceID string) ([]*SpaceCollaborator, error) {
	return s.repo.ListCollaborators(ctx, spaceID)
}

