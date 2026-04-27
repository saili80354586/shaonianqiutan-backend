package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/repositories"
	"gorm.io/gorm"
)

// WeeklyReportService 周报服务
type WeeklyReportService struct {
	db                  *gorm.DB
	reportRepo          *repositories.WeeklyReportRepository
	teamRepo            *repositories.TeamRepository
	userRepo            *models.UserRepository
	notificationService *NotificationService
}

// NewWeeklyReportService 创建周报服务
func NewWeeklyReportService(
	db *gorm.DB,
	reportRepo *repositories.WeeklyReportRepository,
	teamRepo *repositories.TeamRepository,
	userRepo *models.UserRepository,
) *WeeklyReportService {
	return &WeeklyReportService{
		db:         db,
		reportRepo: reportRepo,
		teamRepo:   teamRepo,
		userRepo:   userRepo,
	}
}

// SetNotificationService 设置通知服务（避免循环依赖）
func (s *WeeklyReportService) SetNotificationService(service *NotificationService) {
	s.notificationService = service
}

// IsTeamCoach 检查用户是否是球队教练（公开方法）
func (s *WeeklyReportService) IsTeamCoach(teamID, userID uint) bool {
	var count int64
	s.db.Model(&models.TeamCoach{}).
		Where("team_id = ? AND user_id = ? AND status = ?", teamID, userID, "active").
		Count(&count)
	return count > 0
}

// isTeamCoach 检查用户是否是球队教练（内部方法，保持向后兼容）
func (s *WeeklyReportService) isTeamCoach(teamID, userID uint) bool {
	return s.IsTeamCoach(teamID, userID)
}

// isClubAdminOfTeam 检查用户是否是该球队所属俱乐部的管理员
func (s *WeeklyReportService) isClubAdminOfTeam(teamID, userID uint) bool {
	var club models.Club
	if err := s.db.Where("user_id = ?", userID).First(&club).Error; err != nil {
		return false
	}
	var count int64
	s.db.Model(&models.Team{}).
		Where("id = ? AND club_id = ?", teamID, club.ID).
		Count(&count)
	return count > 0
}

// CanManageTeam 检查用户是否可以管理球队周报（教练或俱乐部管理员）
func (s *WeeklyReportService) CanManageTeam(teamID, userID uint) bool {
	return s.isTeamCoach(teamID, userID) || s.isClubAdminOfTeam(teamID, userID)
}

// Submit 球员提交周报
func (s *WeeklyReportService) Submit(playerID uint, input *models.WeeklyReportSubmit) (*models.WeeklyReport, error) {
	// 解析周起始日期
	weekStart, err := time.Parse("2006-01-02", input.WeekStart)
	if err != nil {
		return nil, errors.New("无效的周起始日期")
	}

	// 计算周结束日期(周日)
	weekEnd := weekStart.AddDate(0, 0, 6)

	// 如果没有提供周结束日期，使用计算的
	weekEndStr := input.WeekEnd
	if weekEndStr == "" {
		weekEndStr = weekEnd.Format("2006-01-02")
	} else {
		weekEnd, _ = time.Parse("2006-01-02", weekEndStr)
	}

	// 获取球队信息
	team, err := s.teamRepo.FindByID(input.TeamID)
	if err != nil {
		return nil, errors.New("球队不存在")
	}

	// 获取球员信息
	player, err := s.userRepo.FindByID(playerID)
	if err != nil {
		return nil, errors.New("球员不存在")
	}

	// 获取球队的主教练
	var coachID uint
	var coaches []models.TeamCoach
	s.db.Where("team_id = ? AND role = ? AND status = ?", input.TeamID, models.CoachRoleHead, "active").
		Preload("User").Find(&coaches)
	if len(coaches) > 0 && coaches[0].UserID != 0 {
		coachID = coaches[0].UserID
	} else {
		// 如果没有主教练，使用第一个教练
		s.db.Where("team_id = ? AND status = ?", input.TeamID, "active").
			Preload("User").Find(&coaches)
		if len(coaches) > 0 {
			coachID = coaches[0].UserID
		}
	}

	// 检查是否已存在周报
	existing, _ := s.reportRepo.GetByPlayerAndWeek(playerID, input.WeekStart)
	if existing != nil {
		return nil, errors.New("该周周报已存在，如需更新请使用更新接口")
	}

	// 当前时间
	now := time.Now()

	// 创建周报
	report := &models.WeeklyReport{
		TeamID:    input.TeamID,
		PlayerID:  playerID,
		CoachID:   coachID,
		WeekStart: weekStart,
		WeekEnd:   weekEnd,

		// 训练出勤情况
		TrainingCount:    input.TrainingCount,
		TrainingDuration: input.TrainingDuration,
		AbsenceCount:     input.AbsenceCount,
		AbsenceReason:    input.AbsenceReason,

		// 训练内容反馈
		KnowledgeSummary:  input.KnowledgeSummary,
		TechnicalContent:  input.TechnicalContent,
		TacticalContent:   input.TacticalContent,
		PhysicalCondition: input.PhysicalCondition,
		MatchPerformance:  input.MatchPerformance,

		// 自我评价 - 多维度评分
		SelfAttitudeRating:  input.SelfAttitudeRating,
		SelfTechniqueRating: input.SelfTechniqueRating,
		SelfTeamworkRating:  input.SelfTeamworkRating,
		ImprovementsDetail:  input.ImprovementsDetail,
		Weaknesses:          input.Weaknesses,

		// 身体状态反馈
		FatigueLevel:  input.FatigueLevel,
		Injuries:      input.Injuries,
		SleepQuality:  input.SleepQuality,
		DietCondition: input.DietCondition,

		// 其他信息
		MessageToCoach: input.MessageToCoach,

		// 提交状态
		SubmitStatus: "submitted",
		SubmittedAt:  &now,
		ReviewStatus: "pending",
	}

	// 设置关联数据用于响应
	report.Player = player
	report.Team = team

	if err := s.reportRepo.Create(report); err != nil {
		return nil, err
	}

	return report, nil
}

