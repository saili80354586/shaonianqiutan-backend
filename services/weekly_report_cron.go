package services

import (
	"log"
	"time"

	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/repositories"
	"gorm.io/gorm"
)

// WeeklyReportCron 周报定时任务
type WeeklyReportCron struct {
	db                 *gorm.DB
	weeklyService      *WeeklyReportService
	teamRepo           *repositories.TeamRepository
	notificationService *NotificationService
	isRunning          bool
}

// NewWeeklyReportCron 创建周报定时任务
func NewWeeklyReportCron(db *gorm.DB, weeklyService *WeeklyReportService, teamRepo *repositories.TeamRepository, notificationService *NotificationService) *WeeklyReportCron {
	return &WeeklyReportCron{
		db:                 db,
		weeklyService:      weeklyService,
		teamRepo:           teamRepo,
		notificationService: notificationService,
		isRunning:          false,
	}
}

// Start 启动定时任务
func (c *WeeklyReportCron) Start() {
	if c.isRunning {
		return
	}
	c.isRunning = true

	log.Println("[WeeklyReportCron] 启动周报定时任务...")

	// 创建定时任务通道（使用Go的ticker模拟cron）
	go c.runCron()
}

// Stop 停止定时任务
func (c *WeeklyReportCron) Stop() {
	c.isRunning = false
	log.Println("[WeeklyReportCron] 停止周报定时任务")
}

// runCron 运行定时任务循环
func (c *WeeklyReportCron) runCron() {
	// 每分钟检查一次是否需要执行任务
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for c.isRunning {
		select {
		case <-ticker.C:
			c.checkAndRunTasks()
		}
	}
}

// checkAndRunTasks 检查并执行到期的任务
func (c *WeeklyReportCron) checkAndRunTasks() {
	now := time.Now()

	// 每周五 00:00 自动发起周报
	if now.Weekday() == time.Friday && now.Hour() == 0 && now.Minute() == 0 {
		log.Println("[WeeklyReportCron] 执行每周五自动发起周报任务")
		go c.AutoCreateWeeklyReports()
	}

	// === 截止前多阶段提醒 ===
	// 周一 00:00: 48小时前提醒
	if now.Weekday() == time.Monday && now.Hour() == 0 && now.Minute() == 0 {
		log.Println("[WeeklyReportCron] 执行48小时前提醒任务")
		go c.SendDeadlineReminderWithMessage(48*time.Hour, "您的周报将在48小时后截止，请尽快提交")
	}

	// 周二 00:00: 24小时前提醒
	if now.Weekday() == time.Tuesday && now.Hour() == 0 && now.Minute() == 0 {
		log.Println("[WeeklyReportCron] 执行24小时前提醒任务")
		go c.SendDeadlineReminderWithMessage(24*time.Hour, "您的周报将在24小时后截止，请及时提交")
	}

	// 周三 12:00: 12小时前提醒
	if now.Weekday() == time.Wednesday && now.Hour() == 12 && now.Minute() == 0 {
		log.Println("[WeeklyReportCron] 执行12小时前提醒任务")
		go c.SendDeadlineReminderWithMessage(12*time.Hour, "您的周报将在12小时后截止，请尽快完成")
	}

	// 周三 22:00: 2小时前最后提醒
	if now.Weekday() == time.Wednesday && now.Hour() == 22 && now.Minute() == 0 {
		log.Println("[WeeklyReportCron] 执行2小时前最后提醒任务")
		go c.SendDeadlineReminderWithMessage(2*time.Hour, "您的周报将在2小时后截止，请立即提交")
	}

	// 每周四 00:00 自动归档逾期周报
	if now.Weekday() == time.Thursday && now.Hour() == 0 && now.Minute() == 0 {
		log.Println("[WeeklyReportCron] 执行周四自动归档任务")
		go c.weeklyService.AutoArchiveOverdueReports()
	}
}

