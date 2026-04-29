package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
)

// SetupUserRoutes 设置用户路由
func SetupUserRoutes(r *gin.RouterGroup, userController *controllers.UserController) {
	// 用户资料路由
	user := r.Group("/user")
	user.Use(middleware.AuthMiddleware())
	{
		user.GET("/profile", userController.GetProfile)
		user.PUT("/profile", userController.UpdateProfile)
		user.GET("/growth-records", userController.GetGrowthRecords)
		user.POST("/growth-records", userController.SaveGrowthRecords)
		user.PUT("/growth-records/:id", userController.UpdateGrowthRecord)
		user.DELETE("/growth-records/:id", userController.DeleteGrowthRecord)
		// 账号设置
		user.PUT("/password", userController.ChangePassword)
		user.PUT("/phone", userController.ChangePhone)
		user.GET("/settings", userController.GetSettings)
		user.PUT("/settings", userController.UpdateSettings)
		user.GET("/devices", userController.GetLoginDevices)
		user.DELETE("/devices/:deviceId", userController.LogoutDevice)
	}

	// 公开资料
	public := r.Group("/users")
	{
		public.GET("/:userId/profile", userController.GetPublicProfile)
		public.GET("/:userId/player", userController.GetPlayerProfile)
		public.GET("/:userId/reports", userController.GetPublicReports)
	}
}
