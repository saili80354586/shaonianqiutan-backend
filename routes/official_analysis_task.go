package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
)

func SetupOfficialAnalysisTaskRoutes(r *gin.RouterGroup, ctrl *controllers.OfficialAnalysisTaskController) {
	publicTopics := r.Group("/official-event-topics")
	{
		publicTopics.GET("", ctrl.ListPublicEventTopics)
		publicTopics.GET("/:matchName", ctrl.GetPublicEventTopic)
	}

	admin := r.Group("/admin/official-analysis-tasks")
	admin.Use(middleware.AuthMiddleware(), middleware.AdminRoleMiddleware())
	{
		perm := middleware.AdminPermissionMiddleware("content.manage")
		admin.GET("", perm, ctrl.ListAdminTasks)
		admin.POST("", perm, ctrl.CreateAdminTask)
		admin.POST("/batch", perm, ctrl.BatchCreateAdminTasks)
		admin.GET("/:id", perm, ctrl.GetAdminTask)
		admin.GET("/:id/acceptances", perm, ctrl.ListTaskAcceptances)
		admin.GET("/:id/submissions", perm, ctrl.ListTaskSubmissions)
		admin.PUT("/:id", perm, ctrl.UpdateAdminTask)
		admin.POST("/:id/publish", perm, ctrl.PublishAdminTask)
		admin.POST("/:id/close", perm, ctrl.CloseAdminTask)
	}
	adminSubmissions := r.Group("/admin/official-analysis-submissions")
	adminSubmissions.Use(middleware.AuthMiddleware(), middleware.AdminRoleMiddleware())
	{
		perm := middleware.AdminPermissionMiddleware("content.manage")
		adminSubmissions.GET("", perm, ctrl.ListAdminSubmissions)
		adminSubmissions.GET("/:id", perm, ctrl.GetSubmission)
		adminSubmissions.POST("/:id/review", perm, ctrl.ReviewSubmission)
		adminSubmissions.POST("/:id/adopt", perm, ctrl.AdoptSubmission)
	}
	adminAdoptions := r.Group("/admin/official-content-adoptions")
	adminAdoptions.Use(middleware.AuthMiddleware(), middleware.AdminRoleMiddleware())
	{
		contentPerm := middleware.AdminPermissionMiddleware("content.manage")
		financePerm := middleware.AdminPermissionMiddleware("finance.manage")
		adminAdoptions.GET("", contentPerm, ctrl.ListOfficialMaterials)
		adminAdoptions.PUT("/:id/public", contentPerm, ctrl.UpdateAdoptionPublic)
		adminAdoptions.POST("/:id/publish-records", contentPerm, ctrl.CreatePublishRecord)
		adminAdoptions.PUT("/:id/publish-records/:recordId", contentPerm, ctrl.UpdatePublishRecord)
		adminAdoptions.DELETE("/:id/publish-records/:recordId", contentPerm, ctrl.DeletePublishRecord)
		adminAdoptions.POST("/:id/playback-bonus", financePerm, ctrl.CreatePlaybackBonus)
	}
	adminTopics := r.Group("/admin/official-event-topics")
	adminTopics.Use(middleware.AuthMiddleware(), middleware.AdminRoleMiddleware())
	{
		perm := middleware.AdminPermissionMiddleware("content.manage")
		adminTopics.GET("", perm, ctrl.ListAdminEventTopics)
		adminTopics.PUT("/:matchName", perm, ctrl.SaveAdminEventTopic)
	}
	adminRewards := r.Group("/admin/analyst-rewards")
	adminRewards.Use(middleware.AuthMiddleware(), middleware.AdminRoleMiddleware())
	{
		perm := middleware.AdminPermissionMiddleware("finance.manage")
		adminRewards.GET("", perm, ctrl.ListAdminRewards)
		adminRewards.GET("/settlement-batches", perm, ctrl.ListRewardSettlementBatches)
		adminRewards.GET("/settlement-batches/:id", perm, ctrl.GetRewardSettlementBatch)
		adminRewards.POST("/batch-settle", perm, ctrl.BatchSettleRewards)
		adminRewards.POST("/:id/settle", perm, ctrl.SettleReward)
		adminRewards.POST("/:id/reverse", perm, ctrl.ReverseReward)
	}

	analyst := r.Group("/analyst/official-analysis-tasks")
	analyst.Use(middleware.AuthMiddleware(), middleware.AnalystRoleMiddleware())
	{
		analyst.GET("", ctrl.ListAvailableTasks)
		analyst.GET("/my", ctrl.ListMyTasks)
		analyst.GET("/mine", ctrl.ListMyTasks)
		analyst.GET("/:id", ctrl.GetAvailableTask)
		analyst.POST("/:id/accept", ctrl.AcceptTask)
		analyst.POST("/:id/submit", ctrl.SubmitTask)
	}
	analystRewards := r.Group("/analyst")
	analystRewards.Use(middleware.AuthMiddleware(), middleware.AnalystRoleMiddleware())
	{
		analystRewards.GET("/official-rewards", ctrl.ListMyRewards)
		analystRewards.GET("/official-adoptions", ctrl.ListMyAdoptions)
	}
}
