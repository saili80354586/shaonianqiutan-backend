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
		perm := middleware.AdminPermissionMiddleware

		authAdmin.GET("/me/permissions", adminController.GetMyAdminPermissions)

		// 数据看板 & 运营洞察
		authAdmin.GET("/statistics", perm("dashboard.view"), adminController.GetStatistics)
		authAdmin.GET("/dashboard/stats", perm("dashboard.view"), adminController.GetDashboardStats)
		authAdmin.GET("/dashboard/growth", perm("operations.view"), adminController.GetGrowthData)
		authAdmin.GET("/dashboard/funnel", perm("operations.view"), adminController.GetFunnelData)
		authAdmin.GET("/dashboard/retention", perm("operations.view"), adminController.GetRetentionData)
		authAdmin.GET("/dashboard/top", perm("operations.view"), adminController.GetTopData)
		authAdmin.GET("/dashboard/revenue", perm("finance.manage"), adminController.GetRevenueTrend)
		authAdmin.GET("/settings", perm("settings.manage"), adminController.GetSystemSettings)
		authAdmin.PUT("/settings", perm("settings.manage"), adminController.UpdateSystemSettings)

		// 用户管理
		authAdmin.GET("/users", perm("users.manage"), adminController.GetUserList)
		authAdmin.PUT("/users/:id", perm("users.manage"), adminController.UpdateUser)
		authAdmin.PUT("/users/:id/status", perm("users.manage"), adminController.UpdateUserStatus)
		authAdmin.DELETE("/users/:id", perm("users.manage"), adminController.DeleteUser)

		// 订单管理
		authAdmin.GET("/orders", perm("orders.manage"), adminController.GetAllOrders)
		authAdmin.GET("/orders/stats", perm("orders.manage"), adminController.GetOrderStats)
		authAdmin.GET("/orders/assignments", perm("dispatch.manage"), adminController.GetAssignmentRecords)
		authAdmin.GET("/orders/:id/status-history", perm("orders.manage"), adminController.GetOrderStatusHistory)
		authAdmin.POST("/orders/:id/assign", perm("dispatch.manage"), adminController.AssignOrder)
		authAdmin.DELETE("/orders/:id", perm("orders.manage"), adminController.CancelOrder)

		// 结算管理
		authAdmin.GET("/settlements", perm("finance.manage"), adminController.GetSettlementList)
		authAdmin.POST("/settlements/process", perm("finance.manage"), adminController.ProcessSettlement)

		// 分析师管理
		authAdmin.GET("/analysts", perm("applications.review"), adminController.GetAnalystList)
		authAdmin.GET("/analysts/available", perm("dispatch.manage"), adminController.GetAvailableAnalysts)
		authAdmin.PUT("/analysts/:id/audit", perm("applications.review"), adminController.AuditAnalyst)
		authAdmin.PUT("/analysts/:id/status", perm("applications.review"), adminController.UpdateAnalystStatus)
		authAdmin.GET("/analysts/:id/income", perm("finance.manage"), adminController.GetAnalystIncomeStats)

		// 报告管理
		authAdmin.GET("/reports/pending", perm("reports.review"), adminController.GetPendingReports)
		authAdmin.POST("/reports/:id/review", perm("reports.review"), adminController.ReviewReport)
		authAdmin.GET("/reports/:id/download", perm("reports.review"), adminController.DownloadReportDoc)    // 下载报告 MD 文档
		authAdmin.POST("/reports/:id/upload-report", perm("reports.review"), adminController.UploadAIReport) // 上传 AI Word 报告
		authAdmin.POST("/reports/:id/upload-video", perm("reports.review"), adminController.UploadAIVideo)   // 上传 AI 视频分析

		// 视频分析 MD 文档下载（来自 video_analyses）
		authAdmin.GET("/video-analysis/:id/download", perm("reports.review"), adminController.DownloadVideoAnalysisDoc) // 下载视频分析 MD 文档

		// 举报处理
		authAdmin.GET("/content-reports", perm("content.manage"), adminController.GetContentReports)
		authAdmin.GET("/content-reports/:id", perm("content.manage"), adminController.GetContentReportDetail)
		authAdmin.POST("/content-reports/:id/handle", perm("content.manage"), adminController.HandleContentReport)

		// 敏感词配置
		authAdmin.GET("/sensitive-words", perm("content.manage"), adminController.GetSensitiveWords)
		authAdmin.POST("/sensitive-words", perm("content.manage"), adminController.CreateSensitiveWord)
		authAdmin.PUT("/sensitive-words/:id", perm("content.manage"), adminController.UpdateSensitiveWord)
		authAdmin.DELETE("/sensitive-words/:id", perm("content.manage"), adminController.DeleteSensitiveWord)

		// 平台公告
		authAdmin.GET("/announcements", perm("platform.manage"), adminController.GetPlatformAnnouncements)
		authAdmin.POST("/announcements", perm("platform.manage"), adminController.CreatePlatformAnnouncement)
		authAdmin.PUT("/announcements/:id", perm("platform.manage"), adminController.UpdatePlatformAnnouncement)
		authAdmin.DELETE("/announcements/:id", perm("platform.manage"), adminController.DeletePlatformAnnouncement)

		// 轮播图
		authAdmin.GET("/banners", perm("platform.manage"), adminController.GetBanners)
		authAdmin.POST("/banners", perm("platform.manage"), adminController.CreateBanner)
		authAdmin.PUT("/banners/:id", perm("platform.manage"), adminController.UpdateBanner)
		authAdmin.DELETE("/banners/:id", perm("platform.manage"), adminController.DeleteBanner)

		// FAQ
		authAdmin.GET("/faqs", perm("platform.manage"), adminController.GetFAQs)
		authAdmin.POST("/faqs", perm("platform.manage"), adminController.CreateFAQ)
		authAdmin.PUT("/faqs/:id", perm("platform.manage"), adminController.UpdateFAQ)
		authAdmin.DELETE("/faqs/:id", perm("platform.manage"), adminController.DeleteFAQ)

		// 登录日志
		authAdmin.GET("/login-logs", perm("login_logs.view"), adminController.GetLoginLogs)
		authAdmin.GET("/login-logs/stats", perm("login_logs.view"), adminController.GetLoginLogStats)

		// 操作审计日志
		authAdmin.GET("/audit-logs", perm("audit.view"), adminController.GetAuditLogs)

		// 角色权限与异常管控
		authAdmin.GET("/role-permissions", perm("role_permissions.view"), adminController.GetRolePermissions)
		authAdmin.GET("/role-permissions/history", perm("role_permissions.view"), adminController.GetAdminRoleAssignmentHistory)
		authAdmin.POST("/role-permissions/roles", perm("role_permissions.manage"), adminController.CreateAdminRole)
		authAdmin.PUT("/role-permissions/roles/:roleKey", perm("role_permissions.manage"), adminController.UpdateAdminRole)
		authAdmin.PUT("/role-permissions/roles/:roleKey/status", perm("role_permissions.manage"), adminController.UpdateAdminRoleStatus)
		authAdmin.PUT("/role-permissions/assignments/:userId", perm("role_permissions.manage"), adminController.AssignAdminRole)
		authAdmin.POST("/role-permissions/batch-assignments", perm("role_permissions.manage"), adminController.BatchAssignAdminRole)
		authAdmin.GET("/exceptions", perm("exceptions.view"), adminController.GetExceptions)
	}
}
