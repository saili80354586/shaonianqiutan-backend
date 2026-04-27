package controllers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shaonianqiutan/backend/models"
	"gorm.io/gorm"
)

// NotificationHelper 通知辅助函数
type NotificationHelper struct {
	db *gorm.DB
}

// NewNotificationHelper 创建通知辅助函数
func NewNotificationHelper(db *gorm.DB) *NotificationHelper {
	return &NotificationHelper{db: db}
}

// CreateNotification 创建通知
func (h *NotificationHelper) CreateNotification(userID uint, notificationType models.NotificationType, title, content string, link string) error {
	notification := &models.Notification{
		UserID:    userID,
		Type:      notificationType,
		Title:     title,
		Content:   content,
		IsRead:    false,
		Priority:  3,
		CreatedAt: time.Now(),
	}
	if link != "" {
		notification.Data = `{"link":"` + link + `"}`
	}
	return h.db.Create(notification).Error
}

// GetDBFromContext 获取 gin.Context 中的 db
func GetDBFromContext(ctx *gin.Context) *gorm.DB {
	return ctx.MustGet("db").(*gorm.DB)
}

// itoa 数字转字符串
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

// NotifyWeeklyReportRejected 通知周报被退回
func (h *NotificationHelper) NotifyWeeklyReportRejected(playerID uint, coachName, comment string, reportID uint) error {
	title := "周报被退回"
	content := coachName + " 退回了你的周报，请修改后重新提交"
	link := "/weekly-reports/" + itoa(int(reportID)) + "/edit"
	return h.CreateNotification(playerID, models.NotificationTypeWeeklyReportRejected, title, content, link)
}

// NotifyWeeklyReportApproved 通知周报审核完成
func (h *NotificationHelper) NotifyWeeklyReportApproved(playerID uint, coachName string, rating int, reportID uint) error {
	title := "周报已审核"
	content := coachName + " 已审核你的周报，评分 " + itoa(rating) + " 星"
	link := "/weekly-reports/" + itoa(int(reportID))
	return h.CreateNotification(playerID, models.NotificationTypeWeeklyReportApproved, title, content, link)
}

// NotifyWeeklyReportCreated 通知球员教练发起了周报
func (h *NotificationHelper) NotifyWeeklyReportCreated(playerID uint, coachName, teamName, weekLabel string, reportID uint) error {
	title := "教练发起了本周周报"
	content := coachName + " 发起了一篇新周报，请填写"
	link := "/weekly-reports/" + itoa(int(reportID)) + "/edit"
	return h.CreateNotification(playerID, models.NotificationTypeWeeklyReportCreated, title, content, link)
}

// NotifyMatchSummaryCreated 通知球员比赛已创建
func (h *NotificationHelper) NotifyMatchSummaryCreated(playerID uint, coachName, matchName string, matchID uint) error {
	title := "新比赛已创建"
	content := coachName + " 创建了比赛 " + matchName + "，请做好准备"
	link := "/match-reports/" + itoa(int(matchID))
	return h.CreateNotification(playerID, models.NotificationTypeMatchSummaryCreated, title, content, link)
}

// NotifyMatchCoachReminder 通知教练填写点评
func (h *NotificationHelper) NotifyMatchCoachReminder(coachID uint, matchName string, matchID uint) error {
	title := "比赛已结束，请填写点评"
	content := matchName + " 已结束，请填写你的比赛点评"
	link := "/match-reports/" + itoa(int(matchID)) + "/coach-summary"
	return h.CreateNotification(coachID, models.NotificationTypeMatchCoachReminder, title, content, link)
}

// NotifyMatchPlayerReminder 通知球员提交比赛自评
func (h *NotificationHelper) NotifyMatchPlayerReminder(playerID uint, matchName string, matchID uint) error {
	title := "请提交比赛自评"
	content := matchName + " 的比赛总结需要你提交自评，请及时完成"
	link := "/match-reports/" + itoa(int(matchID)) + "/self-review"
	return h.CreateNotification(playerID, models.NotificationTypeMatchPlayerReminder, title, content, link)
}

// NotifyMatchSummaryComplete 通知比赛总结完成
func (h *NotificationHelper) NotifyMatchSummaryComplete(playerID uint, coachName, matchName string, matchID uint) error {
	title := "比赛总结已完成"
	content := coachName + " 已完成 " + matchName + " 的点评"
	link := "/match-reports/" + itoa(int(matchID))
	return h.CreateNotification(playerID, models.NotificationTypeMatchSummaryComplete, title, content, link)
}
