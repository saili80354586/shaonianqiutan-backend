package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
)

// SetupAdminRoutes 设置管理后台路由
func SetupAdminRoutes(r *gin.RouterGroup, adminController *controllers.AdminController) {
	// 公开路由: 管理员登录
	admin := r.Group("/admin")
	{
		admin.POST("/login", adminController.AdminLogin)
	}

	// 需要认证的路由
	authAdmin := r.Group("/admin")
	authAdmin.Use(middleware.AuthMiddleware())
	authAdmin.Use(middleware.AdminRoleMiddleware())
	{
		// 数据看板 & 运营洞察
		authAdmin.GET("/statistics", adminController.GetStatistics)
		authAdmin.GET("/dashboard/stats", adminController.GetDashboardStats)
		authAdmin.GET("/dashboard/growth", adminController.GetGrowthData)
		authAdmin.GET("/dashboard/funnel", adminController.GetFunnelData)
		authAdmin.GET("/dashboard/retention", adminController.GetRetentionData)
		authAdmin.GET("/dashboard/top", adminController.GetTopData)
		authAdmin.GET("/dashboard/revenue", adminController.GetRevenueTrend)
		authAdmin.GET("/settings", adminController.GetSystemSettings)
		authAdmin.PUT("/settings", adminController.UpdateSystemSettings)

		// 用户管理
		authAdmin.GET("/users", adminController.GetUserList)
		authAdmin.PUT("/users/:id/status", adminController.UpdateUserStatus)
		authAdmin.DELETE("/users/:id", adminController.DeleteUser)

		// 订单管理
		authAdmin.GET("/orders", adminController.GetAllOrders)
		authAdmin.GET("/orders/stats", adminController.GetOrderStats)
		authAdmin.GET("/orders/assignments", adminController.GetAssignmentRecords)
		authAdmin.GET("/orders/:id/status-history", adminController.GetOrderStatusHistory)
		authAdmin.POST("/orders/:id/assign", adminController.AssignOrder)
		authAdmin.DELETE("/orders/:id", adminController.CancelOrder)

		// 结算管理
		authAdmin.GET("/settlements", adminController.GetSettlementList)
		authAdmin.POST("/settlements/process", adminController.ProcessSettlement)

		// 分析师管理
		authAdmin.GET("/analysts", adminController.GetAnalystList)
		authAdmin.GET("/analysts/available", adminController.GetAvailableAnalysts)
		authAdmin.PUT("/analysts/:id/audit", adminController.AuditAnalyst)
		authAdmin.PUT("/analysts/:id/status", adminController.UpdateAnalystStatus)
		authAdmin.GET("/analysts/:id/income", adminController.GetAnalystIncomeStats)

		// 报告管理
		authAdmin.GET("/reports/pending", adminController.GetPendingReports)
		authAdmin.POST("/reports/:id/review", adminController.ReviewReport)
		authAdmin.GET("/reports/:id/download", adminController.DownloadReportDoc)    // 下载报告 MD 文档
		authAdmin.POST("/reports/:id/upload-report", adminController.UploadAIReport) // 上传 AI Word 报告
		authAdmin.POST("/reports/:id/upload-video", adminController.UploadAIVideo)   // 上传 AI 视频分析

		// 视频分析 MD 文档下载（来自 video_analyses）
		authAdmin.GET("/video-analysis/:id/download", adminController.DownloadVideoAnalysisDoc) // 下载视频分析 MD 文档

		// 举报处理
		authAdmin.GET("/content-reports", adminController.GetContentReports)
		authAdmin.GET("/content-reports/:id", adminController.GetContentReportDetail)
		authAdmin.POST("/content-reports/:id/handle", adminController.HandleContentReport)

		// 敏感词配置
		authAdmin.GET("/sensitive-words", adminController.GetSensitiveWords)
		authAdmin.POST("/sensitive-words", adminController.CreateSensitiveWord)
		authAdmin.PUT("/sensitive-words/:id", adminController.UpdateSensitiveWord)
		authAdmin.DELETE("/sensitive-words/:id", adminController.DeleteSensitiveWord)

		// 平台公告
		authAdmin.GET("/announcements", adminController.GetPlatformAnnouncements)
		authAdmin.POST("/announcements", adminController.CreatePlatformAnnouncement)
		authAdmin.PUT("/announcements/:id", adminController.UpdatePlatformAnnouncement)
		authAdmin.DELETE("/announcements/:id", adminController.DeletePlatformAnnouncement)

		// 轮播图
		authAdmin.GET("/banners", adminController.GetBanners)
		authAdmin.POST("/banners", adminController.CreateBanner)
		authAdmin.PUT("/banners/:id", adminController.UpdateBanner)
		authAdmin.DELETE("/banners/:id", adminController.DeleteBanner)

		// FAQ
		authAdmin.GET("/faqs", adminController.GetFAQs)
		authAdmin.POST("/faqs", adminController.CreateFAQ)
		authAdmin.PUT("/faqs/:id", adminController.UpdateFAQ)
		authAdmin.DELETE("/faqs/:id", adminController.DeleteFAQ)

		// 登录日志
		authAdmin.GET("/login-logs", adminController.GetLoginLogs)
		authAdmin.GET("/login-logs/stats", adminController.GetLoginLogStats)
	}
}