// Update 球员更新周报(被退回后)
func (s *WeeklyReportService) Update(reportID, playerID uint, input *models.WeeklyReportSubmit) (*models.WeeklyReport, error) {
	report, err := s.reportRepo.GetByID(reportID)
	if err != nil {
		return nil, errors.New("周报不存在")
	}

	// 只能修改自己的周报
	if report.PlayerID != playerID {
		return nil, errors.New("无权修改此周报")
	}

	// 只有待填写、已逾期或被退回的才能修改
	if report.SubmitStatus != "draft" && report.SubmitStatus != "overdue" && report.ReviewStatus != "rejected" {
		return nil, errors.New("只有待填写、已逾期或被退回的周报才能修改")
	}

	// 更新字段
	// 训练出勤情况
	report.TrainingCount = input.TrainingCount
	report.TrainingDuration = input.TrainingDuration
	report.AbsenceCount = input.AbsenceCount
	report.AbsenceReason = input.AbsenceReason

	// 训练内容反馈
	report.KnowledgeSummary = input.KnowledgeSummary
	report.TechnicalContent = input.TechnicalContent
	report.TacticalContent = input.TacticalContent
	report.PhysicalCondition = input.PhysicalCondition
	report.MatchPerformance = input.MatchPerformance

	// 自我评价 - 多维度评分
	report.SelfAttitudeRating = input.SelfAttitudeRating
	report.SelfTechniqueRating = input.SelfTechniqueRating
	report.SelfTeamworkRating = input.SelfTeamworkRating
	report.ImprovementsDetail = input.ImprovementsDetail
	report.Weaknesses = input.Weaknesses

	// 身体状态反馈
	report.FatigueLevel = input.FatigueLevel
	report.Injuries = input.Injuries
	report.SleepQuality = input.SleepQuality
	report.DietCondition = input.DietCondition

	// 其他信息
	report.MessageToCoach = input.MessageToCoach

	// 重置审核状态
	report.SubmitStatus = "submitted"
	now := time.Now()
	report.SubmittedAt = &now
	report.ReviewStatus = "pending"
	report.ReviewCoachID = 0
	report.ReviewComment = ""
	report.CoachAttitudeRating = 0
	report.CoachTechniqueRating = 0
	report.CoachTacticsRating = 0
	report.CoachKnowledgeRating = 0
	report.StrengthsAcknowledgment = ""
	report.Suggestions = ""
	report.KnowledgeFeedback = ""
	report.NextWeekFocus = ""
	report.RecommendAward = false
	report.ReviewedAt = nil

	if err := s.reportRepo.Update(report); err != nil {
		return nil, err
	}

	return report, nil
}

