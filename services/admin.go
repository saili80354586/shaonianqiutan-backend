package services

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/shaonianqiutan/backend/models"
	"github.com/shaonianqiutan/backend/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AdminService 管理后台服务
type AdminService struct {
	userRepo            *models.UserRepository
	reportRepo          *models.ReportRepository
	orderRepo           *models.OrderRepository
	analystRepo         *models.AnalystRepository
	applicationRepo     *models.AnalystApplicationRepository
	contentReportRepo   *models.ContentReportRepository
	sensitiveWordRepo   *models.SensitiveWordRepository
	platformAnnRepo     *models.PlatformAnnouncementRepository
	bannerRepo          *models.BannerRepository
	faqRepo             *models.FAQRepository
	loginLogRepo        *models.LoginLogRepository
	videoAnalysisRepo   *models.VideoAnalysisRepository
	assignmentRepo      *models.OrderAssignmentRepository
	statusHistoryRepo   *models.OrderStatusHistoryRepository
	notificationService *NotificationService
}

// NewAdminService 创建管理后台服务
func NewAdminService(
	userRepo *models.UserRepository,
	reportRepo *models.ReportRepository,
	orderRepo *models.OrderRepository,
	analystRepo *models.AnalystRepository,
	applicationRepo *models.AnalystApplicationRepository,
	contentReportRepo *models.ContentReportRepository,
	sensitiveWordRepo *models.SensitiveWordRepository,
	platformAnnRepo *models.PlatformAnnouncementRepository,
	bannerRepo *models.BannerRepository,
	faqRepo *models.FAQRepository,
	loginLogRepo *models.LoginLogRepository,
	videoAnalysisRepo *models.VideoAnalysisRepository,
	assignmentRepo *models.OrderAssignmentRepository,
	statusHistoryRepo *models.OrderStatusHistoryRepository,
) *AdminService {
	return &AdminService{
		userRepo:          userRepo,
		reportRepo:        reportRepo,
		orderRepo:         orderRepo,
		analystRepo:       analystRepo,
		applicationRepo:   applicationRepo,
		contentReportRepo: contentReportRepo,
		sensitiveWordRepo: sensitiveWordRepo,
		platformAnnRepo:   platformAnnRepo,
		bannerRepo:        bannerRepo,
		faqRepo:           faqRepo,
		loginLogRepo:      loginLogRepo,
		videoAnalysisRepo: videoAnalysisRepo,
		assignmentRepo:    assignmentRepo,
		statusHistoryRepo: statusHistoryRepo,
	}
}

// SetNotificationService 注入通知服务
func (s *AdminService) SetNotificationService(notificationService *NotificationService) {
	s.notificationService = notificationService
}

// ========== Dashboard Stats ==========

// DashboardStats 数据看板统计数据
type DashboardStats struct {
	TotalUsers            int64   `json:"total_users"`
	TotalOrders           int64   `json:"total_orders"`
	TotalReports          int64   `json:"total_reports"`
	TotalRevenue          float64 `json:"total_revenue"`
	TodayNewUsers         int64   `json:"today_new_users"`
	TodayOrders           int64   `json:"today_orders"`
	TodayRevenue          float64 `json:"today_revenue"`
	PendingApplications   int64   `json:"pending_applications"`
	PendingReports        int64   `json:"pending_reports"`
	PendingContentReports int64   `json:"pending_content_reports"`
}

