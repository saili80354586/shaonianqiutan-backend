package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/config"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
)

// SetupPaymentRoutes 设置支付路由
func SetupPaymentRoutes(r *gin.RouterGroup, paymentController *controllers.PaymentController) {
	payment := r.Group("/payment")
	payment.Use(middleware.AuthMiddleware())
	{
		payment.POST("/", paymentController.CreatePayment)
		if config.IsDevMode() {
			payment.POST("/simulate", paymentController.SimulatePay)
		}
		payment.GET("/:id", paymentController.GetPaymentStatus)
		payment.POST("/callback", paymentController.PaymentCallback)
		payment.POST("/:id/refund", paymentController.Refund)
	}

	// 订单支付状态查询（公开路由，也需要认证）
	orderPayment := r.Group("/order-payment")
	orderPayment.Use(middleware.AuthMiddleware())
	{
		orderPayment.GET("/:orderId/payment-status", paymentController.GetOrderPaymentStatus)
	}
}
