package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
)

// SetupClubRoutes 设置俱乐部路由
func SetupClubRoutes(
	r *gin.RouterGroup,
	clubController *controllers.ClubController,
	trainingPlanController *controllers.TrainingPlanController,
	matchScheduleController *controllers.MatchScheduleController,
) {
	// 公开路由：俱乐部列表
	r.GET("/clubs", clubController.ListPublicClubs)
	r.GET("/clubs/search", clubController.SearchClubs)
	r.GET("/clubs/:clubId", clubController.GetClubDetail)

	// 俱乐部路由（需要认证）
	club := r.Group("/club")
	club.Use(middleware.AuthMiddleware())
	{
		// 俱乐部资料
		club.GET("/profile", clubController.GetClubProfile)
		club.PUT("/profile", clubController.UpdateClubProfile)

		// 工作台
		club.GET("/dashboard", clubController.GetDashboard)

		// 数据分析
		club.GET("/analytics", clubController.GetAnalytics)

		// 球员管理
		club.GET("/players", clubController.GetPlayers)
		club.GET("/players/selection", clubController.GetPlayerSelection)
		club.GET("/players/:id", clubController.GetPlayerDetail)
		club.POST("/players/invite", clubController.InvitePlayer)
		club.POST("/players/import", clubController.ImportPlayers)
		club.PUT("/players/:id/tags", clubController.UpdatePlayerTags)
		club.DELETE("/players/:id", clubController.RemovePlayer)

		// 比赛汇总统计
		club.GET("/match-summaries/summary", clubController.GetMatchSummaryStats)

		// 报告导出
		club.GET("/reports/:reportId/export", clubController.ExportReport)

		// 筛选方案
		club.GET("/player-filter-presets", clubController.GetPlayerFilterPresets)
		club.POST("/player-filter-presets", clubController.CreatePlayerFilterPreset)
		club.PUT("/player-filter-presets/:id", clubController.UpdatePlayerFilterPreset)
		club.DELETE("/player-filter-presets/:id", clubController.DeletePlayerFilterPreset)

		// 俱乐部动态通知
		club.GET("/notifications", clubController.GetClubNotifications)

		// 公告
		club.GET("/announcements", clubController.GetAnnouncements)
		club.POST("/announcements", clubController.CreateAnnouncement)
		club.PUT("/announcements/:id", clubController.UpdateAnnouncement)
		club.DELETE("/announcements/:id", clubController.DeleteAnnouncement)

		// 管理员操作日志
		club.GET("/admin-logs", clubController.GetAdminOperationLogs)

		// 候选名单
		club.GET("/shortlist", clubController.GetShortlist)
		club.POST("/shortlist", clubController.AddToShortlist)
		club.PUT("/shortlist/:playerId", clubController.UpdateShortlistNote)
		club.DELETE("/shortlist/:playerId", clubController.RemoveFromShortlist)

		// 教练管理
		club.GET("/coaches", clubController.GetClubCoaches)
		club.POST("/coaches", clubController.AddClubCoach)
		club.GET("/coaches/:coachId", clubController.GetClubCoachDetail)
		club.PUT("/coaches/:coachId", clubController.UpdateClubCoach)
		club.DELETE("/coaches/:coachId", clubController.RemoveClubCoach)
		club.POST("/coaches/:coachId/assign", clubController.AssignCoachToTeam)
		club.DELETE("/coaches/team/:teamCoachId", clubController.RemoveCoachFromTeam)

		// 俱乐部教练邀请（管理员视角）
		club.POST("/invitations", clubController.CreateClubInvitation)
		club.GET("/invitations", clubController.GetClubInvitations)

		// 球队赛季档案
		club.GET("/teams/:teamId/season-archives", clubController.GetTeamSeasonArchives)
		club.POST("/teams/:teamId/season-archives", clubController.CreateTeamSeasonArchive)
		club.GET("/teams/:teamId/season-archives/:id", clubController.GetTeamSeasonArchiveDetail)
		club.PUT("/teams/:teamId/season-archives/:id", clubController.UpdateTeamSeasonArchive)
		club.DELETE("/teams/:teamId/season-archives/:id", clubController.DeleteTeamSeasonArchive)

		// 训练计划
		club.GET("/training-plans", trainingPlanController.ListTrainingPlans)
		club.POST("/training-plans", trainingPlanController.CreateTrainingPlan)
		club.GET("/training-plans/:id", trainingPlanController.GetTrainingPlan)
		club.PUT("/training-plans/:id", trainingPlanController.UpdateTrainingPlan)
		club.DELETE("/training-plans/:id", trainingPlanController.DeleteTrainingPlan)

		// 赛程日历
		club.GET("/match-schedules", matchScheduleController.ListMatchSchedules)
		club.POST("/match-schedules", matchScheduleController.CreateMatchSchedule)
		club.GET("/match-schedules/:id", matchScheduleController.GetMatchSchedule)
		club.PUT("/match-schedules/:id", matchScheduleController.UpdateMatchSchedule)
		club.DELETE("/match-schedules/:id", matchScheduleController.DeleteMatchSchedule)
	}

	// 俱乐部邀请（教练视角 - 需要认证但不是俱乐部成员）
	clubInv := r.Group("/club-invitations")
	clubInv.Use(middleware.AuthMiddleware())
	{
		clubInv.GET("/my", clubController.GetMyClubInvitations)
		clubInv.POST("/:code/accept", clubController.AcceptClubInvitation)
		clubInv.POST("/:code/reject", clubController.RejectClubInvitation)
	}
}
