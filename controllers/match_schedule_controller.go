package controllers

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/services"
	"github.com/shaonianqiutan/backend/utils"
)

// MatchScheduleController 赛程日历控制器
type MatchScheduleController struct {
	clubService *services.ClubService
	db          *gorm.DB
}

// NewMatchScheduleController 创建赛程日历控制器
func NewMatchScheduleController(clubService *services.ClubService, db *gorm.DB) *MatchScheduleController {
	return &MatchScheduleController{clubService: clubService, db: db}
}

// ListMatchSchedules 获取赛程列表
func (c *MatchScheduleController) ListMatchSchedules(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.SuccessResponse(ctx, []interface{}{})
		return
	}

	teamID, _ := strconv.ParseUint(ctx.Query("teamId"), 10, 32)
	month := ctx.Query("month") // 格式: 2026-04

	var schedules []models.MatchSchedule
	query := c.db.Where("club_id = ?", club.ID)
	if teamID > 0 {
		query = query.Where("team_id = ?", teamID)
	}
	if month != "" {
		start, _ := time.Parse("2006-01", month)
		end := start.AddDate(0, 1, 0)
		query = query.Where("match_time >= ? AND match_time < ?", start, end)
	}
	query.Order("match_time ASC").Find(&schedules)

	list := make([]gin.H, 0, len(schedules))
	for _, s := range schedules {
		list = append(list, formatMatchSchedule(&s))
	}
	utils.SuccessResponse(ctx, list)
}

// CreateMatchSchedule 创建赛程
func (c *MatchScheduleController) CreateMatchSchedule(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	var req struct {
		TeamID    string `json:"teamId" binding:"required"`
		Name      string `json:"name" binding:"required"`
		MatchType string `json:"matchType" binding:"required"`
		Opponent  string `json:"opponent"`
		MatchTime string `json:"matchTime" binding:"required"`
		Location  string `json:"location"`
		Remark    string `json:"remark"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	teamID, _ := strconv.ParseUint(req.TeamID, 10, 32)
	matchTime, _ := time.Parse(time.RFC3339, req.MatchTime)

	schedule := models.MatchSchedule{
		ClubID:    club.ID,
		TeamID:    uint(teamID),
		Name:      req.Name,
		MatchType: models.MatchScheduleType(req.MatchType),
		Opponent:  req.Opponent,
		MatchTime: matchTime,
		Location:  req.Location,
		Remark:    req.Remark,
		Status:    models.MatchScheduleStatusUpcoming,
		CreatedBy: userID,
	}

	if err := c.db.Create(&schedule).Error; err != nil {
		utils.ServerError(ctx, "创建失败")
		return
	}
	utils.SuccessResponse(ctx, formatMatchSchedule(&schedule))
}

// GetMatchSchedule 获取赛程详情
func (c *MatchScheduleController) GetMatchSchedule(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	var schedule models.MatchSchedule
	if err := c.db.First(&schedule, id).Error; err != nil || schedule.ClubID != club.ID {
		utils.NotFoundError(ctx, "赛程不存在")
		return
	}
	utils.SuccessResponse(ctx, formatMatchSchedule(&schedule))
}

// UpdateMatchSchedule 更新赛程
func (c *MatchScheduleController) UpdateMatchSchedule(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	var schedule models.MatchSchedule
	if err := c.db.First(&schedule, id).Error; err != nil || schedule.ClubID != club.ID {
		utils.NotFoundError(ctx, "赛程不存在")
		return
	}

	var req struct {
		Name      string `json:"name"`
		MatchType string `json:"matchType"`
		Opponent  string `json:"opponent"`
		MatchTime string `json:"matchTime"`
		Location  string `json:"location"`
		Remark    string `json:"remark"`
		Status    string `json:"status"`
		HomeScore *int   `json:"homeScore"`
		AwayScore *int   `json:"awayScore"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.MatchType != "" {
		updates["match_type"] = req.MatchType
	}
	if req.Opponent != "" {
		updates["opponent"] = req.Opponent
	}
	if req.MatchTime != "" {
		mt, _ := time.Parse(time.RFC3339, req.MatchTime)
		updates["match_time"] = mt
	}
	if req.Location != "" {
		updates["location"] = req.Location
	}
	if req.Remark != "" {
		updates["remark"] = req.Remark
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.HomeScore != nil {
		updates["home_score"] = req.HomeScore
	}
	if req.AwayScore != nil {
		updates["away_score"] = req.AwayScore
	}

	c.db.Model(&schedule).Updates(updates)
	c.db.First(&schedule, schedule.ID)
	utils.SuccessResponse(ctx, formatMatchSchedule(&schedule))
}

// DeleteMatchSchedule 删除赛程
func (c *MatchScheduleController) DeleteMatchSchedule(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	var schedule models.MatchSchedule
	if err := c.db.First(&schedule, id).Error; err != nil || schedule.ClubID != club.ID {
		utils.NotFoundError(ctx, "赛程不存在")
		return
	}

	c.db.Delete(&schedule)
	utils.SuccessResponseWithMessage(ctx, nil, "删除成功")
}

func formatMatchSchedule(s *models.MatchSchedule) gin.H {
	return gin.H{
		"id":             s.ID,
		"clubId":         s.ClubID,
		"teamId":         s.TeamID,
		"name":           s.Name,
		"matchType":      s.MatchType,
		"opponent":       s.Opponent,
		"matchTime":      s.MatchTime.Format(time.RFC3339),
		"location":       s.Location,
		"homeScore":      s.HomeScore,
		"awayScore":      s.AwayScore,
		"remark":         s.Remark,
		"status":         s.Status,
		"matchSummaryId": s.MatchSummaryID,
		"createdBy":      s.CreatedBy,
		"createdAt":      s.CreatedAt.Format(time.RFC3339),
	}
}
