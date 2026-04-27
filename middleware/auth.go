package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/shaonianqiutan/backend/config"
	"github.com/shaonianqiutan/backend/models"
)

// Claims JWT声明
type Claims struct {
	UserID uint   `json:"userId"`
	Phone  string `json:"phone"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT令牌
func GenerateToken(userID uint, phone string) (string, error) {
	expiresIn, _ := time.ParseDuration(config.GetJWTExpiresIn())
	claims := Claims{
		UserID: userID,
		Phone:  phone,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.GetJWTSecret()))
}

// AuthMiddleware JWT认证中间件（支持 header Authorization 或 query token 参数）
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenStr string

		// 优先从 Authorization header 获取
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenStr = parts[1]
			}
		}

		// fallback: 从 query token 参数获取（用于 window.open 打开的下载链接）
		if tokenStr == "" {
			tokenStr = c.Query("token")
		}

		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证令牌"})
			c.Abort()
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.GetJWTSecret()), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证令牌无效或已过期"})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("userId", claims.UserID)
		c.Set("phone", claims.Phone)
		c.Next()
	}
}

// AnalystRoleMiddleware 分析师角色权限中间件
func AnalystRoleMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userId")
		if !exists || userID.(uint) == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
			c.Abort()
			return
		}

		// 查询数据库验证是否为分析师
		userRepo := models.NewUserRepository(config.GetDB())
		analyst, err := userRepo.FindAnalystByUserID(userID.(uint))
		if err != nil || analyst == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "您没有分析师权限"})
			c.Abort()
			return
		}

		// 将分析师ID存入上下文，方便后续控制器使用
		c.Set("analystId", analyst.ID)
		c.Next()
	}
}

// GetUserID 从上下文中获取用户ID
func GetUserID(c *gin.Context) uint {
	userID, exists := c.Get("userId")
	if !exists {
		return 0
	}
	return userID.(uint)
}
