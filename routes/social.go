package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
)

// SetupSocialRoutes 设置社交互动路由
func SetupSocialRoutes(r *gin.RouterGroup, socialController *controllers.SocialController) {
	// 需要认证的路由
	auth := r.Group("")
	auth.Use(middleware.AuthMiddleware())
	{
		// 点赞
		auth.POST("/likes", socialController.ToggleLike)
		auth.DELETE("/likes", socialController.RemoveLike)
		auth.GET("/likes/my", socialController.GetMyLikes)

		// 收藏
		auth.POST("/favorites", socialController.ToggleFavorite)
		auth.DELETE("/favorites", socialController.RemoveFavorite)
		auth.GET("/favorites/my", socialController.GetMyFavorites)

		// 评论
		auth.POST("/comments", socialController.CreateComment)
		auth.DELETE("/comments/:id", socialController.DeleteComment)

		// 成就
		auth.GET("/achievements", socialController.GetAchievements)

		// 动态帖子
		auth.POST("/posts", socialController.CreatePost)
		auth.DELETE("/posts/:id", socialController.DeletePost)
		auth.POST("/posts/:id/like", socialController.TogglePostLike)

		// 关注
		auth.POST("/follow", socialController.ToggleFollow)
	}

	// 关注列表（公开查看）
	r.GET("/followers/:userId", socialController.GetFollowers)
	r.GET("/following/:userId", socialController.GetFollowing)
	r.GET("/follow/status/:userId", socialController.GetFollowStatus)

	// 公开路由
	// 点赞列表（公开查看）
	r.GET("/likes", socialController.GetLikes)
	// 评论列表（公开查看）
	r.GET("/comments", socialController.GetComments)
	// 动态流（公开查看）
	r.GET("/feed", socialController.GetFeed)
}
