package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
)

// SetupVideoAnalysisRoutes 设置视频分析路由
func SetupVideoAnalysisRoutes(
	r *gin.RouterGroup,
	vaController *controllers.VideoAnalysisController,
) {
	// 视频分析路由（需要认证，不含角色限制——球员/分析师都能访问）
	va := r.Group("/video-analysis")
	va.Use(middleware.AuthMiddleware())
	{
		// ===== 球员端（所有登录用户可访问）=====
		va.GET("/my", vaController.GetMyAnalyses)           // 球员查看自己的视频分析列表
		va.GET("/my/:id", vaController.GetMyAnalysisDetail) // 球员查看某条分析详情

		// ===== 分析师端（需要分析师权限）=====
		analyst := va.Group("")
		analyst.Use(middleware.AnalystRoleMiddleware())
		{
			// 分析记录管理
			analyst.POST("/create-from-order", vaController.CreateFromOrder) // 从订单创建分析
			analyst.GET("/:id", vaController.GetAnalysis)                    // 获取分析详情
			analyst.GET("/by-order", vaController.GetAnalysisByOrder)        // 根据订单获取分析

			// 评分操作
			analyst.PUT("/:id/scores", vaController.UpdateScores)      // 更新评分
			analyst.POST("/:id/confirm", vaController.ConfirmAnalysis) // 确认分析并生成MD文档

			// 高光时刻管理
			analyst.POST("/highlights", vaController.CreateHighlight)       // 创建高光
			analyst.GET("/:id/highlights", vaController.GetHighlights)      // 获取高光列表
			analyst.PUT("/highlights/:id", vaController.UpdateHighlight)    // 更新高光
			analyst.DELETE("/highlights/:id", vaController.DeleteHighlight) // 删除高光
			analyst.PUT("/highlight/:id", vaController.UpdateHighlight)     // 兼容旧路径
			analyst.DELETE("/highlight/:id", vaController.DeleteHighlight)  // 兼容旧路径
			analyst.POST("/markers/:id/clip", vaController.RetryHighlightClip)
			analyst.GET("/markers/:id/clip", vaController.GetHighlightClip)
			analyst.GET("/markers/:id/clip/download", vaController.DownloadHighlightClip)
			analyst.POST("/:id/clips/export", vaController.ExportHighlightClips)
			analyst.POST("/:id/clips/export/jobs", vaController.CreateHighlightClipsExportJob)
			analyst.GET("/:id/clips/export/jobs", vaController.ListHighlightClipsExportJobs)
			analyst.GET("/:id/clips/export/jobs/:job_id", vaController.GetHighlightClipsExportJob)
			analyst.POST("/:id/clips/export/jobs/:job_id/retry", vaController.RetryHighlightClipsExportJob)
			analyst.GET("/:id/clips/export/jobs/:job_id/download", vaController.DownloadHighlightClipsExportJob)
			analyst.POST("/highlights/:id/clip", vaController.RetryHighlightClip)            // 兼容前端命名
			analyst.GET("/highlights/:id/clip", vaController.GetHighlightClip)               // 兼容前端命名
			analyst.GET("/highlights/:id/clip/download", vaController.DownloadHighlightClip) // 兼容前端命名

			// AI报告生成与管理
			analyst.POST("/generate-ai-report", vaController.GenerateAIReport)   // 触发AI生成
			analyst.GET("/:id/ai-report", vaController.GetAIReport)              // 获取AI报告
			analyst.PUT("/:id/ai-report", vaController.UpdateAIReport)           // 手动修改AI报告
			analyst.POST("/:id/confirm-ai-report", vaController.ConfirmAIReport) // 确认AI报告
		}
	}
}
