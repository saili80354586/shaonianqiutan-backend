package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
)

// SetupMessageRoutes 设置私信路由
func SetupMessageRoutes(r *gin.RouterGroup, messageController *controllers.MessageController) {
	// 需要认证的路由
	auth := r.Group("/messages")
	auth.Use(middleware.AuthMiddleware())
	{
		// 发送私信
		auth.POST("", messageController.SendMessage)
		// 获取与某用户的私信列表
		auth.GET("/user/:userId", messageController.GetMessages)
		// 获取会话列表
		auth.GET("/conversations", messageController.GetConversations)
		// 标记单条已读
		auth.PUT("/:id/read", messageController.MarkAsRead)
		// 标记会话已读
		auth.PUT("/user/:userId/read", messageController.MarkConversationAsRead)
		// 获取未读数
		auth.GET("/unread-count", messageController.GetUnreadCount)
		// 删除私信
		auth.DELETE("/:id", messageController.DeleteMessage)
	}
}
