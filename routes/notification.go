package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
)

// SetupNotificationRoutes 设置通知路由
func SetupNotificationRoutes(r *gin.RouterGroup, ctrl *controllers.NotificationController) {
	notifications := r.Group("/notifications")
	notifications.Use(middleware.AuthMiddleware())
	{
		notifications.GET("", ctrl.List)
		notifications.GET("/unread", ctrl.ListUnread)
		notifications.GET("/unread-count", ctrl.GetUnreadCount)
		notifications.PUT("/:id/read", ctrl.MarkAsRead)
		notifications.PUT("/read-all", ctrl.MarkAllAsRead)
		notifications.DELETE("/:id", ctrl.Delete)
	}
}
