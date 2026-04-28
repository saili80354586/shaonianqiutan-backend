package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
)

// SetupPlayerRoutes 设置球员资料路由
func SetupPlayerRoutes(r *gin.RouterGroup, playerController *controllers.PlayerController) {
	// 球员资料路由（需要认证）
	player := r.Group("/player")
	player.Use(middleware.AuthMiddleware())
	{
		player.GET("/profile", playerController.GetProfile)
		player.PUT("/profile", playerController.UpdateProfile)
		player.PATCH("/profile/partial", playerController.PatchProfile)
		// 体测记录 CRUD
		player.GET("/physical-tests", playerController.GetPhysicalTests)
		player.POST("/physical-tests", playerController.CreatePhysicalTest)
		player.PUT("/physical-tests/:id", playerController.UpdatePhysicalTest)
		player.DELETE("/physical-tests/:id", playerController.DeletePhysicalTest)
	}
}

// SetupPlayerPublicRoutes 设置球员公开路由（无需认证）
func SetupPlayerPublicRoutes(r *gin.RouterGroup, playerController *controllers.PlayerController) {
	r.GET("/players/:playerId/public", playerController.GetPlayerPublicProfile)
	r.GET("/players/:playerId/homepage", playerController.GetHomepage)
	r.GET("/players/:playerId/physical-tests", playerController.GetPublicPhysicalTests)
}