// GetDashboardStats 获取数据看板统计数据
func (s *AdminService) GetDashboardStats() (*DashboardStats, error) {
	stats := &DashboardStats{}

	totalUsers, err := s.userRepo.Count()
	if err != nil {
		return nil, err
	}
	stats.TotalUsers = totalUsers

	today := time.Now().Format("2006-01-02")
	todayUsers, err := s.userRepo.CountByDate(today)
	if err != nil {
		return nil, err
	}
	stats.TodayNewUsers = todayUsers

	orderStats, err := s.orderRepo.GetStatistics()
	if err != nil {
		return nil, err
	}
	stats.TotalOrders = orderStats.TotalCount
	stats.TotalRevenue = orderStats.TotalRevenue
	stats.TodayOrders = orderStats.TodayCount
	stats.TodayRevenue = orderStats.TodayRevenue

	reportStats, err := s.reportRepo.GetStatistics()
	if err != nil {
		return nil, err
	}
	stats.TotalReports = reportStats.TotalCount
	stats.PendingReports = reportStats.PendingCount

	if s.applicationRepo != nil {
		pendingApps, err := s.applicationRepo.CountByStatus(models.ApplicationStatusPending)
		if err == nil {
			stats.PendingApplications = pendingApps
		}
	}

	if s.contentReportRepo != nil {
		pendingCR, _ := s.contentReportRepo.CountByStatus(models.ContentReportStatusPending)
		stats.PendingContentReports = pendingCR
	}

	return stats, nil
}

// GrowthData 增长数据
type GrowthData struct {
	Date    string  `json:"date"`
	Users   int64   `json:"users"`
	Orders  int64   `json:"orders"`
	Revenue float64 `json:"revenue"`
}

// GetGrowthData 获取增长数据
func (s *AdminService) GetGrowthData(days int) ([]GrowthData, error) {
	var result []GrowthData

	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")

		userCount, _ := s.userRepo.CountByDate(date)
		orderCount, orderRevenue, _ := s.orderRepo.GetStatisticsByDate(date)

		result = append(result, GrowthData{
			Date:    date,
			Users:   userCount,
			Orders:  orderCount,
			Revenue: orderRevenue,
		})
	}

	return result, nil
}

// GetFunnelData 获取漏斗数据
func (s *AdminService) GetFunnelData() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	var totalVisitors int64 = 0
	result["visitors"] = totalVisitors

	var registrations int64
	s.userRepo.CountByRole("", &registrations)
	result["registrations"] = registrations

	var orders int64
	s.orderRepo.GetTotalCount(&orders)
	result["orders"] = orders

	var payments int64
	s.orderRepo.GetPaidCount(&payments)
	result["payments"] = payments

	var completed int64
	s.orderRepo.GetCompletedCount(&completed)
	result["completed"] = completed

	return result, nil
}

// GetRetentionData 获取留存数据
func (s *AdminService) GetRetentionData(days int) ([]map[string]interface{}, error) {
	var result []map[string]interface{}

	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")

		newUsers, _ := s.userRepo.CountByDate(date)
		activeUsers, _ := s.userRepo.CountActiveByDate(date)

		result = append(result, map[string]interface{}{
			"date":         date,
			"new_users":    newUsers,
			"active_users": activeUsers,
			"retention":    0,
		})
	}

	return result, nil
}

// GetTopData 获取排行榜数据
func (s *AdminService) GetTopData() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	topPlayers, _ := s.userRepo.GetTopByRole("player", 10)
	result["top_players"] = topPlayers

	topAnalysts, _ := s.analystRepo.GetTopByOrders(10)
	result["top_analysts"] = topAnalysts

	topClubs, _ := s.userRepo.GetTopByRole("club", 10)
	result["top_clubs"] = topClubs

	return result, nil
}

// ========== User Management ==========

// GetUserList 获取用户列表
func (s *AdminService) GetUserList(page, pageSize int) ([]models.User, int64, error) {
	return s.userRepo.FindAll(page, pageSize)
}

// UpdateUserStatus 更新用户状态
func (s *AdminService) UpdateUserStatus(userID uint, status string) error {
	return s.userRepo.UpdateStatus(userID, status)
}

// DeleteUser 删除用户
func (s *AdminService) DeleteUser(userID uint) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	if user.Role == "admin" {
		return errors.New("不能删除管理员账号")
	}
	return s.userRepo.Delete(userID)
}

// ========== Report Management ==========

// GetPendingReports 获取待审核报告列表
func (s *AdminService) GetPendingReports(page, pageSize int) ([]models.Report, int64, error) {
	reports, total, err := s.reportRepo.FindByStatus(models.ReportStatusProcessing, page, pageSize)
	if err != nil {
		log.Printf("[AdminService] get pending reports failed: %v", err)
	}
	return reports, total, err
}

