package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
)

// SetupWeeklyReportRoutes 设置周报路由
func SetupWeeklyReportRoutes(r *gin.RouterGroup, ctrl *controllers.WeeklyReportController) {
	// 周报路由 (需要认证) - 球员端使用
	weeklyReports := r.Group("/weekly-reports")
	weeklyReports.Use(middleware.AuthMiddleware())
	{
		weeklyReports.POST("", ctrl.Create)
		weeklyReports.GET("/pending", ctrl.ListPending)
		weeklyReports.GET("/:id", ctrl.Get)
		weeklyReports.PUT("/:id", ctrl.Update)
		weeklyReports.POST("/:id/review", ctrl.Review)
		weeklyReports.DELETE("/:id", ctrl.Delete)
	}

	// 球员周报路由
	players := r.Group("/players")
	players.Use(middleware.AuthMiddleware())
	{
		players.GET("/:playerId/weekly-reports", ctrl.ListByPlayer)
	}

	// 球队周报路由 /teams/:teamId/weekly-reports 已在 SetupTeamRoutes 中注册
}

// SetupWeeklyPeriodRoutes 设置周报周期路由
func SetupWeeklyPeriodRoutes(r *gin.RouterGroup, ctrl *controllers.WeeklyReportController) {
	// 周报周期管理路由（需要俱乐部/教练权限）
	teams := r.Group("/teams")
	teams.Use(middleware.AuthMiddleware())
	teams.Use(middleware.TeamAccessMiddleware())
	{
		// 周报周期列表
		teams.GET("/:teamId/weekly-periods", ctrl.GetPeriods)
		// 周期统计
		teams.GET("/:teamId/weekly-periods/:periodId/stats", ctrl.GetPeriodStats)
		// 周期内球员提交情况
		teams.GET("/:teamId/weekly-periods/:periodId/players", ctrl.GetPeriodPlayers)
		// 一键提醒未提交球员
		teams.POST("/:teamId/weekly-reports/remind", ctrl.Remind)
		// 导出周报
		teams.GET("/:teamId/weekly-reports/export", ctrl.Export)
	}
}
