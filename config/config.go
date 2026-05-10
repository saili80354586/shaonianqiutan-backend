package config

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

const (
	PaymentModeMock = "mock"
	PaymentModeReal = "real"
	SmsModeMock     = "mock"
	SmsModeReal     = "real"
)

// LoadEnv 加载环境变量
func LoadEnv() {
	candidates := []string{".env"}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, ".env"))
	}
	if _, file, _, ok := runtime.Caller(0); ok {
		projectRoot := filepath.Dir(filepath.Dir(file))
		candidates = append(candidates, filepath.Join(projectRoot, ".env"))
	}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), ".env"))
	}

	seen := make(map[string]bool, len(candidates))
	existing := make([]string, 0, len(candidates))
	for _, path := range candidates {
		abs, err := filepath.Abs(path)
		if err != nil || seen[abs] {
			continue
		}
		seen[abs] = true
		if _, statErr := os.Stat(abs); statErr == nil {
			existing = append(existing, abs)
		}
	}

	if len(existing) == 0 {
		log.Println("Warning: No .env file found, using system environment variables")
		return
	}
	if err := godotenv.Load(existing...); err != nil {
		log.Printf("Warning: failed to load .env files: %v", err)
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

func normalizePaymentMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", PaymentModeMock:
		return PaymentModeMock
	case PaymentModeReal:
		return PaymentModeReal
	default:
		return ""
	}
}

// GetPaymentMode 获取支付模式。未配置时默认 mock，便于本地和演示环境继续走通闭环。
func GetPaymentMode() string {
	mode := normalizePaymentMode(os.Getenv("PAYMENT_MODE"))
	if mode == "" {
		log.Printf("Warning: invalid PAYMENT_MODE=%q, falling back to %s", os.Getenv("PAYMENT_MODE"), PaymentModeMock)
		return PaymentModeMock
	}
	return mode
}

// IsMockPaymentMode 是否启用模拟支付模式
func IsMockPaymentMode() bool {
	return GetPaymentMode() == PaymentModeMock
}

func normalizeSmsMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case SmsModeMock:
		return SmsModeMock
	case SmsModeReal:
		return SmsModeReal
	default:
		return ""
	}
}

// GetSmsMode 获取短信模式。本地开发默认 mock；非开发环境必须通过 SMS_MODE 显式选择。
func GetSmsMode() string {
	rawMode := strings.TrimSpace(os.Getenv("SMS_MODE"))
	if rawMode == "" {
		if configIsDevelopment() {
			return SmsModeMock
		}
		return SmsModeReal
	}

	mode := normalizeSmsMode(rawMode)
	if mode == "" {
		log.Printf("Warning: invalid SMS_MODE=%q, falling back to %s", os.Getenv("SMS_MODE"), SmsModeMock)
		return SmsModeMock
	}
	return mode
}

func IsMockSmsMode() bool {
	return IsDevMode() || GetSmsMode() == SmsModeMock
}

func IsAnalystRegistrationAutoApproved() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv("ANALYST_REGISTRATION_AUTO_APPROVE")))
	if value == "" {
		return configIsDevelopment()
	}
	return value == "1" || value == "true" || value == "yes"
}

func IsAnalystDefaultDemoOrderEnabled() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv("ANALYST_DEFAULT_DEMO_ORDER_ENABLED")))
	return value == "1" || value == "true" || value == "yes"
}

func GetAnalystDefaultDemoOrderTemplateOrderID() (uint, error) {
	raw := strings.TrimSpace(os.Getenv("ANALYST_DEFAULT_DEMO_ORDER_TEMPLATE_ORDER_ID"))
	if raw == "" {
		return 0, nil
	}

	parsed, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || parsed == 0 {
		return 0, strconv.ErrSyntax
	}
	return uint(parsed), nil
}

func configIsDevelopment() bool {
	return os.Getenv("NODE_ENV") == "development"
}

// ValidateRuntimeConfig 校验运行环境，防止漏配时误按开发模式启动
func ValidateRuntimeConfig() {
	if rawMode := os.Getenv("PAYMENT_MODE"); strings.TrimSpace(rawMode) != "" && normalizePaymentMode(rawMode) == "" {
		log.Fatalf("FATAL: PAYMENT_MODE must be one of %q or %q, got %q", PaymentModeMock, PaymentModeReal, rawMode)
	}
	if rawMode := os.Getenv("SMS_MODE"); strings.TrimSpace(rawMode) != "" && normalizeSmsMode(rawMode) == "" {
		log.Fatalf("FATAL: SMS_MODE must be one of %q or %q, got %q", SmsModeMock, SmsModeReal, rawMode)
	}

	if !IsDevMode() && strings.TrimSpace(os.Getenv("PAYMENT_MODE")) == "" {
		log.Fatalf("FATAL: PAYMENT_MODE environment variable must be set to %q or %q outside development", PaymentModeMock, PaymentModeReal)
	}
	if !IsDevMode() && strings.TrimSpace(os.Getenv("SMS_MODE")) == "" {
		log.Fatalf("FATAL: SMS_MODE environment variable must be set to %q or %q outside development", SmsModeMock, SmsModeReal)
	}
	if !IsDevMode() && strings.TrimSpace(os.Getenv("ANALYST_REGISTRATION_AUTO_APPROVE")) == "" {
		log.Fatalf("FATAL: ANALYST_REGISTRATION_AUTO_APPROVE must be explicitly set to true or false outside development")
	}
	if IsAnalystDefaultDemoOrderEnabled() {
		templateOrderID, err := GetAnalystDefaultDemoOrderTemplateOrderID()
		if err != nil || templateOrderID == 0 {
			log.Fatalf("FATAL: ANALYST_DEFAULT_DEMO_ORDER_TEMPLATE_ORDER_ID must be a positive integer when ANALYST_DEFAULT_DEMO_ORDER_ENABLED=true")
		}
	}

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