// ReviewReport 审核报告
func (s *AdminService) ReviewReport(reportID uint, status models.ReportStatus, remark string, adminID uint) error {
	report, err := s.reportRepo.FindByID(reportID)
	if err != nil {
		return err
	}
	if report == nil {
		return errors.New("报告不存在")
	}

	order, err := s.orderRepo.FindByID(report.OrderID)
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("订单不存在")
	}

	now := time.Now()
	err = s.orderRepo.GetDB().Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"status": status,
		}
		if remark != "" || status == models.ReportStatusFailed {
			updates["review_remark"] = remark
		}
		if err := tx.Model(&models.Report{}).Where("id = ?", reportID).Updates(updates).Error; err != nil {
			return err
		}

		switch status {
		case models.ReportStatusCompleted:
			if err := tx.Model(&models.Order{}).Where("id = ?", report.OrderID).Updates(map[string]interface{}{
				"status":       models.OrderStatusCompleted,
				"report_id":    report.ID,
				"completed_at": &now,
			}).Error; err != nil {
				return err
			}
			if s.videoAnalysisRepo != nil {
				if err := tx.Model(&models.VideoAnalysis{}).Where("order_id = ?", report.OrderID).Updates(map[string]interface{}{
					"status":           models.AnalysisStatusCompleted,
					"ai_report_status": "confirmed",
				}).Error; err != nil {
					return err
				}
			}
			if err := s.createStatusHistory(tx, report.OrderID, order.Status, models.OrderStatusCompleted, adminID, "admin", "管理员审核通过报告"); err != nil {
				return err
			}
		case models.ReportStatusFailed:
			if s.videoAnalysisRepo != nil {
				if err := tx.Model(&models.VideoAnalysis{}).Where("order_id = ?", report.OrderID).Updates(map[string]interface{}{
					"status":           models.AnalysisStatusDraft,
					"ai_report_status": "draft",
				}).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	if status == models.ReportStatusCompleted {
		s.notifyReportCompleted(report)
	} else if status == models.ReportStatusFailed {
		s.notifyAnalystReportRejected(report, remark)
	}
	return nil
}

// GetReportByID 获取报告详情
func (s *AdminService) GetReportByID(reportID uint) (*models.Report, error) {
	return s.reportRepo.FindByID(reportID)
}

// UpdateReportAIURL 更新报告的 AI 报告/视频 URL
func (s *AdminService) UpdateReportAIURL(reportID uint, reportURL, videoURL string) error {
	updates := map[string]interface{}{}
	if reportURL != "" {
		updates["ai_report_url"] = reportURL
	}
	if videoURL != "" {
		updates["ai_video_url"] = videoURL
	}
	if len(updates) == 0 {
		return nil
	}
	return s.reportRepo.Update(reportID, updates)
}

// ========== Order Management ==========

// GetAllOrders 获取所有订单
func (s *AdminService) GetAllOrders(page, pageSize int, status string) ([]models.Order, int64, error) {
	return s.orderRepo.FindAll(page, pageSize, status)
}

// CancelOrder 取消订单
func (s *AdminService) CancelOrder(orderID, adminID uint) error {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("订单不存在")
	}

	allowedStatuses := []models.OrderStatus{
		models.OrderStatusPending,
		models.OrderStatusPaid,
		models.OrderStatusUploaded,
		models.OrderStatusAssigned,
	}
	canCancel := false
	for _, status := range allowedStatuses {
		if order.Status == status {
			canCancel = true
			break
		}
	}
	if !canCancel {
		return errors.New("该订单状态不允许取消")
	}

	return s.orderRepo.GetDB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Order{}).Where("id = ?", orderID).Updates(map[string]interface{}{
			"status": models.OrderStatusCancelled,
		}).Error; err != nil {
			return err
		}
		if s.assignmentRepo != nil && order.AnalystID != nil {
			now := time.Now()
			if err := s.assignmentRepo.MarkLatestPendingWithTx(tx, orderID, *order.AnalystID, models.OrderAssignmentStatusExpired, "管理员取消订单", now); err != nil {
				return err
			}
		}
		return s.createStatusHistory(tx, orderID, order.Status, models.OrderStatusCancelled, adminID, "admin", "管理员取消订单")
	})
}

