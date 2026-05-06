package services

import (
	"errors"
	"log"
	"time"

	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/repositories"
	"github.com/shaonianqiutan/backend/wshub"
	"gorm.io/gorm"
)

// NotificationService 通知服务
type NotificationService struct {
	db               *gorm.DB
	notificationRepo *repositories.NotificationRepository
	userRepo         *models.UserRepository
}

// NewNotificationService 创建通知服务
func NewNotificationService(
	db *gorm.DB,
	notificationRepo *repositories.NotificationRepository,
	userRepo *models.UserRepository,
) *NotificationService {
	return &NotificationService{
		db:               db,
		notificationRepo: notificationRepo,
		userRepo:         userRepo,
	}
}

// CreateNotification 创建单条通知
func (s *NotificationService) CreateNotification(userID uint, notificationType models.NotificationType, title, content string, data *models.NotificationData) (*models.Notification, error) {
	notification := &models.Notification{
		UserID:    userID,
		Type:      notificationType,
		Title:     title,
		Content:   content,
		IsRead:    false,
		Priority:  3,
		CreatedAt: time.Now(),
	}

	if data != nil {
		notification.SetData(data)
	}

	if err := s.notificationRepo.Create(notification); err != nil {
		return nil, err
	}

	// WebSocket 实时推送
	go s.sendWebSocketNotification(userID, notificationType, notification)

	return notification, nil
}

// sendWebSocketNotification 发送 WebSocket 通知
func (s *NotificationService) sendWebSocketNotification(userID uint, notificationType models.NotificationType, notification *models.Notification) {
	notifyService := wshub.GetNotifyService()
	payload := wshub.NotificationPayload{
		ID:        notification.ID,
		Title:     notification.Title,
		Content:   notification.Content,
		Data:      notification.GetData(),
		CreatedAt: notification.CreatedAt.Format(time.RFC3339),
	}

	switch notificationType {
	case models.NotificationTypeWeeklyReportCreated:
		notifyService.SendNotification(userID, "weekly_report_created", payload)
	case models.NotificationTypeWeeklyReportRejected:
		notifyService.SendNotification(userID, "weekly_report_rejected", payload)
	case models.NotificationTypeWeeklyReportApproved:
		notifyService.SendNotification(userID, "weekly_report_approved", payload)
	case models.NotificationTypeWeeklyReportReminder:
		notifyService.SendNotification(userID, "weekly_report_reminder", payload)
	case models.NotificationTypeMatchSummaryCreated:
		notifyService.SendNotification(userID, "match_summary_created", payload)
	case models.NotificationTypeMatchPlayerReminder:
		notifyService.SendNotification(userID, "match_player_reminder", payload)
	case models.NotificationTypeMatchCoachReminder:
		notifyService.SendNotification(userID, "match_coach_reminder", payload)
	case models.NotificationTypeMatchSummaryComplete:
		notifyService.SendNotification(userID, "match_summary_complete", payload)
	default:
		notifyService.SendNotification(userID, string(notificationType), payload)
	}
	log.Printf("WebSocket: 通知 %d 已推送给用户 %d", notification.ID, userID)
}

// CreateBatchNotifications 批量创建通知
func (s *NotificationService) CreateBatchNotifications(userIDs []uint, notificationType models.NotificationType, title, content string, data *models.NotificationData) error {
	notifications := make([]*models.Notification, 0, len(userIDs))
	now := time.Now()

	for _, userID := range userIDs {
		notification := &models.Notification{
			UserID:    userID,
			Type:      notificationType,
			Title:     title,
			Content:   content,
			IsRead:    false,
			Priority:  3,
			CreatedAt: now,
		}
		if data != nil {
			notification.SetData(data)
		}
		notifications = append(notifications, notification)
	}

	if err := s.notificationRepo.CreateBatch(notifications); err != nil {
		return err
	}

	// WebSocket 实时推送（批量）
	go func() {
		notifyService := wshub.GetNotifyService()
		for _, notification := range notifications {
			payload := wshub.NotificationPayload{
				ID:        notification.ID,
				Title:     notification.Title,
				Content:   notification.Content,
				Data:      notification.GetData(),
				CreatedAt: notification.CreatedAt.Format(time.RFC3339),
			}
			notifyService.SendNotification(notification.UserID, string(notificationType), payload)
		}
		log.Printf("WebSocket: 批量通知已推送给 %d 个用户", len(userIDs))
	}()

	return nil
}

