package controllers

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/services"
	"github.com/shaonianqiutan/backend/utils"
	"gorm.io/gorm"
)

var (
	errInvalidTrainingAttendanceStatus = errors.New("invalid training attendance status")
	errInvalidTrainingAttendancePlayer = errors.New("invalid training attendance player")
)

// TrainingPlanController 训练计划控制器
type TrainingPlanController struct {
	clubService         *services.ClubService
	db                  *gorm.DB
	notificationService *services.NotificationService
}

// NewTrainingPlanController 创建训练计划控制器
func NewTrainingPlanController(clubService *services.ClubService, db *gorm.DB) *TrainingPlanController {
	return &TrainingPlanController{clubService: clubService, db: db}
}

func (c *TrainingPlanController) SetNotificationService(notificationService *services.NotificationService) {
	c.notificationService = notificationService
}

type trainingAttendanceRequestItem struct {
	PlayerID uint   `json:"playerId" binding:"required"`
	Status   string `json:"status" binding:"required"`
	Remark   string `json:"remark"`
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
	c.notifyTrainingPlanCreated(&plan)
	utils.SuccessResponse(ctx, formatTrainingPlan(&plan))
}

func (c *TrainingPlanController) notifyTrainingPlanCreated(plan *models.TrainingPlan) {
	if c.notificationService == nil || plan == nil {
		return
	}

	playerIDs := plan.GetPlayerIDs()
	if len(playerIDs) == 0 {
		playerIDs = c.getActiveTeamPlayerIDs(plan.TeamID)
	}
	if len(playerIDs) == 0 {
		return
	}

	extra := map[string]interface{}{
		"team_id":    plan.TeamID,
		"start_time": plan.StartTime.Format(time.RFC3339),
		"location":   plan.Location,
	}
	err := c.notificationService.NotifyTeamCalendarEvent(
		playerIDs,
		models.NotificationTypeTrainingPlanCreated,
		"新的训练计划",
		plan.Title+" 已加入球队日历，请按时参加",
		"training_plan",
		plan.ID,
		"/user-dashboard?tab=team_calendar",
		extra,
	)
	if err != nil {
		log.Printf("发送训练计划通知失败 (planID=%d): %v", plan.ID, err)
		return
	}
	_ = c.db.Model(plan).Update("remind_sent", true).Error
}

func (c *TrainingPlanController) getActiveTeamPlayerIDs(teamID uint) []uint {
	var players []models.TeamPlayer
	if err := c.db.Where("team_id = ? AND status = ?", teamID, "active").Find(&players).Error; err != nil {
		return []uint{}
	}

	playerIDs := make([]uint, 0, len(players))
	for _, player := range players {
		playerIDs = append(playerIDs, player.UserID)
	}
	return playerIDs
}

func (c *TrainingPlanController) getClubTrainingPlan(ctx *gin.Context, clubID uint) (*models.TrainingPlan, bool) {
	id, _ := strconv.ParseUint(ctx.Param("id"), 10, 32)
	var plan models.TrainingPlan
	if err := c.db.First(&plan, id).Error; err != nil || plan.ClubID != clubID {
		utils.NotFoundError(ctx, "训练计划不存在")
		return nil, false
	}
	return &plan, true
}

