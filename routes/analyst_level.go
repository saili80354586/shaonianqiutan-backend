package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
)

func SetupAnalystLevelRoutes(r *gin.RouterGroup, ctrl *controllers.AnalystLevelController) {
	analyst := r.Group("/analyst")
	analyst.Use(middleware.AuthMiddleware(), middleware.AnalystRoleMiddleware())
	{
		analyst.GET("/level", ctrl.GetMyLevel)
		analyst.POST("/level-applications", ctrl.SubmitMyApplication)
		analyst.GET("/level-applications", ctrl.ListMyApplications)
	}

	admin := r.Group("/admin")
	admin.Use(middleware.AuthMiddleware(), middleware.AdminRoleMiddleware())
	{
		perm := middleware.AdminPermissionMiddleware("applications.review")
		admin.GET("/analyst-levels", perm, ctrl.ListLevels)
		admin.GET("/analyst-levels/analysts", perm, ctrl.ListAdminAnalysts)
		admin.PUT("/analysts/:id/level", perm, ctrl.SetAnalystLevel)
		admin.PUT("/analysts/:id/official-partnership", perm, ctrl.SetOfficialPartnership)
		admin.POST("/analysts/:id/growth/refresh", perm, ctrl.RefreshAnalystGrowth)
		admin.POST("/analysts/:id/level-suggestion/apply", perm, ctrl.ApplyLevelSuggestion)
		admin.POST("/analysts/:id/level-suggestion/ignore", perm, ctrl.IgnoreLevelSuggestion)
		admin.GET("/analysts/:id/level-histories", perm, ctrl.ListLevelHistories)
		admin.GET("/analyst-level-applications", perm, ctrl.ListAdminApplications)
		admin.GET("/analyst-level-applications/:id", perm, ctrl.GetApplication)
		admin.POST("/analyst-level-applications/:id/review", perm, ctrl.ReviewApplication)
	}
}
