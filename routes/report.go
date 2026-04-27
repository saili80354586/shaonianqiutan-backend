package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
)

// SetupReportRoutes 设置报告路由
func SetupReportRoutes(r *gin.RouterGroup, reportController *controllers.ReportController) {
	report := r.Group("/reports")
	report.Use(middleware.AuthMiddleware())
	{
		report.POST("/", reportController.CreateReport)
		report.GET("/:id", reportController.GetReportDetail)
		report.GET("/:id/download", reportController.DownloadReport)
		report.GET("/my", reportController.GetMyReports)
		report.GET("/published", reportController.GetMyPublishedReports)
		report.POST("/:id/regenerate", reportController.RegeneratePdf)
		report.GET("/statistics", reportController.GetReportStatistics)
	}
}