func (c *TrainingPlanController) getTrainingAttendancePlayerSet(plan *models.TrainingPlan) (map[uint]bool, error) {
	plannedPlayers := make(map[uint]bool)
	for _, id := range plan.GetPlayerIDs() {
		plannedPlayers[id] = true
	}

	var teamPlayers []models.TeamPlayer
	if err := c.db.Where("team_id = ? AND status = ?", plan.TeamID, "active").Find(&teamPlayers).Error; err != nil {
		return nil, err
	}

	result := make(map[uint]bool, len(teamPlayers))
	for _, teamPlayer := range teamPlayers {
		if len(plannedPlayers) > 0 && !plannedPlayers[teamPlayer.UserID] {
			continue
		}
		result[teamPlayer.UserID] = true
	}
	return result, nil
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

// GetTrainingAttendance 获取训练出勤列表
func (c *TrainingPlanController) GetTrainingAttendance(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	plan, ok := c.getClubTrainingPlan(ctx, club.ID)
	if !ok {
		return
	}

	var teamPlayers []models.TeamPlayer
	if err := c.db.Preload("User").
		Where("team_id = ? AND status = ?", plan.TeamID, "active").
		Order("jersey_number ASC, id ASC").
		Find(&teamPlayers).Error; err != nil {
		utils.ServerError(ctx, "获取球队球员失败")
		return
	}

	var attendance []models.TrainingAttendance
	if err := c.db.Where("training_plan_id = ?", plan.ID).Find(&attendance).Error; err != nil {
		utils.ServerError(ctx, "获取出勤记录失败")
		return
	}
	attendanceMap := make(map[uint]models.TrainingAttendance, len(attendance))
	for _, item := range attendance {
		attendanceMap[item.PlayerID] = item
	}

	plannedPlayers := make(map[uint]bool)
	for _, id := range plan.GetPlayerIDs() {
		plannedPlayers[id] = true
	}
	usePlannedFilter := len(plannedPlayers) > 0

	list := make([]gin.H, 0, len(teamPlayers))
	summary := emptyAttendanceSummary()
	for _, teamPlayer := range teamPlayers {
		if usePlannedFilter && !plannedPlayers[teamPlayer.UserID] {
			continue
		}
		user := teamPlayer.User
		if user == nil {
			continue
		}
		record, hasRecord := attendanceMap[teamPlayer.UserID]
		status := "unmarked"
		remark := ""
		recordedAt := ""
		if hasRecord {
			status = string(record.Status)
			remark = record.Remark
			recordedAt = utils.FormatDateTime(record.RecordedAt)
			summary[status]++
		} else {
			summary["unmarked"]++
		}
		list = append(list, gin.H{
			"playerId":     teamPlayer.UserID,
			"teamPlayerId": teamPlayer.ID,
			"name":         user.Name,
			"nickname":     user.Nickname,
			"avatar":       user.Avatar,
			"position":     teamPlayer.Position,
			"jerseyNumber": teamPlayer.JerseyNumber,
			"status":       status,
			"remark":       remark,
			"recordedAt":   recordedAt,
		})
	}
	summary["total"] = len(list)

	utils.SuccessResponse(ctx, gin.H{
		"trainingPlanId": plan.ID,
		"teamId":         plan.TeamID,
		"summary":        summary,
		"records":        list,
	})
}

// SaveTrainingAttendance 保存训练出勤
func (c *TrainingPlanController) SaveTrainingAttendance(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	plan, ok := c.getClubTrainingPlan(ctx, club.ID)
	if !ok {
		return
	}

	var req struct {
		Records []trainingAttendanceRequestItem `json:"records" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(ctx, "参数错误")
		return
	}

	validTeamPlayers, err := c.getTrainingAttendancePlayerSet(plan)
	if err != nil {
		utils.ServerError(ctx, "获取球队球员失败")
		return
	}

	now := time.Now()
	err = c.db.Transaction(func(tx *gorm.DB) error {
		for _, record := range req.Records {
			if !isValidTrainingAttendanceStatus(record.Status) {
				return errInvalidTrainingAttendanceStatus
			}
			if !validTeamPlayers[record.PlayerID] {
				return errInvalidTrainingAttendancePlayer
			}
			attendance := models.TrainingAttendance{
				TrainingPlanID: plan.ID,
				ClubID:         plan.ClubID,
				TeamID:         plan.TeamID,
				PlayerID:       record.PlayerID,
				Status:         models.TrainingAttendanceStatus(record.Status),
				Remark:         record.Remark,
				RecordedBy:     userID,
				RecordedAt:     now,
			}
			var existing models.TrainingAttendance
			if err := tx.Where("training_plan_id = ? AND player_id = ?", plan.ID, record.PlayerID).First(&existing).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					if err := tx.Create(&attendance).Error; err != nil {
						return err
					}
					continue
				}
				return err
			}
			if err := tx.Model(&existing).Updates(map[string]interface{}{
				"status":      attendance.Status,
				"remark":      attendance.Remark,
				"recorded_by": attendance.RecordedBy,
				"recorded_at": attendance.RecordedAt,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		if err == errInvalidTrainingAttendanceStatus {
			utils.ValidationError(ctx, "出勤状态无效")
			return
		}
		if err == errInvalidTrainingAttendancePlayer {
			utils.ValidationError(ctx, "出勤球员不属于本次训练")
			return
		}
		utils.ServerError(ctx, "保存出勤失败")
		return
	}

	c.GetTrainingAttendance(ctx)
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
		Title            string   `json:"title"`
		Theme            string   `json:"theme"`
		Location         string   `json:"location"`
		StartTime        string   `json:"startTime"`
		EndTime          *string  `json:"endTime"`
		PlayerIDs        []uint   `json:"playerIds"`
		Content          string   `json:"content"`
		VideoURLs        []string `json:"videoUrls"`
		Summary          string   `json:"summary"`
		CompletionStatus string   `json:"completionStatus"`
		KeyPlayerNotes   string   `json:"keyPlayerNotes"`
		NextFocus        string   `json:"nextFocus"`
		Status           string   `json:"status"`
		WeeklyReportID   *uint    `json:"weeklyReportId"`
		PhysicalTestID   *uint    `json:"physicalTestId"`
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
	if req.CompletionStatus != "" {
		updates["completion_status"] = req.CompletionStatus
	}
	if req.KeyPlayerNotes != "" {
		updates["key_player_notes"] = req.KeyPlayerNotes
	}
	if req.NextFocus != "" {
		updates["next_focus"] = req.NextFocus
	}
	if req.Status != "" {
		updates["status"] = req.Status
		if req.Status == string(models.TrainingPlanStatusCompleted) {
			now := time.Now()
			updates["reviewed_by"] = userID
			updates["reviewed_at"] = &now
		}
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
		"id":               p.ID,
		"clubId":           p.ClubID,
		"teamId":           p.TeamID,
		"title":            p.Title,
		"theme":            p.Theme,
		"location":         p.Location,
		"startTime":        p.StartTime.Format(time.RFC3339),
		"endTime":          utils.FormatTime(p.EndTime),
		"playerIds":        p.GetPlayerIDs(),
		"content":          p.Content,
		"videoUrls":        p.GetVideoURLs(),
		"summary":          p.Summary,
		"completionStatus": p.CompletionStatus,
		"keyPlayerNotes":   p.KeyPlayerNotes,
		"nextFocus":        p.NextFocus,
		"coachId":          p.CoachID,
		"weeklyReportId":   p.WeeklyReportID,
		"physicalTestId":   p.PhysicalTestID,
		"status":           p.Status,
		"createdBy":        p.CreatedBy,
		"reviewedBy":       p.ReviewedBy,
		"reviewedAt":       utils.FormatTime(p.ReviewedAt),
		"createdAt":        p.CreatedAt.Format(time.RFC3339),
	}
}

func isValidTrainingAttendanceStatus(status string) bool {
	switch models.TrainingAttendanceStatus(status) {
	case models.TrainingAttendancePresent,
		models.TrainingAttendanceLeave,
		models.TrainingAttendanceAbsent,
		models.TrainingAttendanceLate:
		return true
	default:
		return false
	}
}

func emptyAttendanceSummary() map[string]int {
	return map[string]int{
		"total":    0,
		"present":  0,
		"leave":    0,
		"absent":   0,
		"late":     0,
		"unmarked": 0,
	}
}