// NotifyAnalystOrderAssigned 通知分析师有新订单待接单
func (s *NotificationService) NotifyAnalystOrderAssigned(analystUserID, orderID uint, orderNo, playerName string) error {
	title := "新订单已分配"
	content := "订单 " + orderNo + " 已分配给你，请及时接单"
	if playerName != "" {
		content = "球员 " + playerName + " 的订单 " + orderNo + " 已分配给你，请及时接单"
	}
	data := &models.NotificationData{
		TargetType: "order",
		TargetID:   orderID,
		Link:       "/analyst/dashboard",
	}
	_, err := s.CreateNotification(analystUserID, models.NotificationTypeOrder, title, content, data)
	return err
}

// NotifyAdminsOrderRejected 通知管理员分析师拒单
func (s *NotificationService) NotifyAdminsOrderRejected(adminUserIDs []uint, orderID uint, orderNo, reason string) error {
	title := "分析师拒绝订单"
	content := "订单 " + orderNo + " 已被分析师拒绝，请重新派单"
	if reason != "" {
		content += "，原因：" + reason
	}
	data := &models.NotificationData{
		TargetType: "order",
		TargetID:   orderID,
		Link:       "/admin/orders",
	}
	return s.CreateBatchNotifications(adminUserIDs, models.NotificationTypeOrder, title, content, data)
}

// NotifyReportPendingReview 通知管理员有报告待审核
func (s *NotificationService) NotifyReportPendingReview(adminUserIDs []uint, reportID uint, playerName string) error {
	title := "报告待审核"
	content := "有一份新分析报告等待审核"
	if playerName != "" {
		content = "球员 " + playerName + " 的分析报告等待审核"
	}
	data := &models.NotificationData{
		TargetType: "report",
		TargetID:   reportID,
		ReportID:   reportID,
		Link:       "/admin/reports",
	}
	return s.CreateBatchNotifications(adminUserIDs, models.NotificationTypeReport, title, content, data)
}

// NotifyReportCompleted 通知用户报告已通过审核
func (s *NotificationService) NotifyReportCompleted(playerUserID, reportID uint, playerName string) error {
	title := "球探报告已完成"
	content := "你的球探报告已审核完成，可以查看了"
	if playerName != "" {
		content = "球员 " + playerName + " 的球探报告已审核完成，可以查看了"
	}
	data := &models.NotificationData{
		TargetType: "report",
		TargetID:   reportID,
		ReportID:   reportID,
		Link:       "/reports/" + itoa(int(reportID)),
	}
	_, err := s.CreateNotification(playerUserID, models.NotificationTypeReport, title, content, data)
	return err
}

// NotifyAnalystReportRejected 通知分析师报告被退回
func (s *NotificationService) NotifyAnalystReportRejected(analystUserID, reportID uint, playerName, remark string) error {
	title := "报告被退回"
	content := "你提交的分析报告未通过审核，请修改后重新提交"
	if playerName != "" {
		content = "球员 " + playerName + " 的分析报告未通过审核，请修改后重新提交"
	}
	if remark != "" {
		content += "，原因：" + remark
	}
	data := &models.NotificationData{
		TargetType: "report",
		TargetID:   reportID,
		ReportID:   reportID,
		Link:       "/analyst/dashboard",
	}
	_, err := s.CreateNotification(analystUserID, models.NotificationTypeReport, title, content, data)
	return err
}

// NotifyWeeklyReportCreated 通知球员教练发起了周报
func (s *NotificationService) NotifyWeeklyReportCreated(playerID uint, coachName, teamName, weekLabel string, reportID uint) error {
	title := "教练发起了本周周报"
	content := coachName + " 发起了一篇新周报，请填写"
	data := &models.NotificationData{
		TargetType: "weekly_report",
		TargetID:   reportID,
		Link:       "/weekly-reports/" + itoa(int(reportID)) + "/edit",
	}
	_, err := s.CreateNotification(playerID, models.NotificationTypeWeeklyReportCreated, title, content, data)
	return err
}

