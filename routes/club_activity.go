package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
)

// SetupClubActivityRoutes 设置俱乐部活动路由
func SetupClubActivityRoutes(r *gin.RouterGroup, ctrl *controllers.ClubActivityController) {
	clubs := r.Group("/clubs")
	{
		// 公开接口：获取活动列表
		clubs.GET("/:clubId/activities", ctrl.ListActivities)
		// 公开接口：报名活动
		clubs.POST("/:clubId/activities/:id/register", ctrl.RegisterActivity)

		// 球员认证接口：取消报名
		clubs.POST("/:clubId/activities/:id/cancel", middleware.AuthMiddleware(), ctrl.CancelRegistration)

		// 管理接口：需要认证
		admin := clubs.Group("/:clubId/activities")
		admin.Use(middleware.AuthMiddleware())
		{
			admin.POST("", ctrl.CreateActivity)
			admin.PUT("/:id", ctrl.UpdateActivity)
			admin.DELETE("/:id", ctrl.DeleteActivity)
			admin.POST("/:id/publish", ctrl.PublishActivity)
			admin.POST("/:id/unpublish", ctrl.UnpublishActivity)
			admin.GET("/:id/registrations", ctrl.ListRegistrations)
			admin.PUT("/:id/registrations/:regId", ctrl.UpdateRegistrationStatus)
			admin.POST("/:id/registrations/batch", ctrl.BatchUpdateRegistrationStatus)
			admin.GET("/:id/registrations/export", ctrl.ExportRegistrations)
		}
	}

	// 全局公开活动接口
	activities := r.Group("/activities")
	{
		activities.GET("", ctrl.ListPublicActivities)
		activities.GET("/map", ctrl.GetActivitiesMap)
		activities.GET("/:id", ctrl.GetPublicActivity)
	}

	// 用户认证接口：我的报名
	r.GET("/user/registrations", middleware.AuthMiddleware(), ctrl.GetMyRegistrations)
}
