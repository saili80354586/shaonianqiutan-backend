package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
)

// SetupScoutMapRoutes 设置球探地图路由
func SetupScoutMapRoutes(r *gin.RouterGroup, mapController *controllers.MapController) {
	scout := r.Group("/scout")
	{
		// 新版分层地图接口
		scout.GET("/map/national", mapController.GetNationalMapData)
		scout.GET("/map/provincial", mapController.GetProvincialMapData)
		scout.GET("/map/city", mapController.GetCityMapData)

		// Stage 4 新增接口
		scout.GET("/map/dashboard", mapController.GetDashboardStats)
		scout.GET("/map/overseas", mapController.GetOverseasPlayers)
		scout.GET("/map/recommendations", mapController.GetRecommendations)
		scout.GET("/map/rising-stars", mapController.GetRisingStars)

		// 兼容旧版 V2 接口
		scout.GET("/map", mapController.GetScoutMapData)
		scout.GET("/by-province", mapController.GetPlayersByProvince)
	}

	// 公开球员地图资料（放在 scout 组下避免与 /players/:playerId 冲突）
	scout.GET("/players/:userId/map-profile", mapController.GetPlayerMapProfile)

	// 公开推荐接口（兼容前端路径 /map/recommendations）
	r.GET("/map/recommendations", mapController.GetRecommendations)

	// 需要认证的地图接口
	authMap := r.Group("/map")
	authMap.Use(middleware.AuthMiddleware())
	{
		authMap.POST("/export-compare", mapController.ExportCompare)
		authMap.GET("/my-rank", mapController.GetMyRank)
	}
}