// NotifyWeeklyReportRejected 通知球员周报被退回
func (s *NotificationService) NotifyWeeklyReportRejected(playerID uint, coachName, comment string, reportID uint) error {
	title := "周报被退回"
	content := coachName + " 退回了你的周报，请修改后重新提交"
	data := &models.NotificationData{
		TargetType:     "weekly_report",
		TargetID:       reportID,
		CommentContent: comment,
		Link:           "/weekly-reports/" + itoa(int(reportID)) + "/edit",
	}
	_, err := s.CreateNotification(playerID, models.NotificationTypeWeeklyReportRejected, title, content, data)
	return err
}

// NotifyWeeklyReportApproved 通知球员周报审核完成
func (s *NotificationService) NotifyWeeklyReportApproved(playerID uint, coachName string, rating int, reportID uint) error {
	title := "周报已审核"
	content := coachName + " 已审核你的周报，评分 " + itoa(rating) + " 星"
	data := &models.NotificationData{
		TargetType: "weekly_report",
		TargetID:   reportID,
		Link:       "/weekly-reports/" + itoa(int(reportID)),
	}
	_, err := s.CreateNotification(playerID, models.NotificationTypeWeeklyReportApproved, title, content, data)
	return err
}

// NotifyWeeklyReportReminder 通知球员周报即将截止
func (s *NotificationService) NotifyWeeklyReportReminder(playerID uint, hoursRemaining int, reportCount int, reportIDs []uint) error {
	title := "周报即将截止"
	content := "您有 " + itoa(reportCount) + " 篇周报将在 " + itoa(hoursRemaining) + " 小时后截止，请尽快提交"

	// 构建链接列表
	var links []string
	for _, id := range reportIDs {
		links = append(links, "/weekly-reports/"+itoa(int(id))+"/edit")
	}

	data := &models.NotificationData{
		TargetType: "weekly_report",
		TargetIDs:  reportIDs,
		Link:       "/weekly-reports",
		Extra: map[string]interface{}{
			"hours_remaining": hoursRemaining,
			"report_count":    reportCount,
			"links":           links,
		},
	}
	_, err := s.CreateNotification(playerID, models.NotificationTypeWeeklyReportReminder, title, content, data)
	return err
}

// NotifyWeeklyReportReminderBatch 批量发送周报截止提醒
func (s *NotificationService) NotifyWeeklyReportReminderBatch(reminders map[uint]*WeeklyReportReminderInfo) error {
	for playerID, info := range reminders {
		title := "周报即将截止"
		content := "您有 " + itoa(info.ReportCount) + " 篇周报将在 " + itoa(info.HoursRemaining) + " 小时后截止，请尽快提交"

		var links []string
		for _, id := range info.ReportIDs {
			links = append(links, "/weekly-reports/"+itoa(int(id))+"/edit")
		}

		data := &models.NotificationData{
			TargetType: "weekly_report",
			TargetIDs:  info.ReportIDs,
			Link:       "/weekly-reports",
			Extra: map[string]interface{}{
				"hours_remaining": info.HoursRemaining,
				"report_count":    info.ReportCount,
				"links":           links,
			},
		}

		notification := &models.Notification{
			UserID:    playerID,
			Type:      models.NotificationTypeWeeklyReportReminder,
			Title:     title,
			Content:   content,
			IsRead:    false,
			Priority:  3,
			CreatedAt: time.Now(),
		}
		notification.SetData(data)

		if err := s.notificationRepo.Create(notification); err != nil {
			log.Printf("创建提醒通知失败 (player %d): %v", playerID, err)
			continue
		}

		// WebSocket 推送
		go s.sendWebSocketNotification(playerID, models.NotificationTypeWeeklyReportReminder, notification)
	}
	return nil
}

// WeeklyReportReminderInfo 周报提醒信息
type WeeklyReportReminderInfo struct {
	PlayerID       uint
	HoursRemaining int
	ReportCount    int
	ReportIDs      []uint
}

