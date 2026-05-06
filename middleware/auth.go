package middleware

import (
	"errors"
	"net/http"
	"strconv"
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

// ParseTokenAllowExpired 校验签名并解析 Token，但不校验过期时间；仅用于刷新令牌。
func ParseTokenAllowExpired(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, err := parser.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GetJWTSecret()), nil
	})
	if err != nil || token == nil || !token.Valid {
		return nil, errors.New("认证令牌无效")
	}
	return claims, nil
}

func ParseToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GetJWTSecret()), nil
	})
	if err != nil || token == nil || !token.Valid {
		return nil, errors.New("认证令牌无效或已过期")
	}
	return claims, nil
}

// AuthMiddleware JWT认证中间件（仅支持 header Authorization）
func AuthMiddleware() gin.HandlerFunc {
	return authMiddleware(false)
}

// OptionalAuthMiddleware 有 token 时写入用户上下文，无 token 时允许继续访问公开接口。
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证令牌格式错误"})
			c.Abort()
			return
		}

		claims, err := ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证令牌无效或已过期"})
			c.Abort()
			return
		}

		userRepo := models.NewUserRepository(config.GetDB())
		user, err := userRepo.FindByID(claims.UserID)
		if err != nil || user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在或已被删除"})
			c.Abort()
			return
		}
		if user.Status != models.StatusActive {
			c.JSON(http.StatusForbidden, gin.H{"error": "账号未激活或已被禁用"})
			c.Abort()
			return
		}

		c.Set("userId", claims.UserID)
		c.Set("phone", claims.Phone)
		c.Set("user", user)
		c.Next()
	}
}

// QueryTokenAuthMiddleware JWT认证中间件（兼容 WebSocket query token）
func QueryTokenAuthMiddleware() gin.HandlerFunc {
	return authMiddleware(true)
}

func authMiddleware(allowQueryToken bool) gin.HandlerFunc {
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

		if allowQueryToken && tokenStr == "" {
			tokenStr = c.Query("token")
		}

		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证令牌"})
			c.Abort()
			return
		}

		claims, err := ParseToken(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证令牌无效或已过期"})
			c.Abort()
			return
		}

		userRepo := models.NewUserRepository(config.GetDB())
		user, err := userRepo.FindByID(claims.UserID)
		if err != nil || user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在或已被删除"})
			c.Abort()
			return
		}
		if user.Status != models.StatusActive {
			c.JSON(http.StatusForbidden, gin.H{"error": "账号未激活或已被禁用"})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("userId", claims.UserID)
		c.Set("phone", claims.Phone)
		c.Set("user", user)
		c.Next()
	}
}

// AdminRoleMiddleware 管理员角色权限中间件
func AdminRoleMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userValue, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
			c.Abort()
			return
		}

		user, ok := userValue.(*models.User)
		if !ok || user == nil || user.Status != models.StatusActive || user.Role != models.RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "您没有管理员权限"})
			c.Abort()
			return
		}

		c.Set("adminId", user.ID)
		c.Next()
	}
}

// ScoutRoleMiddleware 球探角色权限中间件
func ScoutRoleMiddleware() gin.HandlerFunc {
	return requireActiveRole(models.RoleScout, "您没有球探权限")
}

// CoachRoleMiddleware 教练角色权限中间件
func CoachRoleMiddleware() gin.HandlerFunc {
	return requireActiveRole(models.RoleCoach, "您没有教练权限")
}

// ClubRoleMiddleware 要求当前登录用户必须管理一个俱乐部。
func ClubRoleMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		club, ok := getAuthenticatedClub(c)
		if !ok {
			return
		}

		c.Set("clubId", club.ID)
		c.Set("club", club)
		c.Next()
	}
}

// ClubOwnerMiddleware 要求当前登录用户必须是 URL 参数中指定俱乐部的管理员。
func ClubOwnerMiddleware(paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		clubIDStr := c.Param(paramName)
		if clubIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "缺少俱乐部ID"})
			c.Abort()
			return
		}

		clubID, err := strconv.ParseUint(clubIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "俱乐部ID格式错误"})
			c.Abort()
			return
		}

		club, ok := getAuthenticatedClub(c)
		if !ok {
			return
		}
		if club.ID != uint(clubID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权管理该俱乐部"})
			c.Abort()
			return
		}

		c.Set("clubId", club.ID)
		c.Set("club", club)
		c.Next()
	}
}

func getAuthenticatedClub(c *gin.Context) (*models.Club, bool) {
	userID := GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		c.Abort()
		return nil, false
	}

	userValue, exists := c.Get("user")
	user, ok := userValue.(*models.User)
	if !exists || !ok || user == nil || user.Status != models.StatusActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "您没有俱乐部权限"})
		c.Abort()
		return nil, false
	}

	var club models.Club
	if err := config.GetDB().Where("user_id = ?", userID).First(&club).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "您没有俱乐部权限"})
		c.Abort()
		return nil, false
	}

	return &club, true
}

func requireActiveRole(role models.UserRole, message string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userValue, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
			c.Abort()
			return
		}

		user, ok := userValue.(*models.User)
		if !ok || user == nil || user.Status != models.StatusActive {
			c.JSON(http.StatusForbidden, gin.H{"error": message})
			c.Abort()
			return
		}

		hasRole, err := userHasActiveRole(user, role)
		if err != nil || !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": message})
			c.Abort()
			return
		}

		c.Next()
	}
}

func userHasActiveRole(user *models.User, role models.UserRole) (bool, error) {
	if user.Role == role {
		return true, nil
	}

	db := config.GetDB()
	var count int64
	switch role {
	case models.RoleScout:
		if err := db.Model(&models.Scout{}).Where("user_id = ?", user.ID).Count(&count).Error; err != nil {
			return false, err
		}
		return count > 0, nil
	case models.RoleCoach:
		if err := db.Model(&models.ClubCoach{}).Where("user_id = ? AND status = ?", user.ID, models.ClubCoachStatusActive).Count(&count).Error; err != nil {
			return false, err
		}
		if count > 0 {
			return true, nil
		}

		count = 0
		if err := db.Model(&models.TeamCoach{}).Where("user_id = ? AND status = ?", user.ID, "active").Count(&count).Error; err != nil {
			return false, err
		}
		return count > 0, nil
	default:
		return false, nil
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
		if err != nil || analyst == nil || analyst.Status != models.AnalystStatusActive {
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
