package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/controllers"
	"github.com/shaonianqiutan/backend/middleware"
)

func SetupAccountRoleRoutes(r *gin.RouterGroup, accountRoleController *controllers.AccountRoleController) {
	account := r.Group("/account")
	account.Use(middleware.AuthMiddleware())
	{
		account.GET("/roles", accountRoleController.ListRoles)
		account.POST("/roles/apply", accountRoleController.ApplyRole)
	}

	admin := r.Group("/admin/role-applications")
	admin.Use(middleware.AuthMiddleware(), middleware.AdminRoleMiddleware())
	{
		admin.GET("", accountRoleController.ListAdminApplications)
		admin.POST("/:id/review", accountRoleController.ReviewRoleApplication)
	}
}
