package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
	"github.com/shaonianqiutan/backend/repositories"
	"gorm.io/gorm"
)

// SetupTeamRoutes 设置球队相关路由（统一端点）
func SetupTeamRoutes(
	r *gin.RouterGroup,
	teamRepo *repositories.TeamRepository,
	weeklyReportRepo *repositories.WeeklyReportRepository,
	matchSummaryRepo *repositories.MatchSummaryRepository,
	teamHomeController *controllers.CoachTeamHomeController,
	ptController *controllers.PhysicalTestController,
	db *gorm.DB,
) {
	teamCtrl := controllers.NewTeamController(
		teamRepo,
		weeklyReportRepo,
		matchSummaryRepo,
		nil, nil, nil, db,
	)

	// 俱乐部下的球队路由（需认证 - 俱乐部管理员）
	clubs := r.Group("/clubs")
	clubs.Use(middleware.AuthMiddleware())
	{
		clubs.GET("/:clubId/teams", teamCtrl.GetTeams)
		clubs.POST("/:clubId/teams", teamCtrl.CreateTeam)
	}

	// 统一球队管理路由（需认证 + 球队访问权限）
	// 同时支持俱乐部管理员和球队教练访问
	teams := r.Group("/teams")
	teams.Use(middleware.AuthMiddleware())
	teams.Use(middleware.TeamAccessMiddleware())
	{
		// 球队基本信息
		teams.GET("/:teamId", teamCtrl.GetTeam)
		teams.PUT("/:teamId", teamCtrl.UpdateTeam)
		teams.DELETE("/:teamId", middleware.RequireClubAdmin(), teamCtrl.DeleteTeam)
		teams.PUT("/:teamId/restore", middleware.RequireClubAdmin(), teamCtrl.RestoreTeam)

		// 球员管理
		teams.GET("/:teamId/players", teamCtrl.GetTeamPlayers)
		teams.POST("/:teamId/players", teamCtrl.AddPlayer)
		teams.PUT("/:teamId/players/:playerId", teamCtrl.UpdatePlayer)
		teams.DELETE("/:teamId/players/:playerId", teamCtrl.RemovePlayer)

		// 教练管理（仅俱乐部管理员可添加/移除教练）
		teams.GET("/:teamId/coaches", teamCtrl.GetTeamCoaches)
		teams.POST("/:teamId/coaches", middleware.RequireClubAdmin(), teamCtrl.AddCoach)
		teams.PUT("/:teamId/coaches/:coachId", middleware.RequireClubAdmin(), teamCtrl.UpdateCoach)
		teams.DELETE("/:teamId/coaches/:coachId", middleware.RequireClubAdmin(), teamCtrl.RemoveCoach)

		// 邀请管理
		teams.POST("/:teamId/invitations", teamCtrl.CreateInvitation)
		teams.GET("/:teamId/invitations", teamCtrl.GetInvitations)

		// 入队申请管理（审核等操作需要球队权限；提交申请在公开路由中）
		teams.GET("/:teamId/applications", teamCtrl.GetTeamApplications)
		teams.PUT("/:teamId/applications/:id", teamCtrl.ReviewApplication)

		// 周报管理（需教练或俱乐部管理员权限）
		teams.POST("/:teamId/weekly-reports", middleware.RequireCoach(), teamCtrl.CreateWeeklyReport)
		teams.GET("/:teamId/weekly-reports", teamCtrl.GetWeeklyReports)

		// 比赛总结管理
		teams.POST("/:teamId/match-summaries", middleware.RequireCoach(), teamCtrl.CreateMatchSummary)
		teams.GET("/:teamId/match-summaries", teamCtrl.GetMatchSummaries)

		// 体测管理（统一端点，按俱乐部维度管理）
		teams.GET("/:teamId/physical-tests", ptController.GetPhysicalTests)
		teams.POST("/:teamId/physical-tests", ptController.CreatePhysicalTest)
		teams.GET("/:teamId/physical-tests/:id", ptController.GetPhysicalTest)
		teams.PUT("/:teamId/physical-tests/:id", ptController.UpdatePhysicalTest)
		teams.DELETE("/:teamId/physical-tests/:id", ptController.DeletePhysicalTest)
		teams.GET("/:teamId/physical-tests/:id/records", ptController.GetPhysicalTestRecords)
		teams.POST("/:teamId/physical-tests/:id/records", ptController.CreatePhysicalTestRecord)
	}
}

// SetupPublicTeamRoutes 设置公开的球队路由（无需认证或仅需认证）
func SetupPublicTeamRoutes(r *gin.RouterGroup, teamRepo *repositories.TeamRepository, db *gorm.DB) {
	// 需要创建简化版的控制器，只传入 teamRepo 和 db
	teamCtrl := controllers.NewTeamController(teamRepo, nil, nil, nil, nil, nil, db)

	// 需要认证但不需要球队成员权限的路由（球员申请加入球队）
	// 单独注册，避免被 TeamAccessMiddleware 拦截
	r.POST("/teams/:teamId/applications", middleware.AuthMiddleware(), teamCtrl.CreateApplication)

	// 我的邀请（需认证）— 必须放在 /invitations/:code 之前注册
	invitations := r.Group("/invitations")
	invitations.Use(middleware.AuthMiddleware())
	{
		invitations.GET("/my", teamCtrl.GetMyInvitations)
	}

	// 邀请相关（公开 + 认证）
	r.GET("/invitations/:code", teamCtrl.GetInvitation)
	r.POST("/invitations/:code/accept", middleware.AuthMiddleware(), teamCtrl.AcceptInvitation)
	r.POST("/invitations/:code/reject", middleware.AuthMiddleware(), teamCtrl.RejectInvitation)
}

// SetupUserSearchRoutes 设置用户搜索及个人申请路由
func SetupUserSearchRoutes(r *gin.RouterGroup, teamRepo *repositories.TeamRepository) {
	teamCtrl := controllers.NewTeamController(teamRepo, nil, nil, nil, nil, nil, nil)
	users := r.Group("/users")
	users.Use(middleware.AuthMiddleware())
	{
		users.GET("/search", teamCtrl.SearchUsers)
	}

	// 我的申请列表
	applications := r.Group("/applications")
	applications.Use(middleware.AuthMiddleware())
	{
		applications.GET("/my", teamCtrl.GetMyApplications)
	}
}