// AssignOrder 管理员派单给分析师
func (s *AdminService) AssignOrder(orderID, analystID, adminID uint) (*models.Order, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, errors.New("订单不存在")
	}
	if order.Status != models.OrderStatusUploaded {
		return nil, errors.New("订单状态不允许派单")
	}

	analyst, err := s.analystRepo.FindByID(analystID)
	if err != nil {
		return nil, err
	}
	if analyst == nil {
		return nil, errors.New("分析师不存在")
	}
	if analyst.Status != models.AnalystStatusActive {
		return nil, errors.New("分析师暂不可用")
	}

	deadlineHours := 48
	if order.OrderType == "pro" {
		deadlineHours = 72
	}
	assignedAt := time.Now()
	deadline := assignedAt.Add(time.Duration(deadlineHours) * time.Hour)

	updates := map[string]interface{}{
		"analyst_id":  analystID,
		"status":      models.OrderStatusAssigned,
		"assigned_at": assignedAt,
		"deadline":    deadline,
	}

	if err := s.orderRepo.GetDB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Order{}).Where("id = ?", orderID).Updates(updates).Error; err != nil {
			return err
		}
		if s.assignmentRepo != nil {
			assignment := &models.OrderAssignment{
				OrderID:    orderID,
				AnalystID:  analystID,
				AssignedBy: optionalActorID(adminID),
				AssignedAt: assignedAt,
				Status:     models.OrderAssignmentStatusPending,
			}
			if err := s.assignmentRepo.CreateWithTx(tx, assignment); err != nil {
				return err
			}
		}
		return s.createStatusHistory(tx, orderID, order.Status, models.OrderStatusAssigned, adminID, "admin", "管理员派发订单")
	}); err != nil {
		return nil, err
	}

	assignedOrder, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return nil, err
	}
	s.notifyAnalystOrderAssigned(analyst, assignedOrder)
	return assignedOrder, nil
}

// GetAssignmentRecords 获取订单派发历史
func (s *AdminService) GetAssignmentRecords(page, pageSize int, status string) ([]models.OrderAssignment, int64, error) {
	if status != "" && !models.IsValidOrderAssignmentStatus(status) {
		return nil, 0, errors.New("无效的派发状态")
	}
	return s.assignmentRepo.FindAll(page, pageSize, status)
}

// GetOrderStatusHistory 获取订单状态流转历史
func (s *AdminService) GetOrderStatusHistory(orderID uint) ([]models.OrderStatusHistory, error) {
	return s.statusHistoryRepo.FindByOrderID(orderID)
}

func (s *AdminService) createStatusHistory(tx *gorm.DB, orderID uint, fromStatus, toStatus models.OrderStatus, actorID uint, actorRole, reason string) error {
	if s.statusHistoryRepo == nil || fromStatus == toStatus {
		return nil
	}
	return s.statusHistoryRepo.CreateWithTx(tx, &models.OrderStatusHistory{
		OrderID:    orderID,
		FromStatus: fromStatus,
		ToStatus:   toStatus,
		ActorID:    optionalActorID(actorID),
		ActorRole:  actorRole,
		Reason:     reason,
	})
}

func (s *AdminService) notifyAnalystOrderAssigned(analyst *models.Analyst, order *models.Order) {
	if s.notificationService == nil || analyst == nil || order == nil {
		return
	}
	if err := s.notificationService.NotifyAnalystOrderAssigned(analyst.UserID, order.ID, order.OrderNo, order.PlayerName); err != nil {
		log.Printf("[AdminService] notify analyst %d for order %d failed: %v", analyst.ID, order.ID, err)
	}
}

