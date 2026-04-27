package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
)

// SetupClubOrderRoutes 设置俱乐部订单路由
func SetupClubOrderRoutes(r *gin.RouterGroup, orderController *controllers.ClubOrderController) {
	// 俱乐部订单路由
	orders := r.Group("/club/orders")
	orders.Use(middleware.AuthMiddleware())
	{
		orders.GET("", orderController.GetOrders)
		orders.GET("/stats", orderController.GetStats)
		orders.POST("/batch", orderController.CreateBatchOrders)
		orders.GET("/:id", orderController.GetOrder)
		orders.POST("/:id/cancel", orderController.CancelOrder)
	}
}