// NotifyMatchSummaryCreated 通知球员比赛即将开始
func (s *NotificationService) NotifyMatchSummaryCreated(playerID uint, coachName, matchName string, matchID uint) error {
	title := "比赛即将开始"
	content := coachName + " 创建了比赛 " + matchName + "，请做好准备"
	data := &models.NotificationData{
		TargetType: "match_summary",
		TargetID:   matchID,
		Link:       "/match-reports/" + itoa(int(matchID)),
	}
	_, err := s.CreateNotification(playerID, models.NotificationTypeMatchSummaryCreated, title, content, data)
	return err
}

// NotifyMatchPlayerReminder 通知球员填写自评
func (s *NotificationService) NotifyMatchPlayerReminder(playerID uint, matchName string, matchID uint) error {
	title := "比赛已结束，请填写自评"
	content := matchName + " 已结束，请填写你的比赛自评"
	data := &models.NotificationData{
		TargetType: "match_summary",
		TargetID:   matchID,
		Link:       "/match-reports/" + itoa(int(matchID)) + "/player-summary",
	}
	_, err := s.CreateNotification(playerID, models.NotificationTypeMatchPlayerReminder, title, content, data)
	return err
}

// NotifyMatchCoachReminder 通知教练填写点评
func (s *NotificationService) NotifyMatchCoachReminder(coachID uint, matchName string, matchID uint) error {
	title := "比赛已结束，请填写点评"
	content := matchName + " 已结束，请填写你的比赛点评"
	data := &models.NotificationData{
		TargetType: "match_summary",
		TargetID:   matchID,
		Link:       "/match-reports/" + itoa(int(matchID)) + "/coach-summary",
	}
	_, err := s.CreateNotification(coachID, models.NotificationTypeMatchCoachReminder, title, content, data)
	return err
}

// NotifyMatchSummaryComplete 通知相关人员点评完成
func (s *NotificationService) NotifyMatchSummaryComplete(playerID uint, coachName, matchName string, matchID uint) error {
	title := "比赛总结已完成"
	content := coachName + " 已完成 " + matchName + " 的点评"
	data := &models.NotificationData{
		TargetType: "match_summary",
		TargetID:   matchID,
		Link:       "/match-reports/" + itoa(int(matchID)),
	}
	_, err := s.CreateNotification(playerID, models.NotificationTypeMatchSummaryComplete, title, content, data)
	return err
}

// GetByID 获取通知详情
func (s *NotificationService) GetByID(id uint) (*models.Notification, error) {
	return s.notificationRepo.GetByID(id)
}

// ListByUser 获取用户通知列表
func (s *NotificationService) ListByUser(userID uint, page, pageSize int) ([]models.Notification, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	return s.notificationRepo.ListByUser(userID, page, pageSize)
}

// ListUnread 获取未读通知列表
func (s *NotificationService) ListUnread(userID uint, page, pageSize int) ([]models.Notification, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	return s.notificationRepo.ListUnread(userID, page, pageSize)
}

// CountUnread 获取未读数量
func (s *NotificationService) CountUnread(userID uint) (int64, error) {
	return s.notificationRepo.CountUnread(userID)
}

// MarkAsRead 标记通知为已读
func (s *NotificationService) MarkAsRead(userID, notificationID uint) error {
	notification, err := s.notificationRepo.GetByID(notificationID)
	if err != nil {
		return errors.New("通知不存在")
	}

	// 验证归属
	if notification.UserID != userID {
		return errors.New("无权操作此通知")
	}

	return s.notificationRepo.MarkAsRead(notificationID)
}

// MarkAllAsRead 标记全部为已读
func (s *NotificationService) MarkAllAsRead(userID uint) error {
	return s.notificationRepo.MarkAllAsRead(userID)
}

// Delete 删除通知
func (s *NotificationService) Delete(userID, notificationID uint) error {
	notification, err := s.notificationRepo.GetByID(notificationID)
	if err != nil {
		return errors.New("通知不存在")
	}

	// 验证归属
	if notification.UserID != userID {
		return errors.New("无权操作此通知")
	}

	return s.notificationRepo.Delete(notificationID)
}

// itoa 简单数字转字符串
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}
