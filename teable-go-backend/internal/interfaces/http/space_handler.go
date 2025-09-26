package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"teable-go-backend/internal/domain/space"
	"teable-go-backend/internal/interfaces/middleware"
	"teable-go-backend/pkg/errors"
)

// SpaceHandler 空间相关HTTP处理器
type SpaceHandler struct {
	service space.Service
}

func NewSpaceHandler(service space.Service) *SpaceHandler { return &SpaceHandler{service: service} }

// CreateSpace 创建空间
func (h *SpaceHandler) CreateSpace(c *gin.Context) {
	var req struct {
		Name        string  `json:"name" binding:"required"`
		Description *string `json:"description"`
		Icon        *string `json:"icon"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.ErrorResponse{Error: "参数错误", Code: errors.ErrBadRequest.Code, Details: err.Error()})
		return
	}

	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, errors.ErrorResponse{Error: errors.ErrUnauthorized.Message, Code: errors.ErrUnauthorized.Code})
		return
	}

	sp, err := h.service.CreateSpace(c.Request.Context(), space.CreateSpaceRequest{
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
		CreatedBy:   userID,
	})
	if err != nil {
		status := errors.GetHTTPStatus(err)
		c.JSON(status, errors.ErrorResponse{Error: err.Error(), Code: errors.ErrInternalServer.Code})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": sp})
}

// GetSpace 获取空间
func (h *SpaceHandler) GetSpace(c *gin.Context) {
	id := c.Param("id")
	sp, err := h.service.GetSpace(c.Request.Context(), id)
	if err != nil {
		status := errors.GetHTTPStatus(err)
		c.JSON(status, errors.ErrorResponse{Error: err.Error(), Code: errors.ErrNotFound.Code})
		return
	}
	if sp == nil {
		c.JSON(http.StatusNotFound, errors.ErrorResponse{Error: errors.ErrSpaceNotFound.Message, Code: errors.ErrSpaceNotFound.Code})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": sp})
}

// UpdateSpace 更新空间
func (h *SpaceHandler) UpdateSpace(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Icon        *string `json:"icon"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.ErrorResponse{Error: "参数错误", Code: errors.ErrBadRequest.Code, Details: err.Error()})
		return
	}
	sp, err := h.service.UpdateSpace(c.Request.Context(), id, space.UpdateSpaceRequest{
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
	})
	if err != nil {
		status := errors.GetHTTPStatus(err)
		c.JSON(status, errors.ErrorResponse{Error: err.Error(), Code: errors.ErrInternalServer.Code})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": sp})
}

// DeleteSpace 删除空间
func (h *SpaceHandler) DeleteSpace(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteSpace(c.Request.Context(), id); err != nil {
		status := errors.GetHTTPStatus(err)
		c.JSON(status, errors.ErrorResponse{Error: err.Error(), Code: errors.ErrInternalServer.Code})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ListSpaces 列出空间
func (h *SpaceHandler) ListSpaces(c *gin.Context) {
	var query struct {
		Offset    int     `form:"offset"`
		Limit     int     `form:"limit"`
		OrderBy   string  `form:"order_by"`
		Order     string  `form:"order"`
		Name      *string `form:"name"`
		Search    string  `form:"search"`
		CreatedBy *string `form:"created_by"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, errors.ErrorResponse{Error: "查询参数错误", Code: errors.ErrBadRequest.Code, Details: err.Error()})
		return
	}
	filter := space.ListFilter{Offset: query.Offset, Limit: query.Limit, OrderBy: query.OrderBy, Order: query.Order, Name: query.Name, Search: query.Search, CreatedBy: query.CreatedBy}
	items, total, err := h.service.ListSpaces(c.Request.Context(), filter)
	if err != nil {
		status := errors.GetHTTPStatus(err)
		c.JSON(status, errors.ErrorResponse{Error: err.Error(), Code: errors.ErrInternalServer.Code})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "total": total, "offset": filter.Offset, "limit": filter.Limit})
}