func (s *AdminService) notifyReportCompleted(report *models.Report) {
	if s.notificationService == nil || report == nil {
		return
	}
	if err := s.notificationService.NotifyReportCompleted(report.UserID, report.ID, report.PlayerName); err != nil {
		log.Printf("[AdminService] notify player %d for report %d failed: %v", report.UserID, report.ID, err)
	}
}

func (s *AdminService) notifyAnalystReportRejected(report *models.Report, remark string) {
	if s.notificationService == nil || s.analystRepo == nil || report == nil {
		return
	}
	analyst, err := s.analystRepo.FindByID(report.AnalystID)
	if err != nil || analyst == nil {
		if err != nil {
			log.Printf("[AdminService] query analyst %d for report notification failed: %v", report.AnalystID, err)
		}
		return
	}
	if err := s.notificationService.NotifyAnalystReportRejected(analyst.UserID, report.ID, report.PlayerName, remark); err != nil {
		log.Printf("[AdminService] notify analyst %d for rejected report %d failed: %v", analyst.ID, report.ID, err)
	}
}

func optionalActorID(actorID uint) *uint {
	if actorID == 0 {
		return nil
	}
	return &actorID
}

// GetOrderStats 获取订单统计
func (s *AdminService) GetOrderStats() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	stats, _ := s.orderRepo.GetStatistics()
	result["total_count"] = stats.TotalCount
	result["total_revenue"] = stats.TotalRevenue
	result["today_count"] = stats.TodayCount
	result["today_revenue"] = stats.TodayRevenue

	var statusCounts []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	s.orderRepo.GetStatusCounts(&statusCounts)
	result["status_counts"] = statusCounts

	return result, nil
}

// GetRevenueTrend 获取收入趋势
func (s *AdminService) GetRevenueTrend(days int) ([]map[string]interface{}, error) {
	var result []map[string]interface{}

	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		count, revenue, _ := s.orderRepo.GetStatisticsByDate(date)

		result = append(result, map[string]interface{}{
			"date":    date,
			"orders":  count,
			"revenue": revenue,
		})
	}

	return result, nil
}

// ========== Analyst Management ==========

// GetAnalystList 获取分析师列表
func (s *AdminService) GetAnalystList(page, pageSize int, status string) ([]models.User, int64, error) {
	return s.userRepo.FindByRole("analyst", page, pageSize, status)
}

// AuditAnalyst 审核分析师
func (s *AdminService) AuditAnalyst(analystID uint, status string, remark string) error {
	return nil
}

// UpdateAnalystStatus 更新分析师状态
func (s *AdminService) UpdateAnalystStatus(analystID uint, status string) error {
	return s.userRepo.UpdateStatus(analystID, status)
}

// GetAnalystIncomeStats 获取分析师收益统计
func (s *AdminService) GetAnalystIncomeStats(analystID uint) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	totalIncome, _ := s.orderRepo.GetAnalystTotalIncome(analystID)
	result["total_income"] = totalIncome

	monthIncome, _ := s.orderRepo.GetAnalystMonthIncome(analystID)
	result["month_income"] = monthIncome

	orderCount, _ := s.orderRepo.GetAnalystOrderCount(analystID)
	result["order_count"] = orderCount

	return result, nil
}

// GetSettlementList 获取待结算列表
func (s *AdminService) GetSettlementList(page, pageSize int) ([]models.Order, int64, error) {
	return s.orderRepo.FindCompletedUnsettled(page, pageSize)
}

// ProcessSettlement 处理结算
func (s *AdminService) ProcessSettlement(orderIDs []uint, adminID uint) error {
	for _, id := range orderIDs {
		if err := s.orderRepo.UpdateSettlement(id, adminID, time.Now()); err != nil {
			return err
		}
	}
	return nil
}

// ========== Content Report ==========

// GetContentReports 获取举报列表
func (s *AdminService) GetContentReports(page, pageSize int, status string) ([]models.ContentReport, int64, error) {
	return s.contentReportRepo.FindAll(page, pageSize, status)
}

