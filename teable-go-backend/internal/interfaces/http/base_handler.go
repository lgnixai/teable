package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"teable-go-backend/internal/domain/base"
	"teable-go-backend/pkg/errors"
	"teable-go-backend/pkg/logger"
)

// BaseHandler 基础表处理器
type BaseHandler struct {
	baseService base.Service
}

// NewBaseHandler 创建基础表处理器
func NewBaseHandler(baseService base.Service) *BaseHandler {
	return &BaseHandler{
		baseService: baseService,
	}
}

// CreateBase 创建基础表
// @Summary 创建基础表
// @Description 在指定空间中创建新的基础表
// @Tags 基础表
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body base.CreateBaseRequest true "创建基础表请求"
// @Success 201 {object} base.Base "创建成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 401 {object} ErrorResponse "未授权"
// @Failure 409 {object} ErrorResponse "基础表已存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/bases [post]
func (h *BaseHandler) CreateBase(c *gin.Context) {
	var req base.CreateBaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "请求参数错误",
			Code:    "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}

	// 从JWT中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "未授权",
			Code:  "UNAUTHORIZED",
		})
		return
	}
	req.CreatedBy = userID.(string)

	b, err := h.baseService.CreateBase(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": b})
}

// GetBase 获取基础表详情
// @Summary 获取基础表详情
// @Description 根据ID获取基础表详情
// @Tags 基础表
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "基础表ID"
// @Success 200 {object} base.Base "获取成功"
// @Failure 404 {object} ErrorResponse "基础表不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/bases/{id} [get]
func (h *BaseHandler) GetBase(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "基础表ID不能为空",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	b, err := h.baseService.GetBase(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": b})
}

// UpdateBase 更新基础表
// @Summary 更新基础表
// @Description 更新基础表信息
// @Tags 基础表
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "基础表ID"
// @Param request body base.UpdateBaseRequest true "更新基础表请求"
// @Success 200 {object} base.Base "更新成功"
// @Failure 400 {object} ErrorResponse "请求参数错误"
// @Failure 404 {object} ErrorResponse "基础表不存在"
// @Failure 409 {object} ErrorResponse "基础表名称已存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/bases/{id} [put]
func (h *BaseHandler) UpdateBase(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "基础表ID不能为空",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	var req base.UpdateBaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "请求参数错误",
			Code:    "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}

	b, err := h.baseService.UpdateBase(c.Request.Context(), id, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": b})
}

// DeleteBase 删除基础表
// @Summary 删除基础表
// @Description 软删除基础表
// @Tags 基础表
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "基础表ID"
// @Success 200 {object} object{success=bool} "删除成功"
// @Failure 404 {object} ErrorResponse "基础表不存在"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/bases/{id} [delete]
func (h *BaseHandler) DeleteBase(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "基础表ID不能为空",
			Code:  "INVALID_REQUEST",
		})
		return
	}

	err := h.baseService.DeleteBase(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ListBases 获取基础表列表
// @Summary 获取基础表列表
// @Description 获取基础表列表，支持分页和过滤
// @Tags 基础表
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param space_id query string false "空间ID"
// @Param name query string false "基础表名称（模糊搜索）"
// @Param search query string false "搜索关键词"
// @Param order_by query string false "排序字段" default(created_time)
// @Param order query string false "排序方向" Enums(asc,desc) default(desc)
// @Param limit query int false "每页数量" default(20)
// @Param offset query int false "偏移量" default(0)
// @Success 200 {object} base.PaginatedResult "获取成功"
// @Failure 500 {object} ErrorResponse "服务器内部错误"
// @Router /api/bases [get]
func (h *BaseHandler) ListBases(c *gin.Context) {
	// 解析查询参数
	filter := base.ListFilter{
		OrderBy: "created_time",
		Order:   "desc",
		Limit:   20,
		Offset:  0,
	}

	if spaceID := c.Query("space_id"); spaceID != "" {
		filter.SpaceID = &spaceID
	}
	if name := c.Query("name"); name != "" {
		filter.Name = &name
	}
	if search := c.Query("search"); search != "" {
		filter.Search = search
	}
	if orderBy := c.Query("order_by"); orderBy != "" {
		filter.OrderBy = orderBy
	}
	if order := c.Query("order"); order != "" {
		filter.Order = order
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	// 获取基础表列表
	bases, err := h.baseService.ListBases(c.Request.Context(), filter)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 获取总数
	countFilter := base.CountFilter{
		SpaceID: filter.SpaceID,
		Name:    filter.Name,
		Search:  filter.Search,
	}
	total, err := h.baseService.CountBases(c.Request.Context(), countFilter)
	if err != nil {
		h.handleError(c, err)
		return
	}

	result := base.PaginatedResult{
		Data:   bases,
		Total:  total,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}

	c.JSON(http.StatusOK, result)
}

// handleError 处理错误
func (h *BaseHandler) handleError(c *gin.Context, err error) {
	traceID := c.GetString("request_id")

	if appErr, ok := errors.IsAppError(err); ok {
		logger.Error("Application error",
			logger.String("error", appErr.Message),
			logger.String("code", appErr.Code),
			logger.String("trace_id", traceID),
		)

		c.JSON(appErr.HTTPStatus, ErrorResponse{
			Error:   appErr.Message,
			Code:    appErr.Code,
			Details: appErr.Details,
			TraceID: traceID,
		})
		return
	}

	logger.Error("Internal server error",
		logger.ErrorField(err),
		logger.String("trace_id", traceID),
	)

	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error:   "服务器内部错误",
		Code:    "INTERNAL_SERVER_ERROR",
		TraceID: traceID,
	})
}
