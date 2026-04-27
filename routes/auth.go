package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
)

// SetupAuthRoutes 设置认证路由
func SetupAuthRoutes(r *gin.RouterGroup, authController *controllers.AuthController) {
	auth := r.Group("/auth")
	{
		auth.POST("/send-code", authController.SendCode)
		auth.POST("/verify-code", authController.VerifyCode)
		auth.POST("/register", authController.Register)
		auth.POST("/login", authController.Login)
		auth.POST("/reset-password", authController.ResetPassword)

		// 需要认证的路由
		auth.Use(middleware.AuthMiddleware())
		{
			auth.GET("/me", authController.GetMe)
			auth.PUT("/me", authController.UpdateMe)
			auth.POST("/refresh-token", authController.RefreshToken)
		}
	}
}
