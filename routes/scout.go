package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
)

// SetupScoutRoutes 设置球探路由
func SetupScoutRoutes(r *gin.RouterGroup, scoutController *controllers.ScoutController) {
	// 公开路由（无需认证）
	scouts := r.Group("/scouts")
	{
		scouts.GET("/public", scoutController.GetScoutPublicProfile)     // 通过 user_id 获取球探公开主页
		scouts.GET("/:id/public", scoutController.GetScoutPublicProfileByID) // 通过 scout_id 获取球探公开主页
	}

	// 球探路由（需要认证）
	scout := r.Group("/scout")
	scout.Use(middleware.AuthMiddleware())
	{
		// 球探资料
		scout.GET("/profile", scoutController.GetScoutProfile)
		scout.PUT("/profile", scoutController.UpdateScoutProfile)

		// 工作台
		scout.GET("/dashboard", scoutController.GetScoutDashboard)

		// 关注的球员
		scout.GET("/followed-players", scoutController.GetFollowedPlayers)
		scout.POST("/followed-players", scoutController.FollowPlayer)
		scout.DELETE("/followed-players/:playerId", scoutController.UnfollowPlayer)

		// 球探报告
		scout.GET("/reports", scoutController.GetScoutReports)
		scout.POST("/reports", scoutController.CreateScoutReport)
		scout.GET("/reports/:reportId", scoutController.GetScoutReport)
		scout.PUT("/reports/:reportId", scoutController.UpdateScoutReport)
		scout.DELETE("/reports/:reportId", scoutController.DeleteScoutReport)
		scout.POST("/reports/:reportId/publish", scoutController.PublishScoutReport)

		// 球探任务
		scout.GET("/tasks", scoutController.GetScoutTasks)
		scout.POST("/tasks/:taskId/accept", scoutController.AcceptScoutTask)

		// 球员搜索
		scout.GET("/players/search", scoutController.SearchPlayers)
	}
}