// HandleContentReport 处理举报
func (s *AdminService) HandleContentReport(reportID uint, status models.ContentReportStatus, handlerID uint, handlerName, result string) error {
	return s.contentReportRepo.UpdateStatus(reportID, status, handlerID, handlerName, result)
}

// GetContentReportDetail 获取举报详情
func (s *AdminService) GetContentReportDetail(reportID uint) (*models.ContentReport, error) {
	return s.contentReportRepo.FindByID(reportID)
}

// ========== Sensitive Word ==========

// GetSensitiveWords 获取敏感词列表
func (s *AdminService) GetSensitiveWords(page, pageSize int, category string, enabled *bool) ([]models.SensitiveWord, int64, error) {
	return s.sensitiveWordRepo.FindAll(page, pageSize, category, enabled)
}

// CreateSensitiveWord 创建敏感词
func (s *AdminService) CreateSensitiveWord(word *models.SensitiveWord) error {
	return s.sensitiveWordRepo.Create(word)
}

// UpdateSensitiveWord 更新敏感词
func (s *AdminService) UpdateSensitiveWord(id uint, updates map[string]interface{}) error {
	return s.sensitiveWordRepo.Update(id, updates)
}

// DeleteSensitiveWord 删除敏感词
func (s *AdminService) DeleteSensitiveWord(id uint) error {
	return s.sensitiveWordRepo.Delete(id)
}

// CheckSensitiveWords 检查敏感词
func (s *AdminService) CheckSensitiveWords(text string) ([]string, error) {
	return s.sensitiveWordRepo.CheckText(text)
}

// ========== Platform Announcement ==========

// GetPlatformAnnouncements 获取平台公告列表
func (s *AdminService) GetPlatformAnnouncements(page, pageSize int, annType string, pinned *bool) ([]models.PlatformAnnouncement, int64, error) {
	return s.platformAnnRepo.FindAll(page, pageSize, annType, pinned)
}

// CreatePlatformAnnouncement 创建平台公告
func (s *AdminService) CreatePlatformAnnouncement(ann *models.PlatformAnnouncement) error {
	return s.platformAnnRepo.Create(ann)
}

// UpdatePlatformAnnouncement 更新平台公告
func (s *AdminService) UpdatePlatformAnnouncement(id uint, updates map[string]interface{}) error {
	return s.platformAnnRepo.Update(id, updates)
}

// DeletePlatformAnnouncement 删除平台公告
func (s *AdminService) DeletePlatformAnnouncement(id uint) error {
	return s.platformAnnRepo.Delete(id)
}

// ========== Banner ==========

// GetBanners 获取轮播图列表
func (s *AdminService) GetBanners(page, pageSize int, position string, enabled *bool) ([]models.Banner, int64, error) {
	return s.bannerRepo.FindAll(page, pageSize, position, enabled)
}

// CreateBanner 创建轮播图
func (s *AdminService) CreateBanner(banner *models.Banner) error {
	return s.bannerRepo.Create(banner)
}

// UpdateBanner 更新轮播图
func (s *AdminService) UpdateBanner(id uint, updates map[string]interface{}) error {
	return s.bannerRepo.Update(id, updates)
}

// DeleteBanner 删除轮播图
func (s *AdminService) DeleteBanner(id uint) error {
	return s.bannerRepo.Delete(id)
}

// ========== FAQ ==========

// GetFAQs 获取FAQ列表
func (s *AdminService) GetFAQs(page, pageSize int, category string, enabled *bool) ([]models.FAQ, int64, error) {
	return s.faqRepo.FindAll(page, pageSize, category, enabled)
}

// CreateFAQ 创建FAQ
func (s *AdminService) CreateFAQ(faq *models.FAQ) error {
	return s.faqRepo.Create(faq)
}

// UpdateFAQ 更新FAQ
func (s *AdminService) UpdateFAQ(id uint, updates map[string]interface{}) error {
	return s.faqRepo.Update(id, updates)
}

