package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
	"github.com/shaonianqiutan/backend/repositories"
	"gorm.io/gorm"
)

// SetupCoachRoutes 设置教练路由
func SetupCoachRoutes(r *gin.RouterGroup, coachController *controllers.CoachController, teamRepo *repositories.TeamRepository, footballExpController *controllers.FootballExperienceController, weeklyReportController *controllers.WeeklyReportController, db *gorm.DB) {
	// 公开路由（无需认证）
	coaches := r.Group("/coaches")
	{
		coaches.GET("/public", coachController.GetCoachPublicProfileByUser) // 通过 user_id 获取教练公开主页
		coaches.GET("/:id/public", coachController.GetCoachPublicProfile)  // 通过 coach_id 获取教练公开主页
	}

	// 教练路由（需要认证）
	coach := r.Group("/coach")
	coach.Use(middleware.AuthMiddleware())
	{
		// 教练资料
		coach.GET("/profile", coachController.GetCoachProfile)
		coach.PUT("/profile", coachController.UpdateCoachProfile)

		// 足球经历
		coach.GET("/football-experiences", footballExpController.GetFootballExperiences)
		coach.POST("/football-experiences", footballExpController.CreateFootballExperience)
		coach.PUT("/football-experiences/:id", footballExpController.UpdateFootballExperience)
		coach.DELETE("/football-experiences/:id", footballExpController.DeleteFootballExperience)

		// 工作台
		coach.GET("/dashboard", coachController.GetDashboard)

		// 周报详情（教练查看单个周报）
		coach.GET("/weekly-reports/:id", weeklyReportController.Get)
		// 教练审核周报
		coach.POST("/weekly-reports/:id/review", weeklyReportController.Review)

		// 关注的球员
		coach.GET("/followed-players", coachController.GetFollowedPlayers)
		coach.POST("/followed-players", coachController.FollowPlayer)
		coach.DELETE("/followed-players/:playerId", coachController.UnfollowPlayer)
		coach.PUT("/followed-players/:playerId/notes", coachController.UpdateFollowNotes)

		// 训练笔记
		coach.GET("/training-notes", coachController.GetTrainingNotes)
		coach.POST("/training-notes", coachController.CreateTrainingNote)
		coach.PUT("/training-notes/:id", coachController.UpdateTrainingNote)
		coach.DELETE("/training-notes/:id", coachController.DeleteTrainingNote)

		// 球员进度
		coach.GET("/players/:playerId/progress", coachController.GetPlayerProgress)

		// 获取我的球队列表（球队级统一路由已迁移到 /api/teams/:teamId/*）
		coachTeamController := controllers.NewCoachTeamController(teamRepo, nil, db)
		coach.GET("/teams", coachTeamController.GetMyTeams)
	}
}
