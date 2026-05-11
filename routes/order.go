package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
)

// SetupOrderRoutes 设置订单路由
func SetupOrderRoutes(r *gin.RouterGroup, orderController *controllers.OrderController) {
	order := r.Group("/orders")
	order.Use(middleware.AuthMiddleware())
	{
		// 兼容带/不带尾斜杠的前端请求，避免浏览器对 POST 重定向报 Network Error
		order.POST("", orderController.CreateOrder)
		order.POST("/", orderController.CreateOrder)

		// 获取订单列表
		order.GET("", orderController.GetMyOrders)
		order.GET("/", orderController.GetMyOrders)

		// 获取订单详情
		order.GET("/:id", orderController.GetOrderDetail)

		// 支付后补充订单信息（上传视频和球员资料）
		order.POST("/:id/supplement", orderController.SupplementOrder)
		order.POST("/:id/source-video/upload-url", orderController.CreateSourceVideoUploadIntent)
		order.POST("/:id/source-video/confirm", orderController.ConfirmSourceVideoUpload)

		// 更新订单状态
		order.PUT("/:id/status", orderController.UpdateOrderStatus)

		// 取消订单
		order.DELETE("/:id", orderController.CancelOrder)

		// 获取用户订单统计
		order.GET("/statistics", orderController.GetOrderStatistics)

		// 下载 AI 报告
		order.GET("/:id/ai-report", orderController.DownloadAIReport)
	}
}
