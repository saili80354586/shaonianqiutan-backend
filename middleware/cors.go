package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/config"
)

// CorsMiddleware 配置CORS中间件
// ⚠️ 注意：当前项目的CORS已在 main.go 中统一配置，此中间件仅作为备用
// 如需使用，请确保仅在开发环境启用，且绝不设置 AllowAllOrigins + AllowCredentials 同时启用
func CorsMiddleware() gin.HandlerFunc {
	cfg := cors.DefaultConfig()
	// 生产环境：只允许配置的前端域名
	// 开发环境：允许 localhost
	if config.IsDevMode() {
		cfg.AllowOrigins = config.GetCORSOrigins()
	} else {
		cfg.AllowOrigins = []string{config.GetFrontendURL()}
	}
	cfg.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}
	cfg.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"}
	cfg.ExposeHeaders = []string{"Content-Length"}
	cfg.AllowCredentials = true
	return cors.New(cfg)
}