// DeleteFAQ 删除FAQ
func (s *AdminService) DeleteFAQ(id uint) error {
	return s.faqRepo.Delete(id)
}

// ========== Login Log ==========

// GetLoginLogs 获取登录日志
func (s *AdminService) GetLoginLogs(page, pageSize int, userID uint, status, startDate, endDate string) ([]models.LoginLog, int64, error) {
	return s.loginLogRepo.FindAll(page, pageSize, userID, status, startDate, endDate)
}

// GetLoginLogStats 获取登录日志统计
func (s *AdminService) GetLoginLogStats(days int) (map[string]interface{}, error) {
	return s.loginLogRepo.GetStatistics(days)
}

// CreateLoginLog 创建登录日志
func (s *AdminService) CreateLoginLog(log *models.LoginLog) error {
	return s.loginLogRepo.Create(log)
}

// ========== Admin Login ==========

// ========== Available Analysts for Dispatch ==========

// AvailableAnalyst 可派单的分析师（含工作负载）
type AvailableAnalyst struct {
	AnalystID       uint        `json:"analyst_id"`
	Analyst         models.User `json:"analyst"`
	MaxOrders       int         `json:"max_orders"`
	AcceptedOrders  int         `json:"accepted_orders"`
	CompletedOrders int         `json:"completed_orders"`
	WorkingHours    string      `json:"working_hours"`
	IsAvailable     bool        `json:"is_available"`
	TotalCompleted  int64       `json:"total_completed"`
	AvgRating       float64     `json:"avg_rating"`
	Specialties     []string    `json:"specialties"`
}

// GetAvailableAnalysts 获取可派单的分析师列表（含实时工作负载）
func (s *AdminService) GetAvailableAnalysts() ([]AvailableAnalyst, error) {
	// 查询所有活跃分析师（已预加载 User）
	analysts, _, err := s.analystRepo.FindAll(1, 100)
	if err != nil {
		return nil, err
	}

	var result []AvailableAnalyst
	for _, analyst := range analysts {
		// 统计当前进行中订单数（assigned + processing）
		var activeCount int64
		s.orderRepo.GetDB().Model(&models.Order{}).
			Where("analyst_id = ? AND status IN ?", analyst.ID, []models.OrderStatus{models.OrderStatusAssigned, models.OrderStatusProcessing}).
			Count(&activeCount)

		// 统计历史完成订单数
		var totalCompleted int64
		s.orderRepo.GetDB().Model(&models.Order{}).
			Where("analyst_id = ? AND status = ?", analyst.ID, models.OrderStatusCompleted).
			Count(&totalCompleted)

		maxOrders := 5 // 默认每日最大接单量
		specialties := []string{}
		if analyst.Specialty != "" {
			specialties = strings.Split(analyst.Specialty, ",")
		}

		result = append(result, AvailableAnalyst{
			AnalystID:       analyst.ID,
			Analyst:         analyst.User,
			MaxOrders:       maxOrders,
			AcceptedOrders:  int(activeCount),
			CompletedOrders: int(totalCompleted),
			WorkingHours:    "09:00-18:00",
			IsAvailable:     int(activeCount) < maxOrders,
			TotalCompleted:  totalCompleted,
			AvgRating:       analyst.Rating,
			Specialties:     specialties,
		})
	}

	return result, nil
}

// AdminLogin 管理员登录
func (s *AdminService) AdminLogin(username, password string) (string, *models.User, error) {
	admin, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return "", nil, errors.New("用户名或密码错误")
	}

	if admin.Role != "admin" {
		return "", nil, errors.New("无权限访问")
	}
	if admin.Status != models.StatusActive {
		return "", nil, errors.New("账号未激活或已被禁用")
	}

	err = bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password))
	if err != nil {
		return "", nil, errors.New("用户名或密码错误")
	}

	token, err := utils.GenerateToken(admin.ID, admin.Phone, string(admin.Role))
	if err != nil {
		return "", nil, err
	}

	admin.Password = ""

	return token, admin, nil
}
