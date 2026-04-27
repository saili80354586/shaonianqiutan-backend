package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/config"
	"github.com/shaonianqiutan/backend/models"
)

// TeamAccessContext 球队访问上下文
type TeamAccessContext struct {
	IsClubAdmin  bool   // 是否是俱乐部管理员
	IsTeamCoach  bool   // 是否是球队教练
	CoachRole    string // 教练角色（如果是教练）
	ClubID       uint   // 俱乐部ID
	TeamID       uint   // 球队ID
}

// TeamAccessMiddleware 球队访问权限中间件
// 检查用户是否为俱乐部管理员或球队教练，任一身份即可访问
func TeamAccessMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			c.Abort()
			return
		}

		// 从URL参数获取球队ID（支持 teamId 和 id 两种参数名）
		teamIDStr := c.Param("teamId")
		if teamIDStr == "" {
			teamIDStr = c.Param("id")
		}
		if teamIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "缺少球队ID"})
			c.Abort()
			return
		}

		teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "球队ID格式错误"})
			c.Abort()
			return
		}

		db := config.GetDB()
		accessCtx := &TeamAccessContext{
			TeamID: uint(teamID),
		}

		// 1. 检查用户是否为俱乐部管理员
		var club models.Club
		if err := db.Where("user_id = ?", userID).First(&club).Error; err == nil {
			// 用户是某个俱乐部的管理员，检查该球队是否属于此俱乐部
			// teams.club_id 存的是 clubs.id
			var team models.Team
			if err := db.Where("id = ? AND club_id = ?", teamID, club.ID).First(&team).Error; err == nil {
				accessCtx.IsClubAdmin = true
				accessCtx.ClubID = club.ID
			}
		}

		// 2. 检查用户是否为该球队的教练
		var teamCoach models.TeamCoach
		if err := db.Where("team_id = ? AND user_id = ? AND status = ?", teamID, userID, "active").First(&teamCoach).Error; err == nil {
			accessCtx.IsTeamCoach = true
			accessCtx.CoachRole = string(teamCoach.Role)
			// 获取球队所属的俱乐部ID
			var team models.Team
			if err := db.Where("id = ?", teamID).First(&team).Error; err == nil {
				accessCtx.ClubID = team.ClubID
			}
		}

		// 3. 检查权限：必须是俱乐部管理员或球队教练
		if !accessCtx.IsClubAdmin && !accessCtx.IsTeamCoach {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权访问该球队"})
			c.Abort()
			return
		}

		// 将访问上下文存入 gin context
		c.Set("teamAccess", accessCtx)
		c.Next()
	}
}

// GetTeamAccessContext 从 gin context 获取球队访问上下文
func GetTeamAccessContext(c *gin.Context) *TeamAccessContext {
	accessCtx, exists := c.Get("teamAccess")
	if !exists {
		return nil
	}
	return accessCtx.(*TeamAccessContext)
}

// RequireClubAdmin 要求必须是俱乐部管理员的中间件
// 用于只有俱乐部管理员才能执行的操作（如添加/移除教练）
func RequireClubAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessCtx := GetTeamAccessContext(c)
		if accessCtx == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权访问"})
			c.Abort()
			return
		}

		if !accessCtx.IsClubAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "仅俱乐部管理员可操作"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireHeadCoach 要求必须是主教练或俱乐部管理员的中间件
// 用于主教练特有的操作（如发起周报、审核比赛总结）
func RequireHeadCoach() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessCtx := GetTeamAccessContext(c)
		if accessCtx == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权访问"})
			c.Abort()
			return
		}

		// 俱乐部管理员可以操作
		if accessCtx.IsClubAdmin {
			c.Next()
			return
		}

		// 主教练可以操作
		if accessCtx.IsTeamCoach && accessCtx.CoachRole == string(models.CoachRoleHead) {
			c.Next()
			return
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "仅主教练或俱乐部管理员可操作"})
		c.Abort()
	}
}

// RequireCoach 要求必须是教练（任何角色）或俱乐部管理员的中间件
func RequireCoach() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessCtx := GetTeamAccessContext(c)
		if accessCtx == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "无权访问"})
			c.Abort()
			return
		}

		// 俱乐部管理员或任何教练角色都可以
		if accessCtx.IsClubAdmin || accessCtx.IsTeamCoach {
			c.Next()
			return
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "无权操作"})
		c.Abort()
	}
}
