package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/config"
	"github.com/shaonianqiutan/backend/models"
)

func MaintenanceModeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions || !strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Next()
			return
		}

		settings := models.LoadAdminSystemSettings(config.GetDB())
		if !settings.MaintenanceMode || isMaintenanceBypassPath(c.Request.URL.Path) || hasAdminBearerToken(c) {
			c.Next()
			return
		}

		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "MAINTENANCE_MODE",
				"message": "平台维护中，请稍后再试",
			},
		})
		c.Abort()
	}
}

func isMaintenanceBypassPath(path string) bool {
	return path == "/api/system/public-settings" ||
		path == "/api/admin/login" ||
		strings.HasPrefix(path, "/api/admin/settings")
}

func hasAdminBearerToken(c *gin.Context) bool {
	authHeader := c.GetHeader("Authorization")
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return false
	}

	claims, err := ParseToken(parts[1])
	if err != nil || claims == nil || claims.UserID == 0 {
		return false
	}

	userRepo := models.NewUserRepository(config.GetDB())
	user, err := userRepo.FindByID(claims.UserID)
	return err == nil && user != nil && user.Status == models.StatusActive && user.Role == models.RoleAdmin
}
