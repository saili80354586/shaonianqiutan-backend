package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
)

// SetupPhysicalTestRoutes 设置俱乐部体测路由
func SetupPhysicalTestRoutes(r *gin.RouterGroup, ptController *controllers.PhysicalTestController) {
	physicalTests := r.Group("/club/physical-tests")
	physicalTests.Use(middleware.AuthMiddleware(), middleware.ClubRoleMiddleware())
	{
		// 体测活动管理
		physicalTests.GET("", ptController.GetPhysicalTests)
		physicalTests.POST("", ptController.CreatePhysicalTest)
		physicalTests.GET("/:id", ptController.GetPhysicalTest)
		physicalTests.GET("/:id/training-suggestions", ptController.GetTrainingSuggestions)
		physicalTests.PUT("/:id", ptController.UpdatePhysicalTest)
		physicalTests.DELETE("/:id", ptController.DeletePhysicalTest)

		// 发送通知
		physicalTests.POST("/:id/notify", ptController.NotifyPhysicalTest)

		// 体测数据管理
		physicalTests.GET("/:id/records", ptController.GetPhysicalTestRecords)
		physicalTests.POST("/:id/records", ptController.CreatePhysicalTestRecord)
		physicalTests.POST("/:id/records/batch", ptController.BatchImportPhysicalTestRecords)

		// 报告生成
		physicalTests.POST("/:id/generate-reports", ptController.GeneratePhysicalTestReports)

		// 自定义模板管理
		physicalTests.GET("/templates", ptController.GetCustomTemplates)
		physicalTests.POST("/templates", ptController.CreateCustomTemplate)
		physicalTests.DELETE("/templates/:id", ptController.DeleteCustomTemplate)
	}
}