// Review 教练审核周报
func (s *WeeklyReportService) Review(reportID, coachID uint, input *models.WeeklyReportReview) (*models.WeeklyReport, error) {
	report, err := s.reportRepo.GetByID(reportID)
	if err != nil {
		return nil, errors.New("周报不存在")
	}

	// 只能审核负责球队的周报（教练）
	if !s.isTeamCoach(report.TeamID, coachID) {
		return nil, errors.New("只有球队教练才能审核周报")
	}

	// 更新审核信息
	now := time.Now()
	report.ReviewStatus = input.Status
	report.ReviewCoachID = coachID

	// 多维度评分
	report.CoachAttitudeRating = input.CoachAttitudeRating
	report.CoachTechniqueRating = input.CoachTechniqueRating
	report.CoachTacticsRating = input.CoachTacticsRating
	report.CoachKnowledgeRating = input.CoachKnowledgeRating

	// 教练评语
	report.ReviewComment = input.ReviewComment
	report.StrengthsAcknowledgment = input.StrengthsAcknowledgment
	report.Suggestions = input.Suggestions
	report.KnowledgeFeedback = input.KnowledgeFeedback
	report.NextWeekFocus = input.NextWeekFocus
	report.RecommendAward = input.RecommendAward
	report.ReviewedAt = &now

	if err := s.reportRepo.Update(report); err != nil {
		return nil, err
	}

	// 获取教练信息
	coach, _ := s.userRepo.FindByID(coachID)
	if coach != nil {
		report.Coach = coach
	}

	return report, nil
}

// GetByID 获取周报详情
func (s *WeeklyReportService) GetByID(id uint) (*models.WeeklyReport, error) {
	return s.reportRepo.GetByID(id)
}

// ListByPlayer 列出球员周报
func (s *WeeklyReportService) ListByPlayer(playerID uint, page, pageSize int) ([]models.WeeklyReport, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}
	return s.reportRepo.ListByPlayer(playerID, page, pageSize)
}

// ListByTeam 列出球队周报(教练用)
func (s *WeeklyReportService) ListByTeam(teamID uint, status string, page, pageSize int) ([]models.WeeklyReport, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}
	return s.reportRepo.ListByTeam(teamID, status, page, pageSize)
}

// ListPendingByCoach 列出待审核周报
func (s *WeeklyReportService) ListPendingByCoach(coachID uint, page, pageSize int) ([]models.WeeklyReport, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}
	return s.reportRepo.ListPendingByCoach(coachID, page, pageSize)
}

// GetPendingCount 获取待审核数量
func (s *WeeklyReportService) GetPendingCount(teamID uint) (int64, error) {
	return s.reportRepo.CountPending(teamID)
}

// Delete 删除周报
func (s *WeeklyReportService) Delete(id, playerID uint) error {
	report, err := s.reportRepo.GetByID(id)
	if err != nil {
		return errors.New("周报不存在")
	}

	// 只能删除自己的周报
	if report.PlayerID != playerID {
		return errors.New("无权删除此周报")
	}

	// 只能删除待审核的
	if report.ReviewStatus != "pending" && report.ReviewStatus != "rejected" {
		return errors.New("只能删除待审核或已退回的周报")
	}

	return s.reportRepo.Delete(id)
}

// ==================== 周报周期管理 ====================

// CreatePeriod 创建周报周期
func (s *WeeklyReportService) CreatePeriod(teamID uint, weekStart, weekEnd time.Time, deadline *time.Time) (*models.WeeklyReportPeriod, error) {
	// 检查是否已存在相同周期的记录
	var existing models.WeeklyReportPeriod
	if err := s.db.Where("team_id = ? AND week_start = ?", teamID, weekStart).First(&existing).Error; err == nil {
		return nil, errors.New("该周期周报已存在")
	}

	period := &models.WeeklyReportPeriod{
		TeamID:    teamID,
		WeekStart: weekStart,
		WeekEnd:   weekEnd,
		Deadline:  deadline,
		Status:    "active",
	}

	if err := s.db.Create(period).Error; err != nil {
		return nil, err
	}

	return period, nil
}

