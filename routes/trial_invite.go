package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
)

// SetupTrialInviteRoutes 设置试训邀请路由
func SetupTrialInviteRoutes(r *gin.RouterGroup, trialInviteController *controllers.TrialInviteController) {
	invites := r.Group("/trial-invites")
	invites.Use(middleware.AuthMiddleware())
	{
		invites.POST("", trialInviteController.CreateTrialInvite)
		invites.GET("/my", trialInviteController.GetMyTrialInvites)
		invites.PUT("/:id/respond", trialInviteController.RespondTrialInvite)
	}
}
