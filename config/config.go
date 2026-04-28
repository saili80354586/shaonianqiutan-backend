package config

import (
	"log"
	"os"

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
	return url
}

// GetFrontendURL 获取前端URL（用于CORS配置）
func GetFrontendURL() string {
	url := os.Getenv("FRONTEND_URL")
	if url == "" {
		return "http://localhost:5173"
	}
	return url
}

// GetCORSOrigins 获取允许的CORS源列表
// 开发模式允许localhost，生产环境只允许配置的FRONTEND_URL
func GetCORSOrigins() []string {
	if IsDevMode() {
		return []string{
			"http://localhost:5173",
			"http://127.0.0.1:5173",
			"http://localhost:3000",
			GetFrontendURL(),
		}
	}
	// 生产环境只允许配置的域名
	return []string{GetFrontendURL()}
}