// GetPeriodsByTeam 获取球队的周报周期列表
func (s *WeeklyReportService) GetPeriodsByTeam(teamID uint, status string, page, pageSize int) ([]models.WeeklyReportPeriod, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	var periods []models.WeeklyReportPeriod
	var total int64

	query := s.db.Model(&models.WeeklyReportPeriod{}).Where("team_id = ?", teamID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("week_start DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&periods).Error; err != nil {
		return nil, 0, err
	}

	return periods, total, nil
}

// GetPeriodStats 获取周期统计信息
func (s *WeeklyReportService) GetPeriodStats(periodID uint) (*models.WeeklyReportPeriodResponse, error) {
	var period models.WeeklyReportPeriod
	if err := s.db.First(&period, periodID).Error; err != nil {
		return nil, errors.New("周期不存在")
	}

	// 统计周报数据
	var totalReports int64
	var submittedCount int64
	var pendingCount int64
	var overdueCount int64
	var reviewedCount int64

	s.db.Model(&models.WeeklyReport{}).Where("team_id = ? AND week_start = ?", period.TeamID, period.WeekStart).Count(&totalReports)
	s.db.Model(&models.WeeklyReport{}).Where("team_id = ? AND week_start = ? AND submit_status = ?", period.TeamID, period.WeekStart, "submitted").Count(&submittedCount)
	s.db.Model(&models.WeeklyReport{}).Where("team_id = ? AND week_start = ? AND submit_status = ?", period.TeamID, period.WeekStart, "draft").Count(&pendingCount)
	s.db.Model(&models.WeeklyReport{}).Where("team_id = ? AND week_start = ? AND submit_status = ?", period.TeamID, period.WeekStart, "overdue").Count(&overdueCount)
	s.db.Model(&models.WeeklyReport{}).Where("team_id = ? AND week_start = ? AND review_status = ?", period.TeamID, period.WeekStart, "approved").Count(&reviewedCount)

	// 更新统计数据
	period.TotalPlayers = int(totalReports)
	period.SubmittedCount = int(submittedCount)
	period.PendingCount = int(pendingCount)
	period.OverdueCount = int(overdueCount)
	period.ReviewedCount = int(reviewedCount)

	// 保存更新后的统计
	s.db.Save(&period)

	resp := period.ToResponse()
	return &resp, nil
}

// GetPeriodPlayers 获取周期内的球员提交情况
func (s *WeeklyReportService) GetPeriodPlayers(teamID uint, weekStart time.Time, status string, page, pageSize int) ([]models.WeeklyReport, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}

	var reports []models.WeeklyReport
	var total int64

	weekStartStr := weekStart.Format("2006-01-02")
	weekStartEndStr := weekStart.AddDate(0, 0, 1).Format("2006-01-02")

	query := s.db.Model(&models.WeeklyReport{}).
		Where("team_id = ? AND week_start >= ? AND week_start < ?", teamID, weekStartStr, weekStartEndStr).
		Preload("Player").
		Preload("Coach")

	if status != "" {
		query = query.Where("review_status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("player_id").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&reports).Error; err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

// BatchCreateWeeklyReports 批量创建周报（教练/管理员发起周报）
func (s *WeeklyReportService) BatchCreateWeeklyReports(teamID, coachID uint, playerIDs []uint, weekStart, weekEnd time.Time, deadline *time.Time) (map[string]interface{}, error) {
	// 验证权限（教练或俱乐部管理员）
	if !s.CanManageTeam(teamID, coachID) {
		return nil, errors.New("无权为此球队创建周报")
	}

	// 创建周期记录
	period, err := s.CreatePeriod(teamID, weekStart, weekEnd, deadline)
	if err != nil && !errors.Is(err, errors.New("该周期周报已存在")) {
		return nil, err
	}

	created := 0
	failed := 0

	for _, playerID := range playerIDs {
		// 检查是否已存在
		var existing models.WeeklyReport
		err := s.db.Where("team_id = ? AND player_id = ? AND week_start = ?", teamID, playerID, weekStart).First(&existing).Error
		if err == nil {
			failed++
			continue
		}

		report := &models.WeeklyReport{
			TeamID:       teamID,
			PlayerID:     playerID,
			CoachID:      coachID,
			WeekStart:    weekStart,
			WeekEnd:      weekEnd,
			Deadline:     deadline,
			SubmitStatus: "draft",
			ReviewStatus: "pending",
		}

		if err := s.db.Create(report).Error; err != nil {
			failed++
			continue
		}
		created++

		// 更新周期统计
		if period != nil {
			period.TotalPlayers++
			period.PendingCount++
			s.db.Save(period)
		}
	}

	// 发送通知给球员
	if s.notificationService != nil && created > 0 {
		// 获取教练姓名
		coachName := "教练"
		if coachID > 0 {
			if user, err := s.userRepo.FindByID(coachID); err == nil && user != nil {
				coachName = user.Nickname
			}
		}

		// 获取球队名称
		teamName := "球队"
		if team, err := s.teamRepo.FindByID(teamID); err == nil && team != nil {
			teamName = team.Name
		}

		// 构建周标签
		weekLabel := fmt.Sprintf("%s ~ %s", weekStart.Format("01/02"), weekEnd.Format("01/02"))

		// 为每个球员发送通知
		for _, playerID := range playerIDs {
			// 查找刚创建的周报ID
			var report models.WeeklyReport
			if err := s.db.Where("team_id = ? AND player_id = ? AND week_start = ?",
				teamID, playerID, weekStart).First(&report).Error; err == nil {
				if err := s.notificationService.NotifyWeeklyReportCreated(playerID, coachName, teamName, weekLabel, report.ID); err != nil {
					fmt.Printf("发送周报发起通知失败 (player %d): %v\n", playerID, err)
				}
			}
		}
	}

	return map[string]interface{}{
		"created": created,
		"failed":  failed,
		"period":  period,
		"message": fmt.Sprintf("成功创建 %d 份周报", created),
	}, nil
}

// AutoArchiveOverdueReports 自动归档逾期周报
func (s *WeeklyReportService) AutoArchiveOverdueReports() error {
	now := time.Now()

	// 查找已截止但未提交的周报
	var reports []models.WeeklyReport
	s.db.Where("submit_status = ? AND deadline < ?", "draft", now).Find(&reports)

	for _, report := range reports {
		report.SubmitStatus = "overdue"
		s.db.Save(&report)
	}

	// 查找已关闭的周期并归档
	var periods []models.WeeklyReportPeriod
	s.db.Where("status = ? AND deadline < ?", "active", now.Add(-24*time.Hour)).Find(&periods)

	for _, period := range periods {
		period.Status = "closed"
		s.db.Save(&period)
	}

	return nil
}

// RemindPlayers 一键提醒未提交周报的球员
func (s *WeeklyReportService) RemindPlayers(teamID uint, coachID uint, weekStart string, playerIDs []uint, customMessage string) (map[string]interface{}, error) {
	// 验证权限（教练或俱乐部管理员）
	if !s.CanManageTeam(teamID, coachID) {
		return nil, errors.New("无权为此球队发送提醒")
	}

	// 解析周期起始日期
	weekStartTime, err := time.Parse("2006-01-02", weekStart)
	if err != nil {
		return nil, errors.New("无效的周期起始日期")
	}

	// 获取教练信息
	coach, _ := s.userRepo.FindByID(coachID)
	coachName := "教练"
	if coach != nil {
		coachName = coach.Name
	}

	// 如果没有指定球员ID，查询所有未提交的球员
	if len(playerIDs) == 0 {
		var pendingReports []models.WeeklyReport
		s.db.Where("team_id = ? AND week_start = ? AND submit_status = ?", teamID, weekStartTime, "draft").
			Find(&pendingReports)
		for _, r := range pendingReports {
			playerIDs = append(playerIDs, r.PlayerID)
		}
	}

	// 发送提醒
	remindedCount := 0
	failedCount := 0
	var failedPlayerIDs []uint

	for _, playerID := range playerIDs {
		// 查询该球员的周报
		var report models.WeeklyReport
		err := s.db.Where("team_id = ? AND player_id = ? AND week_start = ?", teamID, playerID, weekStartTime).First(&report).Error
		if err != nil {
			failedCount++
			failedPlayerIDs = append(failedPlayerIDs, playerID)
			continue
		}

		// 只提醒未提交的周报
		if report.SubmitStatus != "draft" {
			continue
		}

		// 验证球员存在
		_, err = s.userRepo.FindByID(playerID)
		if err != nil {
			failedCount++
			failedPlayerIDs = append(failedPlayerIDs, playerID)
			continue
		}

		// 构建提醒消息
		message := customMessage
		if message == "" {
			message = fmt.Sprintf("%s提醒您：请及时提交本周训练周报", coachName)
		}

		// 计算剩余小时数
		hoursRemaining := 48
		if report.Deadline != nil {
			remaining := time.Until(*report.Deadline)
			if remaining > 0 {
				hoursRemaining = int(remaining.Hours())
				if hoursRemaining < 1 {
					hoursRemaining = 1
				}
			}
		}

		// 发送通知（这里使用现有的通知系统）
		if s.notificationService != nil {
			err := s.notificationService.NotifyWeeklyReportReminder(
				playerID, hoursRemaining, 1, []uint{report.ID},
			)
			if err != nil {
				failedCount++
				failedPlayerIDs = append(failedPlayerIDs, playerID)
				continue
			}
		}

		remindedCount++
	}

	return map[string]interface{}{
		"remindedCount":   remindedCount,
		"failedCount":     failedCount,
		"failedPlayerIDs": failedPlayerIDs,
		"message":         fmt.Sprintf("成功提醒 %d 位球员", remindedCount),
	}, nil
}

// ExportWeeklyReports 导出周报为CSV格式
func (s *WeeklyReportService) ExportWeeklyReports(teamID uint, weekStart, status string) (string, string, error) {
	// 构建查询
	query := s.db.Model(&models.WeeklyReport{}).
		Where("team_id = ?", teamID).
		Preload("Player")

	// 按周期筛选
	if weekStart != "" {
		weekStartTime, err := time.Parse("2006-01-02", weekStart)
		if err != nil {
			return "", "", errors.New("无效的周期起始日期")
		}
		query = query.Where("week_start = ?", weekStartTime)
	}

	// 按状态筛选
	if status != "" {
		query = query.Where("submit_status = ?", status)
	}

	// 获取数据
	var reports []models.WeeklyReport
	if err := query.Order("week_start DESC, player_id").Find(&reports).Error; err != nil {
		return "", "", err
	}

	// 构建CSV内容
	var sb strings.Builder

	// 写入表头
	headers := []string{
		"球员姓名", "周期", "提交状态", "审核状态",
		"训练次数", "训练时长(分钟)", "请假次数",
		"态度自评", "技术自评", "协作自评",
		"疲劳程度", "睡眠质量",
		"态度评分", "技术评分", "战术评分", "知识评分",
		"提交时间", "审核时间",
	}
	sb.WriteString(strings.Join(headers, ",") + "\n")

	// 写入数据行
	for _, r := range reports {
		playerName := ""
		if r.Player != nil {
			playerName = r.Player.Name
		}

		weekLabel := r.WeekStart.Format("2006年第2周")
		submitStatus := map[string]string{
			"draft":     "未提交",
			"submitted": "已提交",
			"overdue":   "已逾期",
		}[r.SubmitStatus]
		if submitStatus == "" {
			submitStatus = r.SubmitStatus
		}

		reviewStatus := map[string]string{
			"pending":  "待审核",
			"approved": "已通过",
			"rejected": "已退回",
		}[r.ReviewStatus]
		if reviewStatus == "" {
			reviewStatus = r.ReviewStatus
		}

		submittedAt := ""
		if r.SubmittedAt != nil {
			submittedAt = r.SubmittedAt.Format("2006-01-02 15:04")
		}

		reviewedAt := ""
		if r.ReviewedAt != nil {
			reviewedAt = r.ReviewedAt.Format("2006-01-02 15:04")
		}

		row := []string{
			playerName,
			weekLabel,
			submitStatus,
			reviewStatus,
			fmt.Sprintf("%d", r.TrainingCount),
			fmt.Sprintf("%d", r.TrainingDuration),
			fmt.Sprintf("%d", r.AbsenceCount),
			fmt.Sprintf("%d", r.SelfAttitudeRating),
			fmt.Sprintf("%d", r.SelfTechniqueRating),
			fmt.Sprintf("%d", r.SelfTeamworkRating),
			fmt.Sprintf("%d", r.FatigueLevel),
			fmt.Sprintf("%d", r.SleepQuality),
			fmt.Sprintf("%d", r.CoachAttitudeRating),
			fmt.Sprintf("%d", r.CoachTechniqueRating),
			fmt.Sprintf("%d", r.CoachTacticsRating),
			fmt.Sprintf("%d", r.CoachKnowledgeRating),
			submittedAt,
			reviewedAt,
		}

		// 处理字段中的逗号和换行
		for i, v := range row {
			if strings.Contains(v, ",") || strings.Contains(v, "\n") || strings.Contains(v, "\"") {
				row[i] = `"` + strings.ReplaceAll(v, `"`, `""`) + `"`
			}
		}

		sb.WriteString(strings.Join(row, ",") + "\n")
	}

	// 生成文件名
	filename := fmt.Sprintf("weekly_reports_%d_%s.csv", teamID, time.Now().Format("20060102_150405"))

	return sb.String(), filename, nil
}
