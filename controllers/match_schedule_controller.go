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

// MatchScheduleController 赛程日历控制器
type MatchScheduleController struct {
	clubService         *services.ClubService
	db                  *gorm.DB
	notificationService *services.NotificationService
}

// NewMatchScheduleController 创建赛程日历控制器
func NewMatchScheduleController(clubService *services.ClubService, db *gorm.DB) *MatchScheduleController {
	return &MatchScheduleController{clubService: clubService, db: db}
}

func (c *MatchScheduleController) SetNotificationService(notificationService *services.NotificationService) {
	c.notificationService = notificationService
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
	c.notifyMatchScheduleCreated(&schedule)
	utils.SuccessResponse(ctx, formatMatchSchedule(&schedule))
}

func (c *MatchScheduleController) notifyMatchScheduleCreated(schedule *models.MatchSchedule) {
	if c.notificationService == nil || schedule == nil {
		return
	}

	playerIDs := c.getActiveTeamPlayerIDs(schedule.TeamID)
	if len(playerIDs) == 0 {
		return
	}

	extra := map[string]interface{}{
		"team_id":    schedule.TeamID,
		"match_time": schedule.MatchTime.Format(time.RFC3339),
		"opponent":   schedule.Opponent,
		"location":   schedule.Location,
	}
	err := c.notificationService.NotifyTeamCalendarEvent(
		playerIDs,
		models.NotificationTypeMatchScheduleCreated,
		"新的比赛计划",
		schedule.Name+" 已加入球队日历，请做好赛前准备",
		"match_schedule",
		schedule.ID,
		"/user-dashboard?tab=team_calendar",
		extra,
	)
	if err != nil {
		log.Printf("发送比赛计划通知失败 (scheduleID=%d): %v", schedule.ID, err)
		return
	}
	_ = c.db.Model(schedule).Update("pre_remind_sent", true).Error
}

func (c *MatchScheduleController) getActiveTeamPlayerIDs(teamID uint) []uint {
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

// CreateMatchSummaryFromSchedule 从已结束赛程生成比赛总结任务
func (c *MatchScheduleController) CreateMatchSummaryFromSchedule(ctx *gin.Context) {
	userID := ctx.GetUint("userId")
	club, err := c.clubService.GetClubByUserID(userID)
	if err != nil || club == nil {
		utils.ForbiddenError(ctx, "无权限")
		return
	}

	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ValidationError(ctx, "无效的赛程ID")
		return
	}

	var schedule models.MatchSchedule
	if err := c.db.Where("id = ? AND club_id = ?", id, club.ID).First(&schedule).Error; err != nil {
		utils.NotFoundError(ctx, "赛程不存在")
		return
	}
	if schedule.MatchSummaryID != nil {
		utils.SuccessResponse(ctx, gin.H{
			"id":             *schedule.MatchSummaryID,
			"matchSummaryId": *schedule.MatchSummaryID,
			"created":        false,
		})
		return
	}
	if schedule.Status != models.MatchScheduleStatusCompleted {
		utils.ValidationError(ctx, "只有已结束赛程可以生成比赛总结")
		return
	}

	var team models.Team
	if err := c.db.Where("id = ? AND club_id = ?", schedule.TeamID, club.ID).First(&team).Error; err != nil {
		utils.NotFoundError(ctx, "球队不存在")
		return
	}

	var teamPlayers []models.TeamPlayer
	if err := c.db.Where("team_id = ? AND status = ?", schedule.TeamID, "active").Find(&teamPlayers).Error; err != nil {
		utils.ServerError(ctx, "获取参赛球员失败")
		return
	}
	playerIDs := make([]uint, 0, len(teamPlayers))
	for _, player := range teamPlayers {
		if player.UserID > 0 {
			playerIDs = append(playerIDs, player.UserID)
		}
	}
	if len(playerIDs) == 0 {
		utils.ValidationError(ctx, "该球队暂无在队球员，请先添加球员")
		return
	}

	ourScore := 0
	if schedule.HomeScore != nil {
		ourScore = *schedule.HomeScore
	}
	oppScore := 0
	if schedule.AwayScore != nil {
		oppScore = *schedule.AwayScore
	}
	result := "pending"
	if schedule.HomeScore != nil && schedule.AwayScore != nil {
		result = "draw"
		if ourScore > oppScore {
			result = "win"
		} else if ourScore < oppScore {
			result = "lose"
		}
	}

	playerIDsJSON, _ := json.Marshal(playerIDs)
	summary := models.MatchSummary{
		TeamID:      schedule.TeamID,
		CoachID:     userID,
		MatchName:   schedule.Name,
		MatchDate:   schedule.MatchTime.Format("2006-01-02"),
		Opponent:    schedule.Opponent,
		Location:    "home",
		MatchFormat: "11人制",
		OurScore:    ourScore,
		OppScore:    oppScore,
		Result:      result,
		PlayerIDs:   string(playerIDsJSON),
		PlayerCount: len(playerIDs),
		Status:      "pending",
	}

	alreadyLinkedErr := errors.New("match schedule already linked to summary")
	if err := c.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&summary).Error; err != nil {
			return err
		}
		result := tx.Model(&models.MatchSchedule{}).
			Where("id = ? AND match_summary_id IS NULL", schedule.ID).
			Update("match_summary_id", summary.ID)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return alreadyLinkedErr
		}
		return nil
	}); err != nil {
		if errors.Is(err, alreadyLinkedErr) {
			var latest models.MatchSchedule
			if reloadErr := c.db.First(&latest, schedule.ID).Error; reloadErr == nil && latest.MatchSummaryID != nil {
				utils.SuccessResponse(ctx, gin.H{
					"id":             *latest.MatchSummaryID,
					"matchSummaryId": *latest.MatchSummaryID,
					"created":        false,
				})
				return
			}
		}
		utils.ServerError(ctx, "创建比赛总结失败")
		return
	}

	go func() {
		notificationHelper := NewNotificationHelper(c.db)
		creatorName := "俱乐部管理员"
		var creator models.User
		if err := c.db.First(&creator, userID).Error; err == nil && creator.Name != "" {
			creatorName = creator.Name
		}
		for _, playerID := range playerIDs {
			notificationHelper.NotifyMatchSummaryCreated(playerID, creatorName, summary.MatchName, summary.ID)
		}
	}()

	utils.SuccessResponse(ctx, gin.H{
		"id":             summary.ID,
		"matchSummaryId": summary.ID,
		"created":        true,
	})
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