// AutoCreateWeeklyReports 自动为所有球队创建周报
func (c *WeeklyReportCron) AutoCreateWeeklyReports() {
	// 获取当前周的起始日期（周一）
	now := time.Now()
	weekStart := now.AddDate(0, 0, -int(now.Weekday())+1) // 本周一
	if now.Weekday() == 0 {
		weekStart = now.AddDate(0, 0, -6) // 周日时，周一是6天前
	}
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())
	weekEnd := weekStart.AddDate(0, 0, 6) // 本周日

	// 设置截止时间为下周三 24:00
	deadline := weekStart.AddDate(0, 0, 9).Add(-time.Second) // 下周三 23:59:59

	log.Printf("[WeeklyReportCron] 自动发起周报: 周期 %s ~ %s, 截止时间: %s",
		weekStart.Format("2006-01-02"), weekEnd.Format("2006-01-02"), deadline.Format("2006-01-02 15:04:05"))

	// 获取所有球队
	var teams []models.Team
	if err := c.db.Find(&teams).Error; err != nil {
		log.Printf("[WeeklyReportCron] 获取球队列表失败: %v", err)
		return
	}

	totalCreated := 0
	for _, team := range teams {
		// 获取球队的所有球员
		var players []models.TeamPlayer
		c.db.Where("team_id = ? AND status = ?", team.ID, "active").Find(&players)

		if len(players) == 0 {
			continue
		}

		// 获取球队主教练
		var coachID uint
		var coaches []models.TeamCoach
		c.db.Where("team_id = ? AND role = ? AND status = ?", team.ID, models.CoachRoleHead, "active").
			Find(&coaches)
		if len(coaches) > 0 {
			coachID = coaches[0].UserID
		}

		// 获取球员ID列表
		playerIDs := make([]uint, 0, len(players))
		for _, p := range players {
			playerIDs = append(playerIDs, p.UserID)
		}

		// 批量创建周报
		result, err := c.weeklyService.BatchCreateWeeklyReports(team.ID, coachID, playerIDs, weekStart, weekEnd, &deadline)
		if err != nil {
			log.Printf("[WeeklyReportCron] 为球队 %d 创建周报失败: %v", team.ID, err)
			continue
		}

		created := result["created"].(int)
		totalCreated += created

		// 发送通知给球员
		for _, playerID := range playerIDs {
			// 通知球员周报已发起
			// TODO: 调用通知服务
			_ = playerID
		}

		log.Printf("[WeeklyReportCron] 球队 %d: 成功创建 %d 份周报", team.ID, created)
	}

	log.Printf("[WeeklyReportCron] 自动发起周报完成，共创建 %d 份周报", totalCreated)
}

// SendDeadlineReminder 发送截止提醒（兼容旧方法）
func (c *WeeklyReportCron) SendDeadlineReminder(before time.Duration) {
	c.SendDeadlineReminderWithMessage(before, "")
}

// SendDeadlineReminderWithMessage 发送带自定义消息的截止提醒
func (c *WeeklyReportCron) SendDeadlineReminderWithMessage(before time.Duration, message string) {
	now := time.Now()
	deadlineTime := now.Add(before)

	// 查找即将截止的周报
	var reports []models.WeeklyReport
	c.db.Where("submit_status = ? AND deadline <= ? AND deadline > ?",
		"draft", deadlineTime, now).Find(&reports)

	if len(reports) == 0 {
		log.Printf("[WeeklyReportCron] 没有需要提醒的周报 (before=%v)", before)
		return
	}

	// 计算小时数
	hoursRemaining := int(before.Hours())

	log.Printf("[WeeklyReportCron] 发送截止提醒（%d小时），共 %d 份待提交周报", hoursRemaining, len(reports))

	// 按球员分组统计
	playerReports := make(map[uint][]uint) // playerID -> reportIDs
	for _, r := range reports {
		playerReports[r.PlayerID] = append(playerReports[r.PlayerID], r.ID)
	}

	// 构建批量提醒信息
	reminders := make(map[uint]*WeeklyReportReminderInfo)
	for playerID, reportIDs := range playerReports {
		reminders[playerID] = &WeeklyReportReminderInfo{
			PlayerID:       playerID,
			HoursRemaining: hoursRemaining,
			ReportCount:    len(reportIDs),
			ReportIDs:      reportIDs,
		}
	}

	// 调用通知服务批量发送提醒
	if c.notificationService != nil {
		if err := c.notificationService.NotifyWeeklyReportReminderBatch(reminders); err != nil {
			log.Printf("[WeeklyReportCron] 发送提醒通知失败: %v", err)
		} else {
			log.Printf("[WeeklyReportCron] 成功发送提醒给 %d 名球员", len(reminders))
		}
	} else {
		log.Println("[WeeklyReportCron] 通知服务未初始化，跳过发送通知")
	}
}

// ManualTriggerAutoCreate 手动触发自动创建周报（用于测试）
func (c *WeeklyReportCron) ManualTriggerAutoCreate() {
	log.Println("[WeeklyReportCron] 手动触发自动创建周报")
	go c.AutoCreateWeeklyReports()
}

// ManualTriggerArchive 手动触发归档任务（用于测试）
func (c *WeeklyReportCron) ManualTriggerArchive() {
	log.Println("[WeeklyReportCron] 手动触发归档任务")
	go c.weeklyService.AutoArchiveOverdueReports()
}
