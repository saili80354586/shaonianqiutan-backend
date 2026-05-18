package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
)

// SetupAnalystApplicationRoutes 设置分析师申请路由
func SetupAnalystApplicationRoutes(
	r *gin.RouterGroup,
	appController *controllers.AnalystApplicationController,
) {
	// 用户端路由
	application := r.Group("/analyst-application")
	application.Use(middleware.AuthMiddleware())
	{
		application.POST("/", appController.CreateApplication)
		application.GET("/my", appController.GetMyApplication)
	}

	// 管理后台路由
	admin := r.Group("/admin/applications")
	admin.Use(middleware.AuthMiddleware())
	admin.Use(middleware.AdminRoleMiddleware())
	{
		admin.GET("", middleware.AdminPermissionMiddleware("applications.review"), appController.GetApplicationList)
		admin.POST("/:id/review", middleware.AdminPermissionMiddleware("applications.review"), appController.ReviewApplication)
	}
}
