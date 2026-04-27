package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
)

// SetupTeamHomeRoutes 设置球队主页路由（统一端点 /teams/:teamId/homepage）
func SetupTeamHomeRoutes(r *gin.RouterGroup, teamHomeController *controllers.CoachTeamHomeController) {
	teamHome := r.Group("/teams")
	teamHome.Use(middleware.AuthMiddleware())
	teamHome.Use(middleware.TeamAccessMiddleware())
	{
		// 球队主页管理
		teamHome.GET("/:teamId/homepage", teamHomeController.GetTeamHome)
		teamHome.PUT("/:teamId/homepage", teamHomeController.SaveTeamHome)

		// 分模块更新
		teamHome.PUT("/:teamId/homepage/hero", teamHomeController.UpdateHero)
		teamHome.PUT("/:teamId/homepage/about", teamHomeController.UpdateAbout)
		teamHome.PUT("/:teamId/homepage/contact", teamHomeController.UpdateContact)

		// 荣誉管理
		teamHome.POST("/:teamId/homepage/honors", teamHomeController.AddHonor)
		teamHome.DELETE("/:teamId/homepage/honors/:honorId", teamHomeController.DeleteHonor)

		// 动态管理
		teamHome.GET("/:teamId/homepage/dynamics", teamHomeController.GetDynamics)
		teamHome.POST("/:teamId/homepage/dynamics", teamHomeController.AddDynamic)
		teamHome.DELETE("/:teamId/homepage/dynamics/:dynamicId", teamHomeController.DeleteDynamic)
		teamHome.PUT("/:teamId/homepage/dynamics", teamHomeController.UpdateDynamics)
	}
}
