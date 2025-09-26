package http

import (
	"github.com/gin-gonic/gin"

	"teable-go-backend/internal/application"
	"teable-go-backend/internal/domain/space"
	"teable-go-backend/internal/interfaces/middleware"
)

// RouterConfig 路由配置
type RouterConfig struct {
	UserService *application.UserService
	AuthService middleware.AuthService
    SpaceService space.Service
}

// SetupRoutes 设置路由
func SetupRoutes(router *gin.Engine, config RouterConfig) {
	// 创建处理器
	userHandler := NewUserHandler(config.UserService)
    spaceHandler := NewSpaceHandler(config.SpaceService)

	// API v1 路由组
	v1 := router.Group("/api")
	{
		// 认证相关路由 (无需认证)
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", userHandler.Register)
			authGroup.POST("/login", userHandler.Login)
			authGroup.POST("/logout", middleware.AuthMiddleware(config.AuthService), userHandler.Logout)
		}

        // 用户相关路由 (需要认证)
		userGroup := v1.Group("/users")
		userGroup.Use(middleware.AuthMiddleware(config.AuthService))
		{
			userGroup.GET("/profile", userHandler.GetProfile)
			userGroup.PUT("/profile", userHandler.UpdateProfile)
			userGroup.POST("/change-password", userHandler.ChangePassword)
		}

        // 空间相关路由 (需要认证)
        spaceGroup := v1.Group("/spaces")
        spaceGroup.Use(middleware.AuthMiddleware(config.AuthService))
        {
            spaceGroup.POST("", spaceHandler.CreateSpace)
            spaceGroup.GET("", spaceHandler.ListSpaces)
            spaceGroup.GET(":id", spaceHandler.GetSpace)
            spaceGroup.PUT(":id", spaceHandler.UpdateSpace)
            spaceGroup.DELETE(":id", spaceHandler.DeleteSpace)
        }

		// 管理员相关路由 (需要管理员权限)
		adminGroup := v1.Group("/admin")
		adminGroup.Use(middleware.AuthMiddleware(config.AuthService))
		adminGroup.Use(middleware.AdminRequiredMiddleware())
		{
			adminGroup.GET("/users", userHandler.ListUsers)
			adminGroup.GET("/users/:id", userHandler.GetUser)
			// TODO: 添加更多管理员功能
			// adminGroup.PUT("/users/:id", userHandler.UpdateUser)
			// adminGroup.DELETE("/users/:id", userHandler.DeleteUser)
			// adminGroup.POST("/users/:id/activate", userHandler.ActivateUser)
			// adminGroup.POST("/users/:id/deactivate", userHandler.DeactivateUser)
		}

		// 健康检查和信息路由 (无需认证)
		v1.GET("/health", HealthCheckHandler)
		v1.GET("/info", InfoHandler)
	}
}

// HealthCheckHandler 健康检查处理器
func HealthCheckHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "healthy",
		"service": "teable-go-backend",
		"version": "1.0.0",
	})
}

// InfoHandler 服务信息处理器
func InfoHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"service": "teable-go-backend",
		"version": "1.0.0",
		"description": "Teable后端服务 - Go语言重构版本",
		"author": "Teable Team",
		"license": "AGPL-3.0",
	})
}