package controllers

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/services"
	"github.com/shaonianqiutan/backend/utils"
)

// TrainingPlanController 训练计划控制器
type TrainingPlanController struct {
	clubService *services.ClubService
	db          *gorm.DB
}

// NewTrainingPlanController 创建训练计划控制器
func NewTrainingPlanController(clubService *services.ClubService, db *gorm.DB) *TrainingPlanController {
	return &TrainingPlanController{clubService: clubService, db: db}
}

// ListTrainingPlans 获取训练计划列表
func (c *TrainingPlanController) ListTrainingPlans(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.SuccessResponse(ctx, []interface{}{})
		return
	}

	teamID, _ := strconv.ParseUint(ctx.Query("teamId"), 10, 32)
	month := ctx.Query("month") // 格式: 2026-04

	var plans []models.TrainingPlan
	query := c.db.Where("club_id = ?", club.ID)
	if teamID > 0 {
		query = query.Where("team_id = ?", teamID)
	}
	if month != "" {
		start, _ := time.Parse("2006-01", month)
		end := start.AddDate(0, 1, 0)
		query = query.Where("start_time >= ? AND start_time < ?", start, end)
	}
	query.Order("start_time DESC").Find(&plans)

	list := make([]gin.H, 0, len(plans))
	for _, p := range plans {
		list = append(list, formatTrainingPlan(&p))
	}
	utils.SuccessResponse(ctx, list)
}

// CreateTrainingPlan 创建训练计划
func (c *TrainingPlanController) CreateTrainingPlan(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	var req struct {
		TeamID         uint     `json:"teamId" binding:"required"`
		Title          string   `json:"title" binding:"required"`
		Theme          string   `json:"theme"`
		Location       string   `json:"location"`
		StartTime      string   `json:"startTime" binding:"required"`
		EndTime        *string  `json:"endTime"`
		PlayerIDs      []uint   `json:"playerIds"`
		Content        string   `json:"content"`
		VideoURLs      []string `json:"videoUrls"`
		WeeklyReportID *uint    `json:"weeklyReportId"`
		PhysicalTestID *uint    `json:"physicalTestId"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	startTime, _ := time.Parse(time.RFC3339, req.StartTime)
	var endTime *time.Time
	if req.EndTime != nil {
		et, _ := time.Parse(time.RFC3339, *req.EndTime)
		endTime = &et
	}

	plan := models.TrainingPlan{
		ClubID:         club.ID,
		TeamID:         req.TeamID,
		Title:          req.Title,
		Theme:          req.Theme,
		Location:       req.Location,
		StartTime:      startTime,
		EndTime:        endTime,
		Content:        req.Content,
		CoachID:        userID,
		WeeklyReportID: req.WeeklyReportID,
		PhysicalTestID: req.PhysicalTestID,
		Status:         models.TrainingPlanStatusDraft,
		CreatedBy:      userID,
	}
	if len(req.PlayerIDs) > 0 {
		idsJSON, _ := json.Marshal(req.PlayerIDs)
		plan.PlayerIDs = string(idsJSON)
	}
	if len(req.VideoURLs) > 0 {
		urlsJSON, _ := json.Marshal(req.VideoURLs)
		plan.VideoURLs = string(urlsJSON)
	}

	if err := c.db.Create(&plan).Error; err != nil {
		utils.ServerError(ctx, "创建失败")
		return
	}
	utils.SuccessResponse(ctx, formatTrainingPlan(&plan))
}

// GetTrainingPlan 获取训练计划详情
func (c *TrainingPlanController) GetTrainingPlan(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	var plan models.TrainingPlan
	if err := c.db.First(&plan, id).Error; err != nil || plan.ClubID != club.ID {
		utils.NotFoundError(ctx, "训练计划不存在")
		return
	}
	utils.SuccessResponse(ctx, formatTrainingPlan(&plan))
}

// UpdateTrainingPlan 更新训练计划
func (c *TrainingPlanController) UpdateTrainingPlan(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	var plan models.TrainingPlan
	if err := c.db.First(&plan, id).Error; err != nil || plan.ClubID != club.ID {
		utils.NotFoundError(ctx, "训练计划不存在")
		return
	}

	var req struct {
		Title          string   `json:"title"`
		Theme          string   `json:"theme"`
		Location       string   `json:"location"`
		StartTime      string   `json:"startTime"`
		EndTime        *string  `json:"endTime"`
		PlayerIDs      []uint   `json:"playerIds"`
		Content        string   `json:"content"`
		VideoURLs      []string `json:"videoUrls"`
		Summary        string   `json:"summary"`
		Status         string   `json:"status"`
		WeeklyReportID *uint    `json:"weeklyReportId"`
		PhysicalTestID *uint    `json:"physicalTestId"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	updates := make(map[string]interface{})
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Theme != "" {
		updates["theme"] = req.Theme
	}
	if req.Location != "" {
		updates["location"] = req.Location
	}
	if req.StartTime != "" {
		st, _ := time.Parse(time.RFC3339, req.StartTime)
		updates["start_time"] = st
	}
	if req.EndTime != nil {
		et, _ := time.Parse(time.RFC3339, *req.EndTime)
		updates["end_time"] = et
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.Summary != "" {
		updates["summary"] = req.Summary
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.PlayerIDs != nil {
		idsJSON, _ := json.Marshal(req.PlayerIDs)
		updates["player_ids"] = string(idsJSON)
	}
	if req.VideoURLs != nil {
		urlsJSON, _ := json.Marshal(req.VideoURLs)
		updates["video_urls"] = string(urlsJSON)
	}
	if req.WeeklyReportID != nil {
		updates["weekly_report_id"] = *req.WeeklyReportID
	} else if ctx.Request.Body != nil {
		updates["weekly_report_id"] = nil
	}
	if req.PhysicalTestID != nil {
		updates["physical_test_id"] = *req.PhysicalTestID
	} else if ctx.Request.Body != nil {
		updates["physical_test_id"] = nil
	}

	c.db.Model(&plan).Updates(updates)
	c.db.First(&plan, plan.ID)
	utils.SuccessResponse(ctx, formatTrainingPlan(&plan))
}

// DeleteTrainingPlan 删除训练计划
func (c *TrainingPlanController) DeleteTrainingPlan(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	var plan models.TrainingPlan
	if err := c.db.First(&plan, id).Error; err != nil || plan.ClubID != club.ID {
		utils.NotFoundError(ctx, "训练计划不存在")
		return
	}

	c.db.Delete(&plan)
	utils.SuccessResponseWithMessage(ctx, nil, "删除成功")
}

func formatTrainingPlan(p *models.TrainingPlan) gin.H {
	return gin.H{
		"id":             p.ID,
		"clubId":         p.ClubID,
		"teamId":         p.TeamID,
		"title":          p.Title,
		"theme":          p.Theme,
		"location":       p.Location,
		"startTime":      p.StartTime.Format(time.RFC3339),
		"endTime":        utils.FormatTime(p.EndTime),
		"playerIds":      p.GetPlayerIDs(),
		"content":        p.Content,
		"videoUrls":      p.GetVideoURLs(),
		"summary":        p.Summary,
		"coachId":        p.CoachID,
		"weeklyReportId": p.WeeklyReportID,
		"physicalTestId": p.PhysicalTestID,
		"status":         p.Status,
		"createdBy":      p.CreatedBy,
		"createdAt":      p.CreatedAt.Format(time.RFC3339),
	}
}
