package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
)

// SetupAnalystRoutes 设置分析师路由
func SetupAnalystRoutes(
	r *gin.RouterGroup,
	analystController *controllers.AnalystController,
) {
	// 公开路由(获取分析师列表和详情)
	analyst := r.Group("/analysts")
	{
		analyst.GET("/", analystController.GetAnalystList)                      // 获取分析师列表
		analyst.GET("/public", analystController.GetAnalystPublicProfileByUser) // 通过 user_id 获取分析师公开主页
		analyst.GET("/:id", analystController.GetAnalystByID)                   // 获取分析师详情
		analyst.GET("/:id/public", analystController.GetAnalystPublicProfile)   // 通过 analyst_id 获取分析师公开主页
		analyst.POST("/:id/inquiries", analystController.CreateInquiry)         // 提交咨询意向
	}

	// 分析师专用路由(需要认证和分析师权限)
	myAnalyst := r.Group("/analyst")
	myAnalyst.Use(middleware.AuthMiddleware())
	myAnalyst.Use(middleware.AnalystRoleMiddleware())
	{
		myAnalyst.GET("/profile", analystController.GetMyAnalystProfile) // 获取我的分析师资料
		myAnalyst.PUT("/profile", analystController.UpdateMyProfile)     // 更新我的资料
		myAnalyst.GET("/orders", analystController.GetMyOrders)          // 获取我的订单列表（保留）
		myAnalyst.GET("/revenue", analystController.GetMyRevenue)        // 获取我的收益统计（保留）

		// ===== 新增 =====
		myAnalyst.GET("/dashboard-stats", analystController.GetDashboardStats)           // 工作台统计
		myAnalyst.GET("/orders/pending", analystController.GetPendingOrders)             // 待处理订单
		myAnalyst.GET("/orders/active", analystController.GetActiveOrders)               // 进行中订单
		myAnalyst.GET("/orders/history", analystController.GetHistoryOrders)             // 历史订单
		myAnalyst.POST("/orders/:id/accept", analystController.AcceptOrder)              // 接单
		myAnalyst.POST("/orders/:id/reject", analystController.RejectOrder)              // 拒绝
		myAnalyst.GET("/income-details", analystController.GetIncomeDetails)             // 收益明细
		myAnalyst.GET("/income-trend", analystController.GetIncomeTrend)                 // 收益趋势
		myAnalyst.POST("/orders/:id/submit-report", analystController.SubmitReport)      // 提交报告
		myAnalyst.GET("/orders/:id/download", analystController.DownloadReportDoc)       // 下载报告 MD 文档
		myAnalyst.GET("/orders/:id/ai-report", analystController.DownloadAIReport)       // 下载 AI 报告
		myAnalyst.POST("/orders/:id/ai-report/upload", analystController.UploadAIReport) // 上传本地编辑后的 AI 报告
	}
}
