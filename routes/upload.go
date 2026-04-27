package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
)

// SetupUploadRoutes 设置上传路由
func SetupUploadRoutes(r *gin.RouterGroup, uploadController *controllers.UploadController) {
	upload := r.Group("/upload")
	upload.Use(middleware.AuthMiddleware())
	{
		upload.POST("/file", uploadController.UploadFile)
		upload.POST("/avatar", uploadController.UploadAvatar)
	}
}
