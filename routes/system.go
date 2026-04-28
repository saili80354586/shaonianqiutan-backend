package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/controllers"
)

func SetupSystemRoutes(r *gin.RouterGroup, adminController *controllers.AdminController) {
	system := r.Group("/system")
	{
		system.GET("/public-settings", adminController.GetPublicSystemSettings)
	}
}
