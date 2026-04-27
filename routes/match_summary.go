package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
)

// SetupMatchSummaryRoutes 设置比赛总结路由
func SetupMatchSummaryRoutes(
	r *gin.RouterGroup,
	ctrl *controllers.MatchSummaryController,
	reviewCtrl *controllers.PlayerReviewController,
	videoCtrl *controllers.MatchVideoController,
) {
	// ===== 比赛总结 CRUD =====
	matchSummaries := r.Group("/match-summaries")
	matchSummaries.Use(middleware.AuthMiddleware())
	{
		matchSummaries.POST("", ctrl.Create)           // M1: 教练创建比赛
		matchSummaries.GET("/:id", ctrl.Get)           // M3: 获取比赛详情
		matchSummaries.PUT("/:id", ctrl.Update)        // M4: 教练更新比赛
		matchSummaries.DELETE("/:id", ctrl.Delete)     // M5: 教练删除比赛

		// 教练整体点评
		matchSummaries.POST("/:id/coach-review", ctrl.SubmitCoachReview) // M8

		// 封面图
		matchSummaries.POST("/:id/cover-image", ctrl.UpdateCoverImage) // M9

		// 催办未提交自评的球员
		matchSummaries.POST("/:id/remind", ctrl.Remind) // M10

		// 比赛统计（教练/俱乐部）
		matchSummaries.GET("/stats", ctrl.GetStats) // M11

		// ===== 视频链接 =====
		matchSummaries.POST("/:id/videos", videoCtrl.AddVideo)               // M6: 添加视频
		matchSummaries.DELETE("/:id/videos/:videoId", videoCtrl.DeleteVideo) // M7: 删除视频
		matchSummaries.GET("/:id/videos", videoCtrl.ListVideos)              // 获取视频列表

		// ===== 球员自评 =====
		matchSummaries.POST("/:id/player-review", reviewCtrl.SubmitReview)          // P3: 球员提交自评
		matchSummaries.GET("/:id/player-review", reviewCtrl.GetReview)              // P2: 获取自己的自评
		matchSummaries.POST("/:id/coach-player-review", reviewCtrl.SubmitCoachPlayerReview) // C4: 教练对球员点评
	}

	// ===== 教练维度 =====
	coach := r.Group("/coach")
	coach.Use(middleware.AuthMiddleware())
	{
		coach.GET("/match-summaries", ctrl.ListByCoach) // 教练的比赛列表
	}

	// ===== 球员维度 =====
	players := r.Group("/players")
	players.Use(middleware.AuthMiddleware())
	{
		players.GET("/:playerId/match-summaries", reviewCtrl.ListByPlayer) // P1: 球员的比赛列表
	}

	// 注意：球队比赛总结路由 /teams/:teamId/match-summaries 已在 SetupTeamRoutes 中注册
}
