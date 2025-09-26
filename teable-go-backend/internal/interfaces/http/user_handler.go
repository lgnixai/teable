package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"teable-go-backend/internal/application"
	"teable-go-backend/internal/domain/user"
	"teable-go-backend/internal/interfaces/middleware"
	"teable-go-backend/pkg/errors"
	"teable-go-backend/pkg/logger"
)

// UserHandler 用户HTTP处理器
type UserHandler struct {
	userService *application.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService *application.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// Register 用户注册
// @Summary 用户注册
// @Description 创建新用户账户
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body application.RegisterRequest true "注册信息"
// @Success 201 {object} application.AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /api/auth/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req application.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid register request", logger.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "请求参数错误",
			Code:  "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}

	response, err := h.userService.Register(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response)
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录认证
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body application.LoginRequest true "登录信息"
// @Success 200 {object} application.AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req application.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid login request", logger.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "请求参数错误",
			Code:  "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}

	response, err := h.userService.Login(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// Logout 用户登出
// @Summary 用户登出
// @Description 用户登出，令牌失效
// @Tags 认证
// @Accept json
// @Produce json
// @Success 200 {object} MessageResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/auth/logout [post]
func (h *UserHandler) Logout(c *gin.Context) {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	token := c.GetString("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "令牌不能为空",
			Code:  "MISSING_TOKEN",
		})
		return
	}

	if err := h.userService.Logout(c.Request.Context(), userID, token); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, MessageResponse{
		Message: "登出成功",
	})
}

// GetProfile 获取当前用户资料
// @Summary 获取用户资料
// @Description 获取当前登录用户的资料信息
// @Tags 用户
// @Accept json
// @Produce json
// @Success 200 {object} application.UserResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/users/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response, err := h.userService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateProfile 更新用户资料
// @Summary 更新用户资料
// @Description 更新当前登录用户的资料信息
// @Tags 用户
// @Accept json
// @Produce json
// @Param request body application.UpdateProfileRequest true "更新资料信息"
// @Success 200 {object} application.UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/users/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	var req application.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid update profile request", logger.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "请求参数错误",
			Code:  "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}

	response, err := h.userService.UpdateProfile(c.Request.Context(), userID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 修改当前登录用户的密码
// @Tags 用户
// @Accept json
// @Produce json
// @Param request body application.ChangePasswordRequest true "修改密码信息"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/users/change-password [post]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	var req application.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid change password request", logger.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "请求参数错误",
			Code:  "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}

	if err := h.userService.ChangePassword(c.Request.Context(), userID, req); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, MessageResponse{
		Message: "密码修改成功",
	})
}

// GetUser 获取用户信息(管理员功能)
// @Summary 获取用户信息
// @Description 根据用户ID获取用户详细信息(需要管理员权限)
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Success 200 {object} application.UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/admin/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "用户ID不能为空",
			Code:  "MISSING_USER_ID",
		})
		return
	}

	response, err := h.userService.GetUser(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// ListUsers 列出用户(管理员功能)
// @Summary 列出用户
// @Description 分页获取用户列表(需要管理员权限)
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param search query string false "搜索关键词"
// @Param is_active query bool false "是否激活"
// @Param is_admin query bool false "是否管理员"
// @Success 200 {object} user.PaginatedResult
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/admin/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	search := c.Query("search")

	// 限制分页参数
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 构建过滤器
	filter := user.NewListFilter()
	filter.Offset = (page - 1) * pageSize
	filter.Limit = pageSize
	filter.Search = search

	// 处理布尔查询参数
	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		if isActive, err := strconv.ParseBool(isActiveStr); err == nil {
			filter.IsActive = &isActive
		}
	}

	if isAdminStr := c.Query("is_admin"); isAdminStr != "" {
		if isAdmin, err := strconv.ParseBool(isAdminStr); err == nil {
			filter.IsAdmin = &isAdmin
		}
	}

	result, err := h.userService.ListUsers(c.Request.Context(), filter)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// 响应结构体
type ErrorResponse struct {
	Error   string      `json:"error"`
	Code    string      `json:"code,omitempty"`
	Details interface{} `json:"details,omitempty"`
	TraceID string      `json:"trace_id,omitempty"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

// handleError 统一错误处理
func (h *UserHandler) handleError(c *gin.Context, err error) {
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
		logger.Error(err),
		logger.String("trace_id", traceID),
	)

	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error:   "服务器内部错误",
		Code:    "INTERNAL_SERVER_ERROR",
		TraceID: traceID,
	})
}