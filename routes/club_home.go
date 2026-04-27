package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
)

// SetupClubHomeRoutes 设置俱乐部主页路由
func SetupClubHomeRoutes(r *gin.RouterGroup, clubHomeCtrl *controllers.ClubHomeController) {
	clubs := r.Group("/clubs")
	// 公开访问的俱乐部主页
	{
		clubs.GET("/:clubId/home", clubHomeCtrl.GetClubHome)
		clubs.GET("/:clubId/home/news", clubHomeCtrl.GetNews)
	}
	// 需要认证的俱乐部主页管理
	clubs.Use(middleware.AuthMiddleware())
	{
		clubs.PUT("/:clubId/home", clubHomeCtrl.SaveClubHome)
		clubs.PUT("/:clubId/home/hero", clubHomeCtrl.UpdateHero)
		clubs.PUT("/:clubId/home/about", clubHomeCtrl.UpdateAbout)
		clubs.PUT("/:clubId/home/contact", clubHomeCtrl.UpdateContact)
		clubs.PUT("/:clubId/home/facilities", clubHomeCtrl.UpdateFacilities)
		clubs.PUT("/:clubId/home/recruitment", clubHomeCtrl.UpdateRecruitment)
		clubs.PUT("/:clubId/home/social-links", clubHomeCtrl.UpdateSocialLinks)
		clubs.PUT("/:clubId/home/modules", clubHomeCtrl.UpdateModules)
		clubs.PUT("/:clubId/home/teams", clubHomeCtrl.UpdateTeams)
		clubs.PUT("/:clubId/home/coaches", clubHomeCtrl.UpdateCoaches)
		clubs.PUT("/:clubId/home/players", clubHomeCtrl.UpdatePlayers)
		clubs.PUT("/:clubId/home/news", clubHomeCtrl.UpdateNews)
	}
}
