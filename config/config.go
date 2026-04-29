package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// LoadEnv 加载环境变量
func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found, using system environment variables")
	}
}

// GetJWTSecret 获取JWT密钥
func GetJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("FATAL: JWT_SECRET environment variable must be set in production")
	}
	return secret
}

// GetJWTExpiresIn 获取JWT过期时间
func GetJWTExpiresIn() string {
	expiresIn := os.Getenv("JWT_EXPIRES_IN")
	if expiresIn == "" {
		expiresIn = "168h" // 7天
	}
	return expiresIn
}

// GetPort 获取服务端口
func GetPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return ":" + port
}

// IsDevMode 是否是开发模式
func IsDevMode() bool {
	mode := os.Getenv("NODE_ENV")
	return mode == "development"
}

// ValidateRuntimeConfig 校验运行环境，防止漏配时误按开发模式启动
func ValidateRuntimeConfig() {
	if IsDevMode() {
		return
	}

	required := []string{"JWT_SECRET", "FRONTEND_URL", "BASE_URL"}
	for _, key := range required {
		if os.Getenv(key) == "" {
			log.Fatalf("FATAL: %s environment variable must be set outside development", key)
		}
	}
}

// GetBaseUrl 获取基础URL
func GetBaseUrl() string {
	url := os.Getenv("BASE_URL")
	if url == "" {
		return "http://localhost:8080"
	}
	return normalizeOrigin(url)
}

// GetFrontendURL 获取前端URL（用于CORS配置）
func GetFrontendURL() string {
	url := os.Getenv("FRONTEND_URL")
	if url == "" {
		return "http://localhost:5173"
	}
	return normalizeOrigin(url)
}

// GetCORSOrigins 获取允许的CORS源列表
// CORS_ALLOWED_ORIGINS 支持逗号分隔的多个前端源；FRONTEND_URL 保持向后兼容。
// 开发模式额外允许常用本地前端源，生产环境只允许显式配置的源。
func GetCORSOrigins() []string {
	origins := make([]string, 0, 4)
	origins = appendOrigins(origins, os.Getenv("CORS_ALLOWED_ORIGINS"))
	origins = appendUniqueOrigin(origins, GetFrontendURL())

	if IsDevMode() {
		origins = appendUniqueOrigin(origins, "http://localhost:5173")
		origins = appendUniqueOrigin(origins, "http://127.0.0.1:5173")
		origins = appendUniqueOrigin(origins, "http://localhost:3000")
	}

	return origins
}

func appendOrigins(origins []string, csv string) []string {
	for _, origin := range strings.Split(csv, ",") {
		origins = appendUniqueOrigin(origins, origin)
	}
	return origins
}

func appendUniqueOrigin(origins []string, origin string) []string {
	normalized := normalizeOrigin(origin)
	if normalized == "" {
		return origins
	}
	for _, existing := range origins {
		if existing == normalized {
			return origins
		}
	}
	return append(origins, normalized)
}

func normalizeOrigin(origin string) string {
	return strings.TrimRight(strings.TrimSpace(origin), "/")
}
